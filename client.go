package main

import (
	"encoding/json"
	beanstalk "github.com/nutrun/beanstalk.go"
	"path/filepath"
	"log"
)

type Client struct {
	q *beanstalk.Conn
}

func NewClient() (*Client, error) {
	this := new(Client)
	q, err := beanstalk.Dial(Config.QueueAddr)
	if err != nil {
		return nil, err
	}
	this.q = q
	return this, nil
}

func (this *Client) put(cmd, mailto, workdir string) error {
	msg := make(map[string]string)
	msg["cmd"] = cmd
	msg["mailto"] = mailto
	if workdir == "" {
		workdir = "."
	}
	workdir, e := filepath.Abs(workdir)
	if e != nil {
		return e
	}
	msg["workdir"] = workdir
	msg["out"] = "glow.out"
	message, e := json.Marshal(msg)
	log.Printf("RUNNING: %s\n", message)
	if e != nil {
		return e
	}
	_, e = this.q.Put(string(message), 0, 0, 1000*60*60) // An hour TTR?
	return e
}
