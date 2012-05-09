package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"testing"
)

func TestOutput(t *testing.T) {
	listener := new(Listener)
	msg := make(map[string]string)
	msg["cmd"] = "echo you suck"
	msg["out"] = "test.out"
	msg["workdir"] = "."
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
	msg := make(map[string]string)
	msg["cmd"] = "lsdonmybrain"
	msg["out"] = "test.out"
	msg["workdir"] = "."
	listener.execute(msg)
	listener.q.Watch(Config.errorQueue)
	failed, err := listener.q.ReserveWithTimeout(0)
	if err != nil {
		t.Fatal(err)
	}
	result := make(map[string]string)
	err = json.Unmarshal(failed.Body, &result)
	if result["cmd"] != "lsdonmybrain" {
		t.Errorf("Recieved Unexpected Msg [%v]", failed.Body)
	}
	listener.q.Delete(failed.Id)
	err = os.Remove("test.out")
	if err != nil {
		t.Fatal(err)
	}
}
