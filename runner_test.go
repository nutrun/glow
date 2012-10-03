package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"testing"
)

func TestRunnerOutput(t *testing.T) {
	devnull, e := os.OpenFile(os.DevNull, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if e != nil {
		t.Fatal(e)
	}
	l := log.New(devnull, "", log.LstdFlags)
	runner, e := NewRunner(false, l)
	if e != nil {
		t.Fatal(e)
	}
	msg, e := createTestMessage("echo you suck", "test.out", ".")
	if e != nil {
		t.Fatal(e)
	}
	runner.execute(msg)
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

func TestRunnerShouldPutErrorOnBeanstalk(t *testing.T) {
	devnull, err := os.OpenFile(os.DevNull, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		t.Fatal(err)
	}
	l := log.New(devnull, "", log.LstdFlags)
	runner, err := NewRunner(false, l)
	if err != nil {
		t.Fatal(err)
	}
	log.SetOutput(bytes.NewBufferString(""))
	msg, e := createTestMessage("lsdonmybrain", "test.out", ".")
	if e != nil {
		t.Fatal(e)
	}
	runner.execute(msg)
	runner.q.Watch(Config.errorQueue)
	failed, err := runner.q.ReserveWithTimeout(0)
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
	runner.q.Delete(failed.Id)
	err = os.Remove("test.out")
	if err != nil {
		t.Fatal(err)
	}
}
