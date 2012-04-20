package main

import "os"

type Configuration struct {
	QueueAddr      string
	SmtpServerAddr string
	MailFrom       string
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

var Config = NewConfig()
