package main

import (
	"encoding/json"
)

type Tube struct {
	deps     []string
	name     string
	reserved uint
	ready    uint
	delayed  uint
}

func NewTube(name string, reserved, ready, delayed uint) *Tube {
	this := new(Tube)
	this.name = name
	this.reserved = reserved
	this.ready = ready
	this.delayed = delayed
	this.deps = make([]string, 0)
	if deps, found := Config.deps[name]; found {
		this.deps = deps[:]
	}
	return this
}

func (this *Tube) Jobs() uint {
	return this.delayed + this.ready + this.reserved
}

func (this *Tube) Ready(queue map[string]*Tube) bool {
	for _, dep := range this.deps {
		if tube, found := queue[dep]; found {
			if tube.Jobs() > 0 {
				return false
			}
		}
	}
	return true
}

func (this *Tube) MarshalJSON() ([]byte, error) {
	stats := make(map[string]interface{})
	stats["jobs-ready"] = this.ready
	stats["jobs-reserved"] = this.reserved
	stats["jobs-delayed"] = this.delayed
	return json.Marshal(stats)
}
