package main

import (
	"errors"
	"github.com/nutrun/lentil"
	"sort"
	"strconv"
	"strings"
)

type JobQueue struct {
	q     *lentil.Beanstalkd
	tubes map[int]*Tubes
}

type Tube struct {
	majorPriority int
	minorPriority int
	jobs          int
	name          string
}

type Group struct {
	tubes []*Tube
}

func (this *Group) AddTube(tube *Tube) {
	this.tubes = append(this.tubes, tube)
}

func (this *Group) Jobs() int {
	jobs := 0
	for _, job := range this.tubes {
		jobs += job.jobs
	}
	return jobs
}

type Tubes struct {
	tubes map[int]*Group
}

func (this *Tubes) AddTube(tube *Tube) {
	if _, found := this.tubes[tube.minorPriority]; !found {
		this.tubes[tube.minorPriority] = &Group{make([]*Tube, 0)}
	}
	this.tubes[tube.minorPriority].AddTube(tube)
}

func (this *Tubes) SortMapKeys(in map[int]*Group) []int {
	keys := make([]int, 0)
	for k, _ := range in {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	return keys
}

func (this *Tubes) Groups() chan *Group {
	in := make(chan *Group)
	keys := this.SortMapKeys(this.tubes)
	go func() {
		for _, key := range keys {
			in <- this.tubes[key]
		}
		close(in)
	}()
	return in
}

func NewJobQueue(q *lentil.Beanstalkd) *JobQueue {
	this := new(JobQueue)
	this.q = q
	return this
}

func (this *JobQueue) SortMapKeys(in map[int]*Tubes) []int {
	keys := make([]int, 0)
	for k, _ := range in {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	return keys
}

func (this *JobQueue) Tubes() chan *Tubes {
	in := make(chan *Tubes)
	keys := this.SortMapKeys(this.tubes)
	go func() {
		for _, key := range keys {
			in <- this.tubes[key]
		}
		close(in)
	}()
	return in
}

func (this *JobQueue) ReserveFromGroup(group *Group) (*lentil.Job, error) {
	for _, tube := range group.tubes {
		_, e := this.q.Watch(tube.name)
		if e != nil {
			return nil, e
		}
	}
	job, res_err := this.q.ReserveWithTimeout(1)
	for _, ignored := range group.tubes {
		_, e := this.q.Ignore(ignored.name)
		if e != nil {
			return nil, e
		}
	}
	if res_err != nil {
		if group.Jobs() > 0 {
			return nil, res_err
		}
		return nil, nil
	}
	return job, nil
}

func (this *JobQueue) Next() (*lentil.Job, error) {
	this.refreshTubes()
	for tube := range this.Tubes() {
		for group := range tube.Groups() {
			job, err := this.ReserveFromGroup(group)
			if err != nil {
				return nil, err
			}
			if job != nil {
				return job, nil
			}
		}
	}
	return nil, errors.New("TIMED_OUT")
}

func (this *JobQueue) Delete(id uint64) error {
	return this.q.Delete(id)
}

func (this *JobQueue) priority(tube string) (int, int, error) {
	index := strings.LastIndex(tube, "_")
	priority, err := strconv.Atoi(tube[index+1:])
	if err != nil {
		return 0, 0, err
	}
	return int(priority >> 16), int(priority & 0x0000FFFF), nil
}

func (this *JobQueue) refreshTubes() error {
	this.tubes = make(map[int]*Tubes)
	names, e := this.q.ListTubes()
	if e != nil {
		return e
	}
	for _, tube := range names {
		if tube == "default" {
			continue
		}
		major, minor, e := this.priority(tube)
		if e != nil {
			continue
		}
		tubestats, e := this.q.StatsTube(tube)
		if e != nil {
			return e
		}
		ready, _ := strconv.Atoi(tubestats["current-jobs-ready"])
		reserved, _ := strconv.Atoi(tubestats["current-jobs-reserved"])
		delayed, _ := strconv.Atoi(tubestats["current-jobs-delayed"])
		if ready+reserved+delayed > 0 {
			if _, found := this.tubes[major]; !found {
				this.tubes[major] = &Tubes{make(map[int]*Group)}
			}
			this.tubes[major].AddTube(&Tube{major, minor, ready + reserved, tube})
		}
	}
	return nil
}
