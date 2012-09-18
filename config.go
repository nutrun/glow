package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

const (
	DEFAULT_QUEUE_ADDR = "0.0.0.0:11300"
	DEFAULT_FROM_EMAIL = "glow@example.com"
)

type Configuration struct {
	QueueAddr  string
	SmtpServer string
	MailFrom   string
	depsPath   string
	deps       map[string][]string
	errorQueue string
}

func NewConfig(deps, smtpserver, mailfrom string) *Configuration {
	config := new(Configuration)
	config.QueueAddr = os.Getenv("GLOW_QUEUE")
	if config.QueueAddr == "" {
		config.QueueAddr = DEFAULT_QUEUE_ADDR
	}
	config.SmtpServer = smtpserver
	config.MailFrom = mailfrom
	config.depsPath = deps
	if config.MailFrom == "" {
		config.MailFrom = DEFAULT_FROM_EMAIL
	}
	config.errorQueue = "GLOW_ERRORS"
	return config
}

func (this *Configuration) LoadDeps() error {
	if *deps == "" {
		return nil
	}
	this.deps = make(map[string][]string)
	dependencies, err := ioutil.ReadFile(this.depsPath)
	if err != nil {
		return err
	}
	return json.Unmarshal(dependencies, &this.deps)
}
