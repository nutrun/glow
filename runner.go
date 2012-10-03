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
	q       *lentil.Beanstalkd
	proc    *os.Process
	verbose bool
	logger  *log.Logger
}

func NewRunner(verbose bool, logger *log.Logger) (*Runner, error) {
	this := new(Runner)
	this.logger = logger
	q, err := lentil.Dial(Config.QueueAddr)

	if err != nil {
		return nil, err
	}

	this.q = q
	this.verbose = verbose
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

	stdoutDir := filepath.Dir(msg.Stdout)
	os.MkdirAll(stdoutDir, 0755)
	stderrDir := filepath.Dir(msg.Stderr)
	os.MkdirAll(stderrDir, 0755)
	stdoutF, e := os.OpenFile(msg.Stdout, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if e != nil {
		this.catch(msg, e)
		return e
	}
	defer stdoutF.Close()
	stderrF, e := os.OpenFile(msg.Stderr, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if e != nil {
		this.catch(msg, e)
		return e
	}
	defer stderrF.Close()
	cmd.Stdout = stdoutF
	cmd.Stderr = stderrF
	if this.verbose {
		this.logger.Printf("INFO: Running command '%s %s'\n", msg.Executable, strings.Join(msg.Arguments, " "))
		this.logger.Printf("INFO: STDOUT to %s\n", msg.Stdout)
		this.logger.Printf("INFO: STDERR to %s\n", msg.Stderr)
	}
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
	this.logger.Printf("ERROR: %s\n", e)
	this.mail(msg, e)
	this.publishError(msg, e)
}

func (this *Runner) publishError(msg *Message, e error) {
	err := this.q.Use(Config.errorQueue)
	if err != nil {
		this.logger.Printf("ERROR: %s\n", err)
		return
	}
	payload, err := json.Marshal(NewErrMessage(msg, e))
	if err != nil {
		this.logger.Printf("ERROR: %s\n", err)
		return
	}
	_, err = this.q.Put(0, 0, 60*60, payload)
	if err != nil {
		this.logger.Printf("ERROR: %s\n", err)
	}
}

func (this *Runner) mail(msg *Message, e error) {
	if Config.SmtpServer == "" {
		return
	}
	if len(msg.Mailto) < 1 { //no email addresses
		return
	}
	to := strings.Split(msg.Mailto, ",")
	subject := fmt.Sprintf("Subject: FAILED: %s\r\n\r\n", msg.getCommand())
	hostname, _ := os.Hostname()
	mail := fmt.Sprintf("%s%s", subject, fmt.Sprintf("Ran on [%s]\n%s\n%s\n%s", hostname, subject, e, msg.readOut()))
	e = smtp.SendMail(Config.SmtpServer, nil, Config.MailFrom, to, []byte(mail))
	if e != nil {
		this.logger.Printf("ERROR: %s\n", e)
	}
}
