package main

import (
	"encoding/json"
	"github.com/nutrun/lentil"
	"strconv"
	"strings"
)

type JobQueue struct {
	q         *lentil.Beanstalkd
	tubes     map[string]*Tube
	inclusive bool
	filter    []string
}

func NewJobQueue(q *lentil.Beanstalkd, inclusive bool, filter []string) *JobQueue {
	this := new(JobQueue)
	this.q = q
	this.inclusive = inclusive
	this.filter = filter
	return this
}

func (this *JobQueue) ReadyTubes() []*Tube {
	ready := make([]*Tube, 0)
	for _, tube := range this.tubes {
		if tube.Ready(this.tubes) {
			ready = append(ready, tube)
		}
	}
	return ready
}

func (this *JobQueue) ReserveFromTubes(tubes []*Tube) (*lentil.Job, error) {
	for _, tube := range tubes {
		if this.Include(tube.name) {
			_, e := this.q.Watch(tube.name)
			if e != nil {
				return nil, e
			}
		}
	}
	job, err := this.q.ReserveWithTimeout(0)
	for _, ignored := range tubes {
		if this.Include(ignored.name) {
			_, e := this.q.Ignore(ignored.name)
			if e != nil {
				return nil, e
			}
		}
	}
	return job, err
}

func (this *JobQueue) Next() (*lentil.Job, error) {
	this.refreshTubes()
	return this.ReserveFromTubes(this.ReadyTubes())
}

func (this *JobQueue) Delete(id uint64) error {
	return this.q.Delete(id)
}

func (this *JobQueue) MarshalJSON() ([]byte, error) {
	this.refreshTubes()
	return json.Marshal(this.tubes)
}

func (this *JobQueue) Include(tube string) bool {
	for _, filter := range this.filter {
		if strings.Contains(tube, filter) {
			return this.inclusive
		}
	}
	return !this.inclusive
}

func (this *JobQueue) refreshTubes() error {
	this.tubes = make(map[string]*Tube)
	names, e := this.q.ListTubes()
	if e != nil {
		return e
	}
	for _, tube := range names {
		if tube == "default" || tube == Config.errorQueue {
			continue
		}
		tubestats, e := this.q.StatsTube(tube)
		if e != nil {
			return e
		}
		ready, _ := strconv.Atoi(tubestats["current-jobs-ready"])
		reserved, _ := strconv.Atoi(tubestats["current-jobs-reserved"])
		delayed, _ := strconv.Atoi(tubestats["current-jobs-delayed"])
		this.tubes[tube] = NewTube(tube, uint(reserved), uint(ready), uint(delayed))
	}
	return nil
}
