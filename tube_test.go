package main

import "testing"

func TestNestedDependencies(t *testing.T) {
	Config.deps["tube1"] = []string{"tube2"}
	Config.deps["tube2"] = []string{"tube3"}
	queue := make(map[string]*Tube)
	queue["tube1"] = NewTube("tube1", 0, 1, 0)
	// All deps are empty
	queue["tube2"] = NewTube("tube2", 0, 0, 0)
	queue["tube3"] = NewTube("tube3", 0, 0, 0)
	if !queue["tube1"].Ready(queue) {
		t.Error("y u no redi?")
	}
	queue["tube3"].reserved = 1 // Dep of dep has one reserved job
	if queue["tube1"].Ready(queue) {
		t.Error("y u redi?")
	}
	queue["tube3"].reserved = 0 // Empty deps
	if !queue["tube1"].Ready(queue) {
		t.Error("y u no redi?")
	}
	queue["tube3"].delayed = 4 // Dep of dep has delayed jobs
	if queue["tube1"].Ready(queue) {
		t.Error("y u redi?")
	}
	// All deps have jobs
	queue["tube2"].ready = 1
	if queue["tube1"].Ready(queue) {
		t.Error("y u redi?")
	}
}
