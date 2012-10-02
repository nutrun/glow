package main

import (
	"testing"
)

func TestTubeRequired(t *testing.T) {
	_, e := NewMessage("executable", []string{"arg"}, "", "", "", "", "", 0, 0)
	if e == nil {
		t.Errorf("Should have missing tube error")
	}
	if e.Error() != "Missing required param -tube" {
		t.Errorf("[%s] isn't [%s]", e.Error(), "Missing required param -tube")
	}
	// Test same thing from json
	jsonstr := `[{"cmd": "echo", "args": ["arg1", "arg2"], "out": "out.txt", "pri": 15, "delay": 120}]`
	_, e = MessagesFromJSON([]byte(jsonstr))
	if e == nil {
		t.Errorf("Should have missing tube error")
	}
	if e.Error() != "Missing required param -tube" {
		t.Errorf("[%s] isn't [%s]", e.Error(), "Missing required param -tube")
	}
}

func TestValidJSONUnmarshall(t *testing.T) {
	jsonstr := `[{"tube": "testtube", "cmd": "echo", "args": ["arg1", "arg2"], "stdout": "out.txt", "stderr": "err.txt", "pri": 15, "delay": 120}]`
	messages, e := MessagesFromJSON([]byte(jsonstr))
	if e != nil {
		t.Fatal(e)
	}
	if len(messages) != 1 {
		t.Errorf("[%d] isn't [%d]", len(messages), 1)
	}
	msg := messages[0]
	if msg.Tube != "testtube" {
		t.Errorf("[%s] isn't [%s]", msg.Tube, "testtube")
	}
	if msg.Executable != "echo" {
		t.Errorf("[%s] isn't [%s]", msg.Executable, "echo")
	}
	if msg.Arguments[0] != "arg1" {
		t.Errorf("[%s] isn't [%s]", msg.Arguments[0], "arg1")
	}
	if msg.Arguments[1] != "arg2" {
		t.Errorf("[%s] isn't [%s]", msg.Arguments[1], "arg2")
	}
	if msg.Stdout != "out.txt" {
		t.Errorf("[%s] isn't [%s]", msg.Stdout, "out.txt")
	}
	if msg.Stderr != "err.txt" {
		t.Errorf("[%s] isn't [%s]", msg.Stderr, "err.txt")
	}
	if msg.Priority != 15 {
		t.Errorf("[%d] isn't [%d]", msg.Priority, 15)
	}
	if msg.Delay != 120 {
		t.Errorf("[%d] isn't [%d]", msg.Delay, 120)
	}
}

func TestDefaultWorkdir(t *testing.T) {
	msg, e := NewMessage("executable", []string{"arg"}, "", "", "", "", "testtube", 0, 0)
	if e != nil {
		t.Fatal(e)
	}
	if msg.Workdir != "/tmp" {
		t.Errorf("[%s] isn't [%s]", msg.Workdir, "/tmp")
	}
}
