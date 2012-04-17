package main

import (
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

func (this *Client) put(cmd, mailto, workdir, out, tube string, major, minor, delay int) error {
	msg := make(map[string]string)
	msg["cmd"] = cmd
	msg["mailto"] = mailto
	msg["tube"] = tube
	if tube == "" {
		return errors.New("Missing required param -tube")
	}
	msg["major"] = fmt.Sprintf("%d", major) // Not used except for debugging
	msg["minor"] = fmt.Sprintf("%d", minor) // Not used except for debugging
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
		e = this.q.Use(fmt.Sprintf("%s_%d", tube, int((major<<16)|minor)))
		if e != nil {
			return e
		}
	}
	_, e = this.q.Put(int((major<<16)|minor), delay, 60*60, message) // An hour TTR?
	return e
}

func (this *Client) putMany(input []byte) error {
	jobs := make([]map[string]string, 0)
	e := json.Unmarshal(input, &jobs)
	if e != nil {
		return e
	}
	for _, job := range jobs {
		major := 0
		majorStr, exists := job["major"]
		if exists {
			major, e = strconv.Atoi(majorStr)
			if e != nil {
				return e
			}
		}
		minor := 0
		minorStr, exists := job["minor"]
		if exists {
			minor, e = strconv.Atoi(minorStr)
			if e != nil {
				return e
			}
		}
		delay := 0
		delayStr, exists := job["delay"]
		if exists {
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
		e = this.put(job["cmd"], job["mailto"], workdir, out, job["tube"], major, minor, delay)
		if e != nil {
			return e
		}
	}
	return nil
}

func (this *Client) stats() error {
	q := NewJobQueue(this.q)
	major := -1
	first := true
	q.Stats(func(tubes []*Tube) {
		if major != tubes[0].majorPriority {
			if major != -1 {
				fmt.Printf("\n\t},")
			} else {
				fmt.Printf("{")
			}
			first = true
			major = tubes[0].majorPriority
			fmt.Printf("\n   \"%v\": {", major)
		}
		if !first {
			fmt.Printf(",")
		}
		first = false
		fmt.Printf("\n      \"%v\": {", tubes[0].minorPriority)
		for i, tube := range tubes {
			if i != 0 {
				fmt.Printf(",")
			}
			fmt.Printf("\n         \"%v\":  {\"jobs-ready\" : %v }", tube.name, tube.jobs)
		}
		fmt.Printf("\n      }")
	})
	if major != -1 {
		fmt.Printf("\n   }\n}\n")
	} else {
		fmt.Printf("404: No tubs found\n")
	}
	return nil
}
