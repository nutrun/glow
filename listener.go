package main

import (
	"encoding/json"
	"fmt"
	beanstalk "github.com/nutrun/beanstalk.go"
	"log"
	"net/smtp"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
)

type Listener struct {
	q       *beanstalk.Conn
	stopped bool
}

func NewListener() (*Listener, error) {
	this := new(Listener)
	q, err := beanstalk.Dial(Config.QueueAddr)

	if err != nil {
		return nil, err
	}

	this.q = q
	return this, nil
}

func (this *Listener) run() {
	this.handleSignals()

listenerloop:
	for {
		if this.stopped {
			os.Exit(0)
		}

		// Timeout every 1 second to handle kill signals
		job, e := this.q.ReserveWithTimeout(1 * 1000 * 1000) // 1 second

		if e != nil {
			if strings.Contains(e.Error(), "TIMED_OUT") {
				goto listenerloop
			}
			log.Fatal(e)
		}
		log.Printf("RUNNING: %s", job.Body)
		msg := make(map[string]string)
		json.Unmarshal([]byte(job.Body), &msg)
		e = os.Chdir(msg["workdir"])
		if e != nil {
			this.failmail(msg, e)
			job.Delete()
			goto listenerloop
		}
		messagetokens := strings.Split(msg["cmd"], " ")
		command := messagetokens[0]
		args := messagetokens[1:len(messagetokens)]
		cmd := exec.Command(command, args...)
		_, e = cmd.CombinedOutput()
		if e != nil {
			this.failmail(msg, e)
		}
		job.Delete()
	}
}

func (this *Listener) failmail(msg map[string]string, e error) {
	log.Print(e)
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

func (this *Listener) handleSignals() {
	receivedSignal := make(chan os.Signal)
	signal.Notify(receivedSignal, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		for {
			sig := <-receivedSignal
			log.Printf("Got signal %d. Waiting for current job to complete.", sig)
			this.stopped = true
		}
	}()
}
