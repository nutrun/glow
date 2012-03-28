package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"lentil"
)

type Client struct {
	q *lentil.Beanstalkd
}

func NewClient() (*Client, error) {
	this := new(Client)
	q, err := lentil.Dial(Config.QueueAddr)
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
		e = this.q.Use(tube)
		if e != nil {
			return e
		}
	}
	_, e = this.q.Put(0, 0, 60*60, message) // An hour TTR?
	return e
}

func (this *Client) stats() error {
	tubes, e := this.q.ListTubes()
	if e != nil {
		return e
	}
	allstats := make([]map[string]string, 0)
	for _, tube := range tubes {
		stats, e := this.q.StatsTube(tube)
		if e != nil {
			return e
		}
		allstats = append(allstats, stats)
	}
	statsjson, e := json.Marshal(allstats)
	if e != nil {
		return e
	}
	fmt.Fprintf(os.Stderr, "%s\n", statsjson)
	return nil
}
