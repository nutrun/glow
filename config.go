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
	deps       map[string][]string
	errorQueue string
}

func NewConfig() *Configuration {
	config := new(Configuration)
	config.QueueAddr = os.Getenv("GLOW_QUEUE")
	if config.QueueAddr == "" {
		config.QueueAddr = DEFAULT_QUEUE_ADDR
	}
	config.SmtpServer = *smtpserver
	config.MailFrom = *mailfrom
	if config.MailFrom == "" {
		config.MailFrom = DEFAULT_FROM_EMAIL
	}
	config.errorQueue = "GLOW_ERRORS"
	return config
}

func (this *Configuration) Load() error {
	path := *deps
	this.deps = make(map[string][]string)
	deps, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(deps, &this.deps)
}

var Config = NewConfig()
