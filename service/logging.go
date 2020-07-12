package service

import (
	syslog "log/syslog"
	"os"

	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.WarnLevel)
	log.SetFormatter(&log.TextFormatter{})
}

func EnableVerbose() {
	log.SetLevel(log.InfoLevel)
}

func EnableVeryVerbose() {
	log.SetLevel(log.DebugLevel)
}

func EnableSysLog() {
	hook, err := NewSyslogHook("", "", syslog.LOG_INFO, "")
	if err == nil {
		log.AddHook(hook)
		log.Debug("added hook")
		//reformat the log entries to better work with syslog
		log.SetFormatter(&log.TextFormatter{
			DisableColors:    true,
			DisableTimestamp: true,
		})
	} else {
		log.Errorf("Failed to add SysLog hook: %s", err)
	}
}
