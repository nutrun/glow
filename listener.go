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
	logpath  string
	logfile  *os.File
}

func NewListener(verbose, inclusive bool, filter []string, logpath string) (*Listener, error) {
	this := new(Listener)
	this.logpath = logpath
	err := this.resetLog()
	if err != nil {
		return nil, err
	}
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
		e := Config.LoadDeps()
		if e != nil {
			this.logger.Fatalf("Error loading dependency config: %s\n", e)
		}
		job, e := this.jobqueue.Next()
		if e != nil {
			if strings.Contains(e.Error(), "TIMED_OUT") {
				time.Sleep(time.Second)
				goto listenerloop
			}
			this.logger.Fatal(e)
		}
		if this.verbose {
			this.logger.Printf("RUNNING: %s", job.Body)
		}
		msg := new(Message)
		e = json.Unmarshal([]byte(job.Body), &msg)
		if e != nil {
			this.catch(msg, e)
		}
		e = this.execute(msg)
		if e == nil {
			if this.verbose {
				this.logger.Printf("COMPLETE: %s", job.Body)
			}
		} else {
			this.logger.Printf("FAILED: %s", job.Body)
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
	signal.Notify(receivedSignal, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
	this.sig = <-receivedSignal
	if this.sig.String() == syscall.SIGTERM.String() {
		this.logger.Printf("Got signal %d. Killing current job.\n", this.sig)
		if this.proc != nil {
			this.proc.Kill()
		}
		os.Exit(1)
	} else if this.sig.String() == syscall.SIGHUP.String() {
		this.logger.Printf("Got signal %d. Reopening log.\n", this.sig)
		e := this.resetLog()
		if e != nil {
			panic(e)
		}
		this.sig = nil
	} else if this.sig.String() == syscall.SIGINT.String() {
		this.logger.Printf("Got signal %d. Waiting for current job to complete.\n", this.sig)
	}
	go this.trap()
}

func (this *Listener) resetLog() error {
	if this.logfile != nil {
		this.logfile.Close() // Ignoring this error...
	}
	logfile, e := os.OpenFile(this.logpath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if e != nil {
		return e
	}
	this.logfile = logfile
	this.logger = log.New(this.logfile, "", log.LstdFlags)
	return nil
}
