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
	verbose  bool
	jobqueue *JobQueue
	sig      os.Signal
	proc     *os.Process
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
		return
	}
	messagetokens := strings.Split(msg["cmd"], " ")
	command := messagetokens[0]
	args := messagetokens[1:len(messagetokens)]
	cmd := exec.Command(command, args...)
	f, e := os.Create(msg["out"])
	if e != nil {
		this.catch(msg, e)
		return
	}

	defer f.Close()

	cmd.Stderr = f
	cmd.Stdout = f

	e = cmd.Start()
	if e != nil {
		this.catch(msg, e)
		return
	}

	this.proc = cmd.Process
	e = cmd.Wait()

	if e != nil {
		this.catch(msg, e)
	}
	this.proc = nil
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
	this.mail(msg, e)
	this.publishError(msg, e)
}

func (this *Listener) publishError(msg map[string]string, e error) {
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

func (this *Listener) readLog(msg map[string]string) string {
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

func (this *Listener) mail(msg map[string]string, e error) {
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
	log.Printf("Got signal %d. Waiting for current job to complete. sig term is [%v]", this.sig, syscall.SIGTERM)
}
