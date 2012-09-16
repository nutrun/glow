package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type Message struct {
	Executable string   `json:"cmd"`
	Arguments  []string `json:"args"`
	Mailto     string   `json:"mailto"`
	Workdir    string   `json:"workdir"`
	Out        string   `json:"out"`
	Tube       string   `json:"tube"`
	Priority   int      `json:"pri"`
	Delay      int      `json:"delay"`
}

func NewMessage(executable string, args []string, mailto, workdir, out, tube string, pri, delay int) (*Message, error) {
	if tube == "" {
		return nil, errors.New("Missing required param -tube")
	}
	if workdir == "" {
		workdir = "/tmp"
	}
	absoluteWorkdir, e := filepath.Abs(workdir)
	if e != nil {
		return nil, e
	}
	if out == "" {
		out = "/dev/null"
	}
	return &Message{executable, args, mailto, absoluteWorkdir, out, tube, pri, delay}, nil
}

func MessagesFromJSON(jsonstr []byte) ([]*Message, error) {
	vals := make([]*Message, 0)
	e := json.Unmarshal(jsonstr, &vals)
	if e != nil {
		return nil, e
	}
	messages := make([]*Message, len(vals))
	for i, m := range vals {
		msg, e := NewMessage(m.Executable, m.Arguments, m.Mailto, m.Workdir, m.Out, m.Tube, m.Priority, m.Delay)
		if e != nil {
			return nil, e
		}
		messages[i] = msg
	}
	return messages, nil
}

func (this *Message) getCommand() string {
	cmd := this.Executable
	if len(this.Arguments) > 0 {
		cmd += " " + strings.Join(this.Arguments, " ")
	}
	return cmd
}

func (this *Message) readOut() string {
	if this.Out == "/dev/stdout" || this.Out == "/dev/stderr" {
		return ""
	}
	hostname, _ := os.Hostname()
	content := make([]byte, 0)
	content = append(content, []byte(fmt.Sprintf("hostname: %v\n", hostname))...)

	info, err := os.Stat(this.Out)
	if err != nil {
		content = append(content, []byte(fmt.Sprintf("Could not read job log from [%s]. %s", this.Out, err.Error()))...)
		return string(content)
	}
	if info.Size() > 60000 {
		content = append(content, []byte(fmt.Sprintf("Could not send job log [%s]. File too big", this.Out))...)
	} else {
		content, err = ioutil.ReadFile(this.Out)
		if err != nil {
			content = append(content, []byte(fmt.Sprintf("Could not read job log from [%s]", this.Out))...)
		}
	}
	return string(content)
}

type ErrMessage struct {
	Cmd   string `json:"cmd"`
	Error string `json:"error"`
	Log   string `json:"log"`
}

func NewErrMessage(msg *Message, e error) *ErrMessage {
	return &ErrMessage{msg.getCommand(), e.Error(), msg.readOut()}
}
