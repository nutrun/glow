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
	"strings"
)

type Runner struct {
	q        *lentil.Beanstalkd
	proc     *os.Process
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

func (this *Runner) execute(msg map[string]string) error{
	workdir := msg["workdir"]
	e := os.Chdir(workdir)
	if e != nil {
		this.catch(msg, e)
		return e
	}
	messagetokens := strings.Split(msg["cmd"], " ")
	command := messagetokens[0]
	args := messagetokens[1:len(messagetokens)]
	cmd := exec.Command(command, args...)
	f, e := os.Create(msg["out"])
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
func (this *Runner) catch(msg map[string]string, e error) {
	log.Printf("ERROR: %s\n", e.Error())
	this.mail(msg, e)
	this.publishError(msg, e)
}

func (this *Runner) publishError(msg map[string]string, e error) {
	err := this.q.Use(Config.errorQueue)
	if err != nil {
		log.Printf("ERROR: %s\n", err)
		return
	}
	msg["error"] = e.Error()
	msg["log"] = this.readLog(msg)
	payload, err := json.Marshal(msg)
	if err != nil {
		log.Printf("ERROR: %s\n", err)
		return
	}
	_, err = this.q.Put(0, 0, 60*60, payload)
	if err != nil {
		log.Printf("ERROR: %s\n", err)
	}
}

func (this *Runner) readLog(msg map[string]string) string {
	content := make([]byte, 0)
	info, err := os.Stat(msg["out"])
	if err != nil {
		content = []byte(fmt.Sprintf("Could not read job log from [%s]. %s", msg["out"], err.Error()))
		return string(content)
	}
	if info.Size() > 104857 {
		content = []byte(fmt.Sprintf("Could send job log [%s]. File too big", msg["out"]))
	} else {
		content, err = ioutil.ReadFile(msg["out"])
		if err != nil {
			content = []byte(fmt.Sprintf("Could not read job log from [%s]", msg["out"]))
		}
	}
	return string(content)

}

func (this *Runner) mail(msg map[string]string, e error) {
	if Config.SmtpServerAddr == "" {
		return
	}
	if len(msg["mailto"]) < 1 { //no email addresses
		return
	}
	to := strings.Split(msg["mailto"], ",")
	subject := fmt.Sprintf("Subject: FAILED: %s\r\n\r\n", msg["cmd"])
	hostname, _ := os.Hostname()
	content := []byte(this.readLog(msg))
	mail := fmt.Sprintf("%s%s", subject, fmt.Sprintf("Ran on [%s]\n%s\n%s\n%s", hostname, subject, e, content))
	e = smtp.SendMail(Config.SmtpServerAddr, nil, Config.MailFrom, to, []byte(mail))
	if e != nil {
		log.Printf("ERROR: %s\n", e)
	}
}
