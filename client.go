package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/nutrun/lentil"
	"log"
	"path/filepath"
	"strconv"
)

type Client struct {
	q       *lentil.Beanstalkd
	verbose bool
}

func NewClient(verbose bool) (*Client, error) {
	this := new(Client)
	q, err := lentil.Dial(Config.QueueAddr)
	if err != nil {
		return nil, err
	}
	this.q = q
	this.verbose = verbose
	return this, nil
}

func (this *Client) put(cmd, mailto, workdir, out, tube string, priority, delay int) error {
	msg := make(map[string]string)
	msg["cmd"] = cmd
	msg["mailto"] = mailto
	msg["tube"] = tube
	msg["priority"] = fmt.Sprintf("%d", priority)
	if tube == "" {
		return errors.New("Missing required param -tube")
	}
	msg["delay"] = fmt.Sprintf("%d", delay) // Not used except for debugging
	workdir, e := filepath.Abs(workdir)
	if e != nil {
		return e
	}
	msg["workdir"] = workdir
	msg["out"] = out
	message, e := json.Marshal(msg)
	if this.verbose {
		log.Printf("QUEUEING UP: %s\n", message)
	}
	if e != nil {
		return e
	}
	if tube != "default" {
		e = this.q.Use(tube)
		if e != nil {
			return e
		}
	}
	_, e = this.q.Put(priority, delay, 60*60, message) // An hour TTR?
	return e
}

func (this *Client) putMany(input []byte) error {
	jobs := make([]map[string]string, 0)
	e := json.Unmarshal(input, &jobs)
	if e != nil {
		return e
	}
	for _, job := range jobs {
		priority := 0
		if priorityStr, exists := job["major"]; exists {
			priority, e = strconv.Atoi(priorityStr)
			if e != nil {
				return e
			}
		}
		delay := 0
		if delayStr, exists := job["delay"]; exists {
			delay, e = strconv.Atoi(delayStr)
			if e != nil {
				return e
			}
		}
		out, exists := job["out"]
		if !exists {
			out = "/dev/null"
		}
		workdir := "/tmp"
		dir, exists := job["workdir"]
		if exists {
			workdir = dir
		}
		e = this.put(job["cmd"], job["mailto"], workdir, out, job["tube"], priority, delay)
		if e != nil {
			return e
		}
	}
	return nil
}

func (this *Client) stats() error {
	q := NewJobQueue(this.q, false, make([]string, 0))
	stats, err := json.Marshal(q)
	if err != nil {
		return err
	}
	buffer := bytes.NewBufferString("")
	err = json.Indent(buffer, stats, "", "\t")
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", buffer.String())
	return nil
}
