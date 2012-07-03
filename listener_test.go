package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"
)

func createTestMessage(cmd, out, workdir string) *Message {
	tokens := strings.Split(cmd, " ")
	return &Message{tokens[0], tokens[1:len(tokens)], "", workdir, out, "", 0, 0}
}

func TestOutput(t *testing.T) {
	listener := new(Listener)
	msg := createTestMessage("echo you suck", "test.out", ".")
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
	listener, err := NewListener(false, false, []string{})
	if err != nil {
		t.Fatal(err)
	}
	log.SetOutput(bytes.NewBufferString(""))
	msg := createTestMessage("lsdonmybrain", "test.out", ".")
	listener.execute(msg)
	listener.q.Watch(Config.errorQueue)
	failed, err := listener.q.ReserveWithTimeout(0)
	if err != nil {
		t.Fatal(err)
	}
	result := new(GlerrMessage)
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
