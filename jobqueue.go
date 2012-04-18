package main

import (
	"errors"
	"github.com/nutrun/lentil"
	"sort"
	"strconv"
	"strings"
)

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

func (this *Tubes) Groups() []*Group {
	out := make([]*Group, 0)
	keys := this.SortMapKeys(this.tubes)
	for _, key := range keys {
		out = append(out, this.tubes[key])
	}
	return out
}

type JobQueue struct {
	q         *lentil.Beanstalkd
	tubes     map[int]*Tubes
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

func (this *JobQueue) SortMapKeys(in map[int]*Tubes) []int {
	keys := make([]int, 0)
	for k, _ := range in {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	return keys
}

func (this *JobQueue) Tubes() []*Tubes {
	out := make([]*Tubes, 0)
	keys := this.SortMapKeys(this.tubes)
	for _, key := range keys {
		out = append(out, this.tubes[key])
	}
	return out
}

func (this *JobQueue) ReserveFromGroup(group *Group) (*lentil.Job, error) {
	for _, tube := range group.tubes {
		if this.Include(tube.name) {
			_, e := this.q.Watch(tube.name)
			if e != nil {
				return nil, e
			}
		}
	}
	job, res_err := this.q.ReserveWithTimeout(0)
	for _, ignored := range group.tubes {
		if this.Include(ignored.name) {
			_, e := this.q.Ignore(ignored.name)
			if e != nil {
				return nil, e
			}
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
	for _, tube := range this.Tubes() {
		for _, group := range tube.Groups() {
			job, err := this.ReserveFromGroup(group)
			if err != nil {
				return nil, err
			}
			if job != nil {
				return job, nil
			}
		}
		return nil, errors.New("TIMED_OUT")
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

func (this *JobQueue) Stats(f func(tubes []*Tube)) {
	this.refreshTubes()
	for _, tube := range this.Tubes() {
		for _, group := range tube.Groups() {
			f(group.tubes)
		}
	}
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
