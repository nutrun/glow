package main

import (
	"encoding/json"
	"lentil"
	"strings"
)

type JobQueue struct {
	q       *lentil.Beanstalkd
	tubes   map[string]*Tube
	watched string
}

type Tube struct {
	depends []string
}

func NewJobQueue(q *lentil.Beanstalkd) *JobQueue {
	this := new(JobQueue)
	this.q = q
	return this
}

func NewTube(depends string) *Tube {
	this := new(Tube)
	this.depends = make([]string, 0)
	for _, dependency := range strings.Split(depends, ",") {
		this.depends = append(this.depends, dependency)
	}
	return this
}

func (this *JobQueue) Next() (*lentil.Job, error) {
	this.refreshTubes()
	// Ignore watched tube to allow it to get deleted when empty
	if this.watched != "" {
		_, e := this.q.Ignore(this.watched)
		if e != nil {
			return nil, e
		}
	}
	// Ignore watched tubes
	for key, tube := range this.tubes {
		for _, dependency := range tube.depends {
			_, exists := this.tubes[dependency]
			if exists {
				break
			}
		}
		// Tube doesn't have any active deps, grab a job from it
		_, e := this.q.Watch(key)
		if e != nil {
			return nil, e
		}
		this.watched = key
		return this.q.ReserveWithTimeout(1)
	}
	panic("should never get here")
}

func (this *JobQueue) refreshTubes() error {
	tubes, e := this.q.ListTubes()
	if e != nil {
		return e
	}
	// Only gather tube info if list of tubes has changed since last time we checked
	skipRefresh := true
	if this.tubes == nil || len(this.tubes) == len(tubes) {
		for _, tube := range tubes {
			
			_, exists := this.tubes[tube]
			if !exists {
				skipRefresh = false
				break
			}
		}
	}
	if skipRefresh {
		return nil
	}
	this.tubes = make(map[string]*Tube)
	jobinfo := make(map[string]string)
	for _, tubeName := range tubes {
		if tubeName == "default" {
			continue
		}
		_, e := this.q.Watch(tubeName)
		if e != nil {
			return e
		}
		job, e := this.q.PeekReady()
		if e != nil {
			return e
		}
		e = json.Unmarshal(job.Body, &jobinfo)
		if e != nil {
			return e
		}
		tube := NewTube(jobinfo["depends"])
		this.tubes[tubeName] = tube
		_, e = this.q.Ignore(tubeName)
		if e != nil {
			return e
		}
	}
	return nil
}
