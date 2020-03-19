package service

import (
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
