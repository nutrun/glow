package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"github.com/nutrun/lentil"
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

type Message struct {
    Command     string      `json:command`
    Executable  string      `json:executable`
    Arguments   []string    `json:arguments`
    Mailto      string      `json:mailto`
    Workdir     string      `json:workdir`
    Out         string      `json:out`
    Tube        string      `json:tube`
    Priority    int         `json:priority`
    Delay       int         `json:delay`
}

func isValidMessage(msg *Message) error {
    if msg.Command != "" && msg.Executable != "" {
        return errors.New("Found both executable and cmd in message. Don't know which one to use")
    }
    if msg.Command == "" && msg.Executable == "" {
        return errors.New("Neither executable nor cmd field provided in message")
    }
    if msg.Tube == "" {
		return errors.New("Missing required param -tube")
    }
    return nil
}

func (this *Client) put_message(msg *Message) error {
    if e := isValidMessage(msg); e != nil {
        return e
    }

	message, e := json.Marshal(msg)
	if this.verbose {
		log.Printf("QUEUEING UP: %s\n", message)
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

func (this *Client) put(cmd, executable string, arguments []string, mailto string, workdir, out, tube string, priority, delay int) error {
    workdir, e := filepath.Abs(workdir)
	if e != nil {
		return e
	}
    msg := &Message{cmd, executable, arguments, mailto, workdir, out, tube, priority, delay}
    return this.put_message(msg)
}

func (this *Client) putMany(input []byte) error {
	jobs := make([]*Message, 0)
	e := json.Unmarshal(input, &jobs)
	if e != nil {
		return e
	}
	for _, job := range jobs {
        if e := isValidMessage(job); e != nil {
            panic(e)
        }
        if job.Out == "" {
            job.Out = "/dev/null"
        }
		if job.Workdir == "" {
            job.Workdir = "/tmp"
        }

		e = this.put_message(job)
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

func (this *Client) drain(tube string) error {
	for _, tube := range strings.Split(tube, ",") {
		_, err := this.q.Watch(tube)
		if err != nil {
			return err
		}
		_, err = this.q.Ignore("default")
		if err != nil {
			return err
		}
		for job, err := this.q.ReserveWithTimeout(0); err == nil; job, err = this.q.ReserveWithTimeout(0) {
			log.Printf("DRAINED: %s", job.Body)
			this.q.Delete(job.Id)
		}
		if err != nil {
			return err
		}
	}
	return nil
}
