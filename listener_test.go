package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func createTestMessage(cmd, out, workdir string) (*Message, error) {
	tokens := strings.Split(cmd, " ")
	return NewMessage(tokens[0], tokens[1:len(tokens)], "", workdir, out, out, "testtube", 0, 0)
}

func TestOutput(t *testing.T) {
	listener := new(Listener)
	msg, e := createTestMessage("echo you suck", "test.out", ".")
	if e != nil {
		t.Fatal(e)
	}
	listener.execute(msg)
	out, e := ioutil.ReadFile("test.out")
	if e != nil {
		t.Fatal(e)
	}
	if string(out) != "you suck\n" {
		t.Errorf("[%s] isn't you suck", out)
	}
	e = os.Remove("test.out")
	if e != nil {
		t.Fatal(e)
	}
}

func TestPutErrorOnBeanstalk(t *testing.T) {
	listener, err := NewListener(false, false, []string{}, "/dev/null")
	if err != nil {
		t.Fatal(err)
	}
	msg, e := createTestMessage("lsdonmybrain", "test.out", ".")
	if e != nil {
		t.Fatal(e)
	}
	listener.execute(msg)
	listener.q.Watch(Config.errorQueue)
	failed, err := listener.q.ReserveWithTimeout(0)
	if err != nil {
		t.Fatal(err)
	}
	result := new(ErrMessage)
	err = json.Unmarshal(failed.Body, result)
	if err != nil {
		t.Fatal(err)
	}
	if result.Cmd != "lsdonmybrain" {
		t.Errorf("Recieved Unexpected Msg [%v]", string(failed.Body))
	}
	listener.q.Delete(failed.Id)
	err = os.Remove("test.out")
	if err != nil {
		t.Fatal(err)
	}
}
