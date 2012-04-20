package main

import (
	"io/ioutil"
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
	os.Remove("test.out")
	if e != nil {
		t.Fatal(e)
	}
}
