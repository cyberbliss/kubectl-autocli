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
		return err
	}

	bind, err := GetBind(cmd)
	if err != nil {
		return fmt.Errorf("unexpected error: %s", err)
	}

	l, err := net.Listen("tcp", bind)
	if err != nil {
		return fmt.Errorf("failed to bind on %s: %v", bind, err)
	}

	//interval, err := cmd.Flags().GetDuration("interval")
	//if err != nil {
	//	return errors.New("could not parse value of --interval")
	//}

	enabledResources, err := cmd.Flags().GetString("only")
	if err != nil {
		return errors.New("could not parse value of --only")
	}

	c := b.WatchCache()

	for _, arg := range args {
		cc, err := BuildConfigFromFlags(arg, kubeConfigFile)
		if err != nil {
			return err
		}
		clientset, err := kubernetes.NewForConfig(cc)
		if err != nil {
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

	for _, arg := range args {
		kc.Ping(arg)
	}

	for _, arg := range args {
		for _, watchResource := range []string{"pod"} {
			if isWatching(watchResource, enabledResources) {
				loopWatchObjects(c, kc, watchResource, arg)
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
