package cmd

import (
	"autocli/model"
	"autocli/service"
	"errors"
	"fmt"

	"net"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
)

func NewWatchCommand(b Builder) *cobra.Command {
	var watchCmd = &cobra.Command{
		Use:          "watch [flags] [contexts]...",
		Short:        "Start watching Kube servers",
		Long:         "",
		SilenceUsage: true,
		Args:         cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := RunCommon(cmd); err != nil {
				log.Error(err)
				return err
			}
			return RunWatch(b, cmd, args)
		},
	}

	AddCommonFlags(watchCmd)
	watchCmd.Flags().Duration("interval", 2*time.Minute, "Interval between requests to the server")
	watchCmd.Flags().String("only", "", "Coma-separated names of resources to watch, empty to watch all supported")

	return watchCmd
}

func RunWatch(b Builder, cmd *cobra.Command, args []string) error {
	clients := make(map[string]kubernetes.Interface)

	kubeConfigFile, err := cmd.Flags().GetString("kubeconfig")
	if err != nil {
		log.Error(err)
		return err
	}

	bind, err := GetBind(cmd)
	if err != nil {
		msg := fmt.Sprintf("failed to generate watch server bind address: %s", err)
		log.Error(msg)
		return errors.New(msg)
	}

	l, err := net.Listen("tcp", bind)
	if err != nil {
		msg := fmt.Sprintf("failed to bind on %s: %v", bind, err)
		log.Error(msg)
		return errors.New(msg)
	}

	interval, err := cmd.Flags().GetDuration("interval")
	if err != nil {
		msg := fmt.Sprintf("could not parse value of --interval")
		log.Error(msg)
		return errors.New(msg)
	}

	enabledResources, err := cmd.Flags().GetString("only")
	if err != nil {
		msg := fmt.Sprintf("could not parse value of --only")
		log.Error(msg)
		return errors.New(msg)
	}

	c := b.WatchCache()

	for _, arg := range args {
		cc, err := BuildConfigFromFlags(arg, kubeConfigFile)
		if err != nil {
			log.Error(err)
			return err
		}

		clientset, err := kubernetes.NewForConfig(cc)
		if err != nil {
			log.Error(err)
			return err
		}
		clients[arg] = clientset
		fields := log.Fields{
			"context": arg,
			"host":    cc.Host,
		}
		log.WithFields(fields).Info("created client")
	}

	kc := b.KubeClient(clients)

	for _, ctx := range args {
		if err := kc.Ping(ctx); err != nil {
			msg := fmt.Sprintf("failed to ping server: %s", err)
			log.Error(msg)
			return errors.New(msg)
		}
	}

	for _, ctx := range args {
		for _, watchResource := range []string{"pod"} {
			if isWatching(watchResource, enabledResources) {
				loopWatchObjects(c, kc, watchResource, ctx)
			}
		}

		for _, getResource := range []string{"node"} {
			if isWatching(getResource, enabledResources) {
				loopGetObjects(c, kc, getResource, ctx, interval)
			}
		}
	}

	log.WithField("bind", bind).Info("started to listen")
	err = b.Serve(l, c)
	if err != nil {
		return fmt.Errorf("unexpected error: %v", err)
	}

	return errors.New("watch has stopped")
}

func isWatching(r string, rs string) bool {
	return len(rs) == 0 || strings.Contains(rs, r)
}

func loopWatchObjects(c *WatchCache, kc service.KubeClient, kind, context string) {
	events := make(chan *model.ResourceEvent)
	l := log.WithField("kind", kind).WithField("context", context)

	watch := func() {
		for {
			l.Info("started to watch")
			err := kc.WatchResources(context, kind, events)
			fields := log.Fields{}
			if err != nil {
				fields["error"] = err.Error()
			}
			l.WithFields(fields).Info("watch connection was closed, retrying")
			c.deleteKubeObjects(context, kind)
		}
	}

	update := func() {
		for {
			select {
			case e := <-events:
				l.
					WithField("name", e.Resource.Name).
					WithField("type", e.Type).
					Info("received event")
				switch e.Type {
				case model.Deleted:
					c.deleteKubeObject(context, *e.Resource)
				case model.Added, model.Modified:
					c.updateKubeObject(context, *e.Resource)
				}
				l.WithField("cache", c.Resources).Debugf("objects in cache")
			}
		}
	}

	go watch()
	go update()
}

func loopGetObjects(c *WatchCache, kc service.KubeClient, kind, context string, interval time.Duration) {
	l := log.WithField("kind", kind).WithField("context", context)
	update := func() {
		for {
			l.Info("updating resource...")
			resources, err := kc.GetResources(context, kind)
			if err != nil {
				l.WithField("error", err).Error("unexpected error while updating resources")
				time.Sleep(10 * time.Second)
				continue
			}

			l.WithField("resources", resources).Debug("received resources")
			c.deleteKubeObjects(context, kind)
			for i := range resources {
				c.updateKubeObject(context, resources[i])
			}
			l.Infof("put %d resources into cache", len(resources))

			time.Sleep(interval)
		}
	}

	go update()
}
