package main

import (
	"beanstalk"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const beanstalkaddr = "0.0.0.0:11300"

var listener *bool = flag.Bool("l", false, "Start listener")
var help *bool = flag.Bool("h", false, "Show help")

func main() {
	flag.Parse()

	if *listener {
		l, e := NewListener(beanstalkaddr)
		handleErr(e)
		l.run()
	} else if *help || len(flag.Args()) == 0 {
		flag.Usage()
		os.Exit(1)
	} else {
		l, e := NewListener(beanstalkaddr)
		handleErr(e)
		e = l.put(flag.Args())
		handleErr(e)
		os.Exit(0)
	}
}

func handleErr(e error) {
    if e != nil {
		fmt.Fprintln(os.Stderr, e.Error())
		os.Exit(1)
	}
}

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

		messagetokens := strings.Split(job.Body, " ")
		command := messagetokens[0]
		args := messagetokens[1:len(messagetokens)]
		cmd := exec.Command(command, args...)
		out, e := cmd.CombinedOutput()

		if e != nil {
			panic(e)
		}

		fmt.Fprintf(os.Stderr, "%s", out)
		job.Delete()
	}
}

func (this *Listener) put(args []string) error {
    cmd := strings.Join(args, " ")
	_, e := this.q.Put(cmd, 0, 0, 1000*60*60) // An hour TTR?
	return e
}