package main

import (
	"encoding/json"
	"github.com/nutrun/lentil"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type Listener struct {
	Runner
	verbose  bool
	jobqueue *JobQueue
	sig      os.Signal
}

func NewListener(verbose, inclusive bool, filter []string) (*Listener, error) {
	this := new(Listener)
	q, err := lentil.Dial(Config.QueueAddr)

	if err != nil {
		return nil, err
	}

	this.q = q
	this.verbose = verbose
	this.jobqueue = NewJobQueue(this.q, inclusive, filter)
	return this, nil
}

func (this *Listener) run() {
	go this.trap()

listenerloop:
	for {
		if this.sig != nil {
			os.Exit(0)
		}
		Config.Load()
		job, e := this.jobqueue.Next()
		if e != nil {
			if strings.Contains(e.Error(), "TIMED_OUT") {
				time.Sleep(time.Second)
				goto listenerloop
			}
			log.Fatal(e)
		}
		if this.verbose {
			log.Printf("RUNNING: %s", job.Body)
		}
		msg := new(Message)
		e = json.Unmarshal([]byte(job.Body), &msg)
		if e != nil {
			this.catch(msg, e)
		}
		e = this.execute(msg)
		if e == nil {
			if this.verbose {
				log.Printf("COMPLETE: %s", job.Body)
			}
		} else {
			log.Printf("FAILED: %s", job.Body)
		}
		e = this.jobqueue.Delete(job.Id)
		if e != nil {
			this.catch(msg, e)
		}
	}
}

// Wait for currently running job to finish before exiting on SIGTERM and SIGINT
func (this *Listener) trap() {
	receivedSignal := make(chan os.Signal)
	signal.Notify(receivedSignal, syscall.SIGTERM, syscall.SIGINT)
	this.sig = <-receivedSignal
	if this.sig.String() == syscall.SIGTERM.String() {
		log.Printf("Got signal %d. Killing current job.", this.sig)
		if this.proc != nil {
			this.proc.Kill()
		}
		os.Exit(1)
	}
	go this.trap()
	log.Printf("Got signal %d. Waiting for current job to complete.\n", this.sig)
}
