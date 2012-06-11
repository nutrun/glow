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
		// Copy the job body because it appears to be getting corrupted in this.q.Use inside publishError()
		job_desc := string(job.Body)
		if this.verbose {
			log.Printf("RUNNING: %s", job_desc)
		}
		msg := make(map[string]string)
		json.Unmarshal([]byte(job.Body), &msg)
		e = this.execute(msg)
		if e == nil {
			log.Printf("COMPLETE: %s", job_desc)
		} else {
			log.Printf("FAILED: %s", job_desc)
		}
		// Weird corruption problem - this is getting hit on job failure
		// if job_desc != string(job.Body) {
		// 	log.Fatalf("body has changed!")
		// }
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
	log.Printf("Got signal %d. Waiting for current job to complete. sig term is [%v]", this.sig, syscall.SIGTERM)
}
