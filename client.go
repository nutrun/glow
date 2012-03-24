package main

import (
	"encoding/json"
	"log"
	"path/filepath"
)

type Client struct {
	q *Conn
}

func NewClient() (*Client, error) {
	this := new(Client)
	q, err := Dial(Config.QueueAddr)
	if err != nil {
		return nil, err
	}
	this.q = q
	return this, nil
}

func (this *Client) put(cmd, mailto, workdir, out, tube string) error {
	msg := make(map[string]string)
	msg["cmd"] = cmd
	msg["mailto"] = mailto
	msg["tube"] = tube
	workdir, e := filepath.Abs(workdir)
	if e != nil {
		return e
	}
	msg["workdir"] = workdir
	msg["out"] = out
	message, e := json.Marshal(msg)
	log.Printf("RUNNING: %s\n", message)
	if e != nil {
		return e
	}
	if tube != "default" {
		t, e := NewTube(this.q, tube)
		if e != nil {
			return e
		}
		_, e = t.Put(string(message), 0, 0, 1000*60*60) // An hour TTR?
	} else {
		_, e = this.q.Put(string(message), 0, 0, 1000*60*60) // An hour TTR?
	}
	return e
}
