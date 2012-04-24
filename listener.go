package main

import (
	"encoding/json"
	"fmt"
	"github.com/nutrun/lentil"
	"io/ioutil"
	"log"
	"net/smtp"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type Listener struct {
	q        *lentil.Beanstalkd
	stopped  bool
	verbose  bool
	jobqueue *JobQueue
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

func (this *Listener) execute(msg map[string]string) {
	workdir := msg["workdir"]
	e := os.Chdir(workdir)
	if e != nil {
		this.catch(msg, e)
	}
	messagetokens := strings.Split(msg["cmd"], " ")
	command := messagetokens[0]
	args := messagetokens[1:len(messagetokens)]
	cmd := exec.Command(command, args...)
	f, e := os.Create(msg["out"])
	if e != nil {
		this.catch(msg, e)
	}
	cmd.Stderr = f
	cmd.Stdout = f
	e = cmd.Run()
	if e != nil {
		this.catch(msg, e)
	}
	f.Close()
}

func (this *Listener) run() {
	go this.trap()

listenerloop:
	for {
		if this.stopped {
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
		msg := make(map[string]string)
		json.Unmarshal([]byte(job.Body), &msg)
		this.execute(msg)
		e = this.jobqueue.Delete(job.Id)
		if e != nil {
			this.catch(msg, e)
		}
	}
}

// Log and email errors
func (this *Listener) catch(msg map[string]string, e error) {
	log.Printf("ERROR: %s\n", e.Error())
	if Config.SmtpServerAddr == "" {
		return
	}
	if len(msg["mailto"]) < 1 { //no email addresses
		return
	}
	to := strings.Split(msg["mailto"], ",")
	subject := fmt.Sprintf("Subject: FAILED: %s\r\n\r\n", msg["cmd"])
	hostname, _ := os.Hostname()
	content := make([]byte, 0)
	info, err := os.Stat(msg["out"])
	if err != nil {
		content = []byte(fmt.Sprintf("Could not read job log from [%s]. %s", msg["out"], err.Error()))
	}
	if info.Size() > 1024 {
		content = []byte(fmt.Sprintf("Could send job log [%s]. File too big", msg["out"]))

	} else {
		content, err = ioutil.ReadFile(msg["out"])
		if err != nil {
			content = []byte(fmt.Sprintf("Could not read job log from [%s]", msg["out"]))
		}
	}
	mail := fmt.Sprintf("%s%s", subject, fmt.Sprintf("Ran on [%s]\n%s\n%s", hostname, e, content))
	e = smtp.SendMail(Config.SmtpServerAddr, nil, Config.MailFrom, to, []byte(mail))
	if e != nil {
		log.Printf("ERROR: %s\n", e)
	}
}

// Wait for currently running job to finish before exiting on SIGTERM and SIGINT
func (this *Listener) trap() {
	receivedSignal := make(chan os.Signal)
	signal.Notify(receivedSignal, syscall.SIGTERM, syscall.SIGINT)
	sig := <-receivedSignal
	log.Printf("Got signal %d. Waiting for current job to complete.", sig)
	this.stopped = true
}
