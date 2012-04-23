package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type Configuration struct {
	QueueAddr      string
	SmtpServerAddr string
	MailFrom       string
	deps           map[string][]string
}

func NewConfig() *Configuration {
	config := new(Configuration)
	config.QueueAddr = os.Getenv("GLOW_QUEUE")
	if config.QueueAddr == "" {
		config.QueueAddr = "0.0.0.0:11300"
	}
	config.SmtpServerAddr = os.Getenv("GLOW_SMTP_SERVER")
	config.MailFrom = os.Getenv("GLOW_MAIL_FROM")
	if config.MailFrom == "" {
		config.MailFrom = "glow@example.com"
	}
	return config
}

func (this *Configuration) Load() error {
	path := os.Getenv("GLOW_CONFIG")
	this.deps = make(map[string][]string)
	deps, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(deps, &this.deps)
}

var Config = NewConfig()
