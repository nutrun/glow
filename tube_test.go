package main

import (
	"encoding/json"
	"testing"
)

func TestNestedDependencies(t *testing.T) {
	resetConfig() // Implemented in jobqueue_test.go
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

func TestLargeDependencyGraph(t *testing.T) {
	resetConfig() // Implemented in jobqueue_test.go
	json.Unmarshal(krazijson(), &Config.deps)
	queue := make(map[string]*Tube)
	for k, _ := range Config.deps {
		queue[k] = NewTube(k, 0, 0, 0)
	}
	queue["SessionTriggers"].ready = 1
	queue["ValidateBeast"].ready = 1
	if !queue["SessionTriggers"].Ready(queue) {
		t.Error("y u no redi?")
	}
	if queue["ValidateBeast"].Ready(queue) {
		t.Error("y u redi?")
	}
	queue["FillRsyncBeast"].ready = 1
	if queue["FillRsyncBeast"].Ready(queue) {
		t.Error("y u redi?")
	}
}

func krazijson() []byte {
	return []byte(`{"FillEnchilada":["Arca","ArcaTrade","Arcaxdp","Bats","Byx","Edga","Edgx","Nasdaq","Openbook","Cme","Ice","Hotspot","Creditsuisse","Currenex","Ebs","Brokertec"],"FillRsync":["Arca","ArcaTrade","Arcaxdp","Bats","Byx","Edga","Edgx","Nasdaq","Openbook","Cme","Ice","Hotspot","Creditsuisse","Currenex","Ebs","FillEnchilada","Books","Trades","ImpliedTrades"],"FXDatahook":["FillEnchilada","Books","Trades","ImpliedTrades","FillRsync","Rsync"],"OrderJoin":["SessionSplit","Orders"],"Hedge":["OrderJoin"],"ExtractQuotes":["ExtractCausalities"],"ExtractArcaQuotes":["ExtractCausalities"],"ExtractCmeQuotes":["ExtractCausalities"],"GatherQuotes":["ExtractArcaQuotes","ExtractCmeQuotes","ExtractQuotes"],"FillSessionTriggers":["ExtractCausalities","GatherQuotes"],"SessionTriggers":["FillSessionTriggers"],"ArcaSessionTriggers":["FillSessionTriggers"],"CmeSessionTriggers":["FillSessionTriggers"],"TriggerJoin":["SessionTriggers","ArcaSessionTriggers","CmeSessionTriggers","OrderJoin"],"ProdSplit":["Hedge","TriggerJoin","Orders"],"TcpStats":["SessionSplit"],"ValidateBeast":["ProdSplit"],"FillRsyncBeast":["OrderJoin","TcpStats","Hedge","ExtractCausalities","ExtractQuotes","ExtractCmeQuotes","ExtractArcaQuotes","GatherQuotes","FillSessionTriggers","TriggerJoin","ProdSplit","ValidateBeast"]}`)
}
