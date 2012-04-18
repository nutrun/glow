package main

import (
	"testing"
)

func TestOutputOfRun(t *testing.T) {
	listener := NewTestListener(false, []string{})
	msg := make(map[string]string)
	msg["cmd"] = "ls -lh"
	msg["out"] = "test.out"
	msg["workdir"] = "."
	listener.execute(msg)
}

func NewTestListener(inclusive bool, filter []string) *Listener {
	this := new(Listener)
	return this
}
