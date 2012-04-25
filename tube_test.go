package main

import "testing"

func TestDependencies(t *testing.T) {
	resetConfig() // Implemented in jobqueue_test.go
	Config.deps["tube1"] = []string{"tube2"}
	queue := make(map[string]*Tube)
	queue["tube1"] = NewTube("tube1", 0, 1, 0)
	queue["tube2"] = NewTube("tube2", 0, 0, 0)
	if !queue["tube1"].Ready(queue) {
		t.Error("y u no redi?")
	}
	queue["tube2"].ready = 1
	if queue["tube1"].Ready(queue) {
		t.Error("y u no redi?")
	}
	queue["tube2"].ready = 0
	queue["tube2"].delayed = 1
	if queue["tube1"].Ready(queue) {
		t.Error("y u no redi?")
	}
}
