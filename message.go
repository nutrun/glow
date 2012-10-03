package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const MAX_OUTFILE_READ_LEN = 16 * 1024

type Message struct {
	Executable string   `json:"cmd"`
	Arguments  []string `json:"args"`
	Mailto     string   `json:"mailto"`
	Workdir    string   `json:"workdir"`
	Stdout     string   `json:"stdout"`
	Stderr     string   `json:"stderr"`
	Tube       string   `json:"tube"`
	Priority   int      `json:"pri"`
	Delay      int      `json:"delay"`
}

func NewMessage(executable string, args []string, mailto, workdir, stdout, stderr, tube string, pri, delay int) (*Message, error) {
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
	if stdout == "" {
		stdout = "/dev/null"
	}
	if stderr == "" {
		stderr = "/dev/null"
	}
	return &Message{executable, args, mailto, absoluteWorkdir, stdout, stderr, tube, pri, delay}, nil
}

func MessagesFromJSON(jsonstr []byte) ([]*Message, error) {
	vals := make([]*Message, 0)
	e := json.Unmarshal(jsonstr, &vals)
	if e != nil {
		return nil, e
	}
	messages := make([]*Message, len(vals))
	for i, m := range vals {
		msg, e := NewMessage(m.Executable, m.Arguments, m.Mailto, m.Workdir, m.Stdout, m.Stderr, m.Tube, m.Priority, m.Delay)
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

// Read up to MAX_OUTFILE_READ_LEN from the files we send stdout or stderr to
func (this *Message) readOutputFile(path string) ([]byte, error) {
	if path == "/dev/stdout" || path == "/dev/stderr" {
		return []byte{}, nil
	}
	f, e := os.Open(path)
	if e != nil {
		return nil, e
	}
	br := bufio.NewReader(f)
	lr := &io.LimitedReader{br, MAX_OUTFILE_READ_LEN}
	buf := make([]byte, MAX_OUTFILE_READ_LEN)
	n, e := lr.Read(buf)
	if e != nil {
		return nil, e
	}
	return buf[:n], nil
}

func (this *Message) readOut() string {
	hostname, _ := os.Hostname()
	content := make([]byte, 0)
	content = append(content, []byte(fmt.Sprintf("hostname: %v\n", hostname))...)
	stdout, e := this.readOutputFile(this.Stdout)
	if e != nil {
		content = append(content, []byte(
			fmt.Sprintf("Could not read stdout output from [%s]. %s\n", this.Stdout, e))...)
	} else {
		content = append(content, []byte("STDOUT:\n")...)
		content = append(content, stdout...)
		content = append(content, []byte("\n")...)
	}
	stderr, e := this.readOutputFile(this.Stderr)
	if e != nil {
		content = append(content, []byte(
			fmt.Sprintf("Could not read stderr output from [%s]. %s\n", this.Stderr, e))...)
	} else {
		content = append(content, []byte("STDERR:\n")...)
		content = append(content, stderr...)
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
