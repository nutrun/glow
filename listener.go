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
)

type Listener struct {
	q       *lentil.Beanstalkd
	stopped bool
	verbose bool
}

func NewListener(verbose bool) (*Listener, error) {
	this := new(Listener)
	q, err := lentil.Dial(Config.QueueAddr)

	if err != nil {
		return nil, err
	}

	this.q = q
	this.verbose = verbose
	return this, nil
}

func (this *Listener) run() {
	go this.trap()
	jobqueue := NewJobQueue(this.q)

listenerloop:
	for {
		if this.stopped {
			os.Exit(0)
		}
		job, e := jobqueue.Next()
		if e != nil {
			if strings.Contains(e.Error(), "TIMED_OUT") {
				goto listenerloop
			}
			log.Fatal(e)
		}
		if this.verbose {
			log.Printf("RUNNING: %s", job.Body)
		}
		msg := make(map[string]string)
		json.Unmarshal([]byte(job.Body), &msg)
		e = os.MkdirAll(msg["workdir"], os.ModePerm)
		if e != nil {
			this.catch(msg, e)
			jobqueue.Delete(job.Id)
			goto listenerloop
		}
		os.Chdir(msg["workdir"])
		messagetokens := strings.Split(msg["cmd"], " ")
		command := messagetokens[0]
		args := messagetokens[1:len(messagetokens)]
		cmd := exec.Command(command, args...)
		out, e := cmd.CombinedOutput()
		if e != nil {
			this.catch(msg, e)
		}
		e = jobqueue.Delete(job.Id)
		if e != nil {
			this.catch(msg, e)
		}
		e = os.RemoveAll(msg["workdir"])
		if e != nil {
			this.catch(msg, e)
		}
		if len(out) > 0 {
			e = ioutil.WriteFile(msg["out"], out, 0644)
			if e != nil {
				this.catch(msg, e)
			}
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
	mail := fmt.Sprintf("%s%s", subject, fmt.Sprintf("%s", e))
	e = smtp.SendMail(Config.SmtpServerAddr, nil, Config.MailFrom, to, []byte(mail))
	if e != nil {
		log.Printf("ERROR: %s\n", e)
	}
}

// Wait for currently running job to finish before exiting on SIGTERM and SIGINT
func (this *Listener) trap() {
	receivedSignal := make(chan os.Signal)
	signal.Notify(receivedSignal, syscall.SIGTERM, syscall.SIGINT)

	for {
		sig := <-receivedSignal
		log.Printf("Got signal %d. Waiting for current job to complete.", sig)
		this.stopped = true
	}
}
