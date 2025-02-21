package main

import (
	"os"
	"task-exec-agent/executor"

	log "github.com/sirupsen/logrus"
)

func main() {
	setLogConfigFromEnv()
	executor := executor.New(executor.NewConfig())
	executor.Run()
}

func setLogConfigFromEnv() {
	level, err := log.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		level = log.DebugLevel
	}
	log.SetLevel(level)
}
