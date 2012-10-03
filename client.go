package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/nutrun/lentil"
	"os"
	"strings"
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

func (this *Client) put(msg *Message) error {
	message, e := json.Marshal(msg)
	if this.verbose {
		fmt.Fprintf(os.Stderr, "QUEUEING UP: %s\n", message)
	}
	if e != nil {
		return e
	}
	if msg.Tube != "default" {
		e = this.q.Use(msg.Tube)
		if e != nil {
			return e
		}
	}
	_, e = this.q.Put(msg.Priority, msg.Delay, 60*60, message) // An hour TTR?
	return e
}

func (this *Client) putMany(input []byte) error {
	jobs, e := MessagesFromJSON(input)
	if e != nil {
		return e
	}
	for _, job := range jobs {
		e = this.put(job)
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

func (this *Client) drain(tubes string) error {
	drainedJobs := []byte("[\n")
	isFirstDrained := true
	for _, tube := range strings.Split(tubes, ",") {
		_, err := this.q.Watch(tube)
		if err != nil {
			return err
		}
		_, err = this.q.Ignore("default")
		if err != nil {
			return err
		}

		for job, err := this.q.ReserveWithTimeout(0); err == nil; job, err = this.q.ReserveWithTimeout(0) {
			this.q.Delete(job.Id)
			if !isFirstDrained {
				drainedJobs = append(drainedJobs, []byte(",\n")...)
			}
			drainedJobs = append(drainedJobs, job.Body...)
			isFirstDrained = false
		}
		if err != nil {
			return err
		}
	}
	drainedJobs = append(drainedJobs, []byte("\n]")...)
	fmt.Fprintf(os.Stderr, "%s", string(drainedJobs))
	return nil
}

func (this *Client) pause(tubes string, delay int) error {
	for _, tube := range strings.Split(tubes, ",") {
		e := this.q.PauseTube(tube, delay)
		if e != nil {
			return e
		}
		fmt.Fprintf(os.Stderr, "Paused %s for %d seconds", tubes, delay)
	}
	return nil
}
