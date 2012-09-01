package main

import (
	"encoding/json"
	"fmt"
	"github.com/nutrun/lentil"
	"log"
	"net/smtp"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Runner struct {
	q    *lentil.Beanstalkd
	proc *os.Process
}

func NewRunner() (*Runner, error) {
	this := new(Runner)
	q, err := lentil.Dial(Config.QueueAddr)

	if err != nil {
		return nil, err
	}

	this.q = q
	return this, nil
}

func (this *Runner) execute(msg *Message) error {
	workdir := msg.Workdir
	e := os.Chdir(workdir)
	if e != nil {
		this.catch(msg, e)
		return e
	}

	cmd := exec.Command(msg.Executable, msg.Arguments...)

	outputDir := filepath.Dir(msg.Out)
	os.MkdirAll(outputDir, 0755)

	f, e := os.Create(msg.Out)
	if e != nil {
		this.catch(msg, e)
		return e
	}

	defer f.Close()

	cmd.Stderr = f
	cmd.Stdout = f

	e = cmd.Start()
	if e != nil {
		this.catch(msg, e)
		return e
	}

	this.proc = cmd.Process
	e = cmd.Wait()

	if e != nil {
		this.catch(msg, e)
	}
	this.proc = nil
	return e
}

// Log and email errors
func (this *Runner) catch(msg *Message, e error) {
	log.Printf("ERROR: %s\n", e.Error())
	this.mail(msg, e)
	this.publishError(msg, e)
}

func (this *Runner) publishError(msg *Message, e error) {
	err := this.q.Use(Config.errorQueue)
	if err != nil {
		log.Printf("ERROR: %s\n", err)
		return
	}
	payload, err := json.Marshal(NewErrMessage(msg, e))
	if err != nil {
		log.Printf("ERROR: %s\n", err)
		return
	}
	_, err = this.q.Put(0, 0, 60*60, payload)
	if err != nil {
		log.Printf("ERROR: %s\n", err)
	}
}

func (this *Runner) mail(msg *Message, e error) {
	if Config.SmtpServerAddr == "" {
		return
	}
	if len(msg.Mailto) < 1 { //no email addresses
		return
	}
	to := strings.Split(msg.Mailto, ",")
	subject := fmt.Sprintf("Subject: FAILED: %s\r\n\r\n", msg.getCommand())
	hostname, _ := os.Hostname()
	mail := fmt.Sprintf("%s%s", subject, fmt.Sprintf("Ran on [%s]\n%s\n%s\n%s", hostname, subject, e, msg.readOut()))
	e = smtp.SendMail(Config.SmtpServerAddr, nil, Config.MailFrom, to, []byte(mail))
	if e != nil {
		log.Printf("ERROR: %s\n", e)
	}
}
