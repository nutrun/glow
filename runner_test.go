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
	runner := new(Runner)
	msg := createTestMessage("echo you suck", "test.out", ".")
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
	runner, err := NewRunner()
	if err != nil {
		t.Fatal(err)
	}
	log.SetOutput(bytes.NewBufferString(""))
	msg := createTestMessage("lsdonmybrain", "test.out", ".")
	runner.execute(msg)
	runner.q.Watch(Config.errorQueue)
	failed, err := runner.q.ReserveWithTimeout(0)
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
	runner.q.Delete(failed.Id)
	err = os.Remove("test.out")
	if err != nil {
		t.Fatal(err)
	}
}
