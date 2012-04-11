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
	tubes Tubes
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

func (this Tubes) TrimMajor() Tubes {
	sort.Sort(this)
	for i := 0; i < this.Len(); i++ {
		if this[0].majorPriority != this[i].majorPriority {
			return this[0:i]

		}
	}
	return this
}

func (this Tubes) TrimMinor() Tubes {
	tubes := make(Tubes, 0)
	for i := 0; i < this.Len(); i++ {
		if this[i].jobcnt > 0 {
			if tubes.Len() > 0 {
				if tubes[0].minorPriority < this[i].minorPriority {
					return tubes
				}
			}
			tubes = append(tubes, this[i])
		}
	}
	return tubes
}

func NewJobQueue(q *lentil.Beanstalkd) *JobQueue {
	this := new(JobQueue)
	this.q = q
	return this
}

func (this *JobQueue) Next() (*lentil.Job, error) {
	e := this.refreshTubes()
	if e != nil {
		return nil, e
	}
	// Keep on timing out until there's jobs (we don't watch "default")
	if len(this.tubes) == 0 {
		time.Sleep(time.Second)
		return nil, errors.New("TIMED_OUT")
	}
	for _, tube := range this.tubes {
		_, e = this.q.Watch(tube.name)
		if e != nil {
			return nil, e
		}
		// Timeout every 1 second to handle kill signals
		job, e := this.q.ReserveWithTimeout(1)
		if e != nil {
			_, e = this.q.Ignore(tube.name)
			if e != nil {
				return nil, e
			}
			if tube.jobcnt > 0 {
				return nil, errors.New("TIMED_OUT")
			}
			continue
		}
		_, e = this.q.Ignore(tube.name)
		if e != nil {
			return nil, e
		}
		return job, nil
	}
	return nil, errors.New("TIMED_OUT")
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
	this.tubes = make(Tubes, 0)
	tubes, e := this.q.ListTubes()
	if e != nil {
		return e
	}
	for _, tube := range tubes {
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
			this.tubes = append(this.tubes, &Tube{major, minor, ready + reserved, tube})
		}
	}
	this.tubes = this.tubes.TrimMajor()
	this.tubes = this.tubes.TrimMinor()
	return nil
}
