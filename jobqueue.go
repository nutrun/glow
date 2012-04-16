package main

import (
	"errors"
	"github.com/nutrun/lentil"
	"sort"
	"strconv"
	"strings"
	"time"
)

type JobQueue struct {
	q     *lentil.Beanstalkd
	major Tubes
	minor Tubes
}

type Tube struct {
	majorPriority uint
	minorPriority uint
	jobcnt        int
	name          string
}

type Tubes []*Tube

func (t Tubes) Len() int {
	return len(t)
}

func (t Tubes) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func (this Tubes) Jobs() int {
	jobs := 0
	for _, tube := range this {
		jobs += tube.jobcnt
	}
	return jobs
}

// Sort tubes on ascending priority and descending job count
func (t Tubes) Less(i, j int) bool {
	if t[i].majorPriority != t[j].majorPriority {
		return t[i].majorPriority < t[j].majorPriority
	}
	if t[i].minorPriority == t[j].minorPriority {
		return t[i].jobcnt > t[j].jobcnt
	}
	return t[i].minorPriority < t[j].minorPriority
}

func (this Tubes) FirstMajor() Tubes {
	sort.Sort(this)
	for i := 0; i < this.Len(); i++ {
		if this[0].majorPriority != this[i].majorPriority || this[0].minorPriority != this[i].minorPriority {
			return this[0:i]
		}
	}
	return this
}

func (this Tubes) FirstMinor() Tubes {
	if this.Len() > 0 {
		index := 1
		for i := 0; i < this.Len(); i++ {
			if this[0].minorPriority != this[i].minorPriority {
				index = i
			}
			if this[0].majorPriority != this[i].majorPriority {
				return this[index-1 : i]
			}
		}
		return this[index-1:]
	}
	return this
}

func NewJobQueue(q *lentil.Beanstalkd) *JobQueue {
	this := new(JobQueue)
	this.q = q
	return this
}

func (this *JobQueue) ReserveFromTubes(tubes Tubes) (*lentil.Job, error) {
	watched := make(Tubes, 0)
	for _, tube := range tubes {
		_, e := this.q.Watch(tube.name)
		if e != nil {
			return nil, e
		}
		watched = append(watched, tube)
	}
	job, res_err := this.q.ReserveWithTimeout(1)
	for _, ignored := range watched {
		_, e := this.q.Ignore(ignored.name)
		if e != nil {
			return nil, e
		}
	}
	if res_err != nil {
		if tubes.Jobs() > 0 {
			return nil, res_err
		}
		return nil, nil
	}
	return job, nil
}

func (this *JobQueue) Next() (*lentil.Job, error) {
	e := this.refreshTubes()
	if e != nil {
		return nil, e
	}
	// Keep on timing out until there's jobs (we don't watch "default")
	if len(this.major) == 0 && len(this.minor) == 0 {
		time.Sleep(time.Second)
		return nil, errors.New("TIMED_OUT")
	}
	job, e := this.ReserveFromTubes(this.major)
	if e != nil {
		return nil, e
	}
	if job == nil {
		job, e = this.ReserveFromTubes(this.minor)
		if e != nil {
			return nil, e
		}
		if job != nil {
			return job, nil
		}
	}
	return job, nil
}

func (this *JobQueue) Delete(id uint64) error {
	return this.q.Delete(id)
}

func (this *JobQueue) priority(tube string) (uint, uint, error) {
	index := strings.LastIndex(tube, "_")
	priority, err := strconv.Atoi(tube[index+1:])
	if err != nil {
		return 0, 0, err
	}
	return uint(priority >> 16), uint(priority & 0x0000FFFF), nil
}

// TODO We shouldn't refresh tubes if the list hasn't changed
func (this *JobQueue) refreshTubes() error {
	tubes := make(Tubes, 0)
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
			tubes = append(tubes, &Tube{major, minor, ready + reserved, tube})
		}
	}
	this.major = tubes.FirstMajor()
	this.minor = tubes.FirstMinor()
	return nil
}
