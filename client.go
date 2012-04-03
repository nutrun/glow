package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"lentil"
	"log"
	"os"
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

func (this *Client) put(cmd, mailto, workdir, out, tube string, pri int) error {
	msg := make(map[string]string)
	msg["cmd"] = cmd
	msg["mailto"] = mailto
	msg["tube"] = tube
	if tube == "" {
		return errors.New("Missing required param -tube")
	}
	msg["pri"] = string(pri) // Not used except for debugging
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
	_, e = this.q.Put(pri, 0, 60*60, message) // An hour TTR?
	return e
}

func (this *Client) putMany(input []byte) error {
	jobs := make([]map[string]string, 0)
	e := json.Unmarshal(input, &jobs)
	if e != nil {
		return e
	}
	for _, job := range jobs {
		pri := 0
		priorityStr, exists := job["pri"]
		if exists {
			pri, e = strconv.Atoi(priorityStr)
			if e != nil {
				return e
			}
		}
		out, exists := job["out"]
		if !exists {
			out = "/dev/null"
		}
		e = this.put(job["cmd"], job["mailto"], job["workdir"], out, job["tube"], pri)
		if e != nil {
			return e
		}
	}
	return nil
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
