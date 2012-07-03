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

func getCommand(msg map[string]interface{}) *exec.Cmd{
    if msg["executable"] != "" {
        return exec.Command(msg["executable"].(string), msg["arguments"].([]string)...)
    }
    messagetokens := strings.Split(msg["cmd"].(string), " ")
    command := messagetokens[0]
    args := messagetokens[1:len(messagetokens)]
    return exec.Command(command, args...)
}

func (this *Runner) execute(msg map[string]interface{}) error {
	workdir := msg["workdir"].(string)
	e := os.Chdir(workdir)
	if e != nil {
		this.catch(msg, e)
		return e
	}

    cmd := getCommand(msg)

	f, e := os.Create(msg["out"].(string))
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
func (this *Runner) catch(msg map[string]interface{}, e error) {
	log.Printf("ERROR: %s\n", e.Error())
	this.mail(msg["mailTo"].(string), get_mail_subject(msg), msg["out"].(string), e)
	this.publishError(msg, e)
}

func get_mail_subject(msg map[string]interface{}) string {
    run_command := msg["cmd"].(string)
    if msg["executable"].(string) != "" {
        run_command = msg["executable"].(string) + strings.Join(msg["arguments"].([]string), " ")
    }

	return fmt.Sprintf("Subject: FAILED: %s\r\n\r\n", run_command)
}

func (this *Runner) publishError(msg map[string]interface{}, e error) {
	err := this.q.Use(Config.errorQueue)
	if err != nil {
		log.Printf("ERROR: %s\n", err)
		return
	}
	msg["error"] = e.Error()
	msg["log"] = this.readLog(msg["out"].(string))
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

func (this *Runner) readLog(out string) string {
	if out == "/dev/stdout" || out == "/dev/stderr" {
		return ""
	}
	content := make([]byte, 0)
	info, err := os.Stat(out)
	if err != nil {
		content = []byte(fmt.Sprintf("Could not read job log from [%s]. %s", out, err.Error()))
		return string(content)
	}
	if info.Size() > 104857 {
		content = []byte(fmt.Sprintf("Could send job log [%s]. File too big", out))
	} else {
		content, err = ioutil.ReadFile(out)
		if err != nil {
			content = []byte(fmt.Sprintf("Could not read job log from [%s]", out))
		}
	}
	return string(content)

}

func (this *Runner) mail(mailTo string, subject, out string, e error) {
	if Config.SmtpServerAddr == "" {
		return
	}
	if len(mailTo) < 1 { //no email addresses
		return
	}
	to := strings.Split(mailTo, ",")
	hostname, _ := os.Hostname()
	content := []byte(this.readLog(out))
	mail := fmt.Sprintf("%s%s", subject, fmt.Sprintf("Ran on [%s]\n%s\n%s\n%s", hostname, subject, e, content))
	e = smtp.SendMail(Config.SmtpServerAddr, nil, Config.MailFrom, to, []byte(mail))
	if e != nil {
		log.Printf("ERROR: %s\n", e)
	}
}
