package main

import (
	"encoding/json"
	"fmt"
	"github.com/nutrun/lentil"
	"testing"
)

func resetConfig() {
	Config = NewConfig("", "", "")
	Config.deps = make(map[string][]string)
}

func TestPriority(t *testing.T) {
	q := connect(t)
	resetConfig()
	Config.deps["tube1"] = []string{"tube2"}
	put(t, "job1", "tube1", 0, q)
	put(t, "job2", "tube2", 0, q)
	jobs := NewJobQueue(q, false, make([]string, 0))
	assertNextJob(t, jobs, "job2")
	assertNextJob(t, jobs, "job1")
}

func TestIncludeExclude(t *testing.T) {
	q := connect(t)
	resetConfig()
	all := NewJobQueue(q, false, make([]string, 0))
	if !all.Include("tube") {
		t.Errorf("Should include tube")
	}
	if !all.Include("another") {
		t.Errorf("Should include another")
	}
	none := NewJobQueue(q, true, make([]string, 0))
	if none.Include("none") {
		t.Errorf("Should not include tube none")
	}
	include := NewJobQueue(q, true, []string{"in"})
	if !include.Include("in") {
		t.Errorf("Should include tube in")
	}
	if include.Include("out") {
		t.Errorf("Should not include tube out")
	}
	exclude := NewJobQueue(q, false, []string{"out"})
	if !exclude.Include("in") {
		t.Errorf("Should not include tube in")
	}
	if exclude.Include("out") {
		t.Errorf("Should not include tube out")
	}

}

func TestMoarPriorities(t *testing.T) {
	q := connect(t)
	resetConfig()
	Config.deps["tube3"] = []string{"tube2"}
	Config.deps["tube1"] = []string{"tube3"}
	put(t, "job11", "tube1", 0, q)
	put(t, "job21", "tube2", 0, q)
	put(t, "job31", "tube3", 0, q)
	put(t, "job22", "tube2", 0, q)
	put(t, "job32", "tube3", 0, q)
	put(t, "job12", "tube1", 0, q)
	jobs := NewJobQueue(q, false, make([]string, 0))
	assertNextJob(t, jobs, "job21")
	assertNextJob(t, jobs, "job22")
	assertNextJob(t, jobs, "job31")
	assertNextJob(t, jobs, "job32")
	assertNextJob(t, jobs, "job11")
	assertNextJob(t, jobs, "job12")
}

func TestSleepWhenNoJobs(t *testing.T) {
	q := connect(t)
	resetConfig()
	jobs := NewJobQueue(q, false, make([]string, 0))
	no_job, err := reserveNextJob(t, jobs, "job11")

	if no_job != nil {
		t.Error(fmt.Sprintf("Reserved %v when should not have", no_job))
	}
	if err == nil {
		t.Error(fmt.Sprintf("Should have thrown a TIME_OUT, threw  %v instead", err))
	}

}

func TestBlockOnReserved(t *testing.T) {
	q := connect(t)
	resetConfig()
	Config.deps["tube1"] = []string{"tube2"}
	put(t, "job1", "tube1", 0, q)
	put(t, "job2", "tube2", 0, q)
	jobs := NewJobQueue(q, false, make([]string, 0))
	job, err := reserveNextJob(t, jobs, "job2")
	if err != nil {
		t.Error(fmt.Sprintf("Could not reserve job %s", job))
	}
	no_job, err := reserveNextJob(t, jobs, "job1")
	if no_job != nil {
		t.Error(fmt.Sprintf("Reserved %v when should not have", no_job))
	}
	if err == nil {
		t.Error(fmt.Sprintf("Should have thrown a TIME_OUT, threw  %v instead", err))
	}

}

func TestBlockOnIgnored(t *testing.T) {
	q := connect(t)
	resetConfig()
	Config.deps["another"] = []string{"block_on"}
	put(t, "job", "block_on", 0, q)
	put(t, "another", "another", 0, q)
	jobs := NewJobQueue(q, false, []string{"block_on"})
	no_job, err := reserveNextJob(t, jobs, "job")
	if no_job != nil {
		t.Error(fmt.Sprintf("Reserved %v when should not have", no_job))
	}
	if err == nil {
		t.Error(fmt.Sprintf("Should have thrown a TIME_OUT, threw  %v instead", err))
	}

}

func assertNextJob(t *testing.T, jobqueue *JobQueue, expected string) {
	jobinfo := make(map[string]string)
	job, e := jobqueue.Next()
	if e != nil {
		t.Error(fmt.Sprintf("%v on [%v]", e, expected))
		return
	}
	json.Unmarshal(job.Body, &jobinfo)
	if jobinfo["name"] != expected {
		t.Errorf("%s != %s\n", expected, jobinfo["name"])
	}
	jobqueue.Delete(job.Id)
}

func reserveNextJob(t *testing.T, jobqueue *JobQueue, expected string) (*lentil.Job, error) {
	job, e := jobqueue.Next()
	if e != nil {
		return nil, e
	}
	return job, e
}

func put(t *testing.T, jobName, tube string, delay int, q *lentil.Beanstalkd) {
	job := make(map[string]string)
	job["tube"] = tube
	job["name"] = jobName
	jobjson, _ := json.Marshal(job)
	e := q.Use(tube)
	if e != nil {
		t.Fatal(e)
	}
	_, e = q.Put(0, delay, 60, jobjson)
	if e != nil {
		t.Error(e)
	}
}

func connect(t *testing.T) *lentil.Beanstalkd {
	q, e := lentil.Dial("0.0.0.0:11300")
	if e != nil {
		t.Fatal(e)
	}
	// Clear beanstalkd
	tubes, e := q.ListTubes()
	if e != nil {
		t.Fatal(e)
	}
	for _, tube := range tubes {
		if tube == "default" {
			continue
		}
		_, e = q.Watch(tube)
		if e != nil {
			t.Fatal(e)
		}
		for {
			job, e := q.ReserveWithTimeout(0)
			if e != nil {
				break
			}
			q.Delete(job.Id)
		}
		_, e := q.Ignore(tube)
		if e != nil {
			t.Fatal(e)
		}
	}
	return q
}
