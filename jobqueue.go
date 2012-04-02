package main

import (
	"errors"
	"lentil"
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
	pri  int
	name string
}

type Tubes []*Tube

func (t Tubes) Len() int {
	return len(t)
}

func (t Tubes) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func (t Tubes) Less(i, j int) bool {
	return t[i].pri < t[j].pri
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
	_, e = this.q.Watch(this.tubes[0].name)
	if e != nil {
		return nil, e
	}
	// Timeout every 1 second to handle kill signals
	job, e := this.q.ReserveWithTimeout(1)
	if e != nil {
		return nil, e
	}
	_, e = this.q.Ignore(this.tubes[0].name)
	if e != nil {
		return nil, e
	}
	return job, nil
}

func (this *JobQueue) Delete(id uint64) error {
	return this.q.Delete(id)
}

func (this *JobQueue) refreshTubes() error {
	this.tubes = make([]*Tube, 0)
	tubes, e := this.q.ListTubes()
	if e != nil {
		return e
	}
	for _, tube := range tubes {
		if tube == "default" {
			continue
		}
		_, e := this.q.Watch(tube)
		if e != nil {
			return e
		}
		job, e := this.q.ReserveWithTimeout(0)
		if e != nil {
			if strings.Contains(e.Error(), "TIMED_OUT") {
				continue
			}
			return e
		}
		stats, e := this.q.StatsJob(job.Id)
		if e != nil {
			return e
		}
		priority, _ := strconv.Atoi(stats["pri"])
		delay, _ := strconv.Atoi(stats["delay"])
		e = this.q.Release(job.Id, priority, delay)
		if e != nil {
			return e
		}
		this.tubes = append(this.tubes, &Tube{priority, tube})
		_, e = this.q.Ignore(tube)
		if e != nil {
			return e
		}
	}
	sort.Sort(this.tubes)
	return nil
}
