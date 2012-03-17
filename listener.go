package main

import (
	"encoding/json"
	"fmt"
	beanstalk "github.com/nutrun/beanstalk.go"
	"log"
	"net/smtp"
	"os"
	"os/exec"
	"strings"
)

type Listener struct {
	q *beanstalk.Conn
}

func NewListener(addr string) (*Listener, error) {
	this := new(Listener)
	q, err := beanstalk.Dial(addr)

	if err != nil {
		return nil, err
	}

	this.q = q
	return this, nil
}

func (this *Listener) run() {
	for {
		job, e := this.q.Reserve()

		if e != nil {
			job.Delete()
			panic(e)
		}

		msg := make(map[string]string)
		json.Unmarshal([]byte(job.Body), &msg)

		messagetokens := strings.Split(msg["cmd"], " ")
		command := messagetokens[0]
		args := messagetokens[1:len(messagetokens)]
		cmd := exec.Command(command, args...)
		out, e := cmd.CombinedOutput()

		if e != nil {
			if len(msg["mailto"]) > 0 {
				subject := fmt.Sprintf("FAILED: %s", msg["cmd"])
				to := strings.Split(msg["mailto"], ",")
				e := this.email(subject, fmt.Sprintf("%s", e), to)
				if e != nil {
					log.Printf("ERROR: %s\n", e)
				}
			}
		}

		fmt.Fprintf(os.Stderr, "%s", out)
		job.Delete()
	}
}

func (this *Listener) email(sbjct, body string, to []string) error {
	subject := fmt.Sprintf("Subject: %s\r\n\r\n", sbjct)
	msg := fmt.Sprintf("%s%s", subject, body)
	return smtp.SendMail("smtp.us.drwholdings.com:25", nil, "glow@drw.com", to, []byte(msg))
}
