package main

import (
	"encoding/json"
	"github.com/nutrun/lentil"
	"testing"
)

func TestTrim(t *testing.T) {
	tubes := make(Tubes, 3)
	tubes[0] = &Tube{1, 2, 10, "one"}
	tubes[1] = &Tube{0, 3, 10, "two"}
	tubes[2] = &Tube{0, 2, 10, "three"}
	tubes = tubes.TrimMajor()
	if tubes.Len() != 2 {
		t.Errorf("Trim failed to remove tubes, the number of tubes is [%v]", tubes.Len())
	}
}

func TestPriority(t *testing.T) {
	q := connect(t)
	put(t, "job1", "tube1", 2, 0, q)
	put(t, "job2", "tube2", 0, 0, q)
	jobs := NewJobQueue(q)
	assertNextJob(t, jobs, "job2")
}

func TestMoarPriorities(t *testing.T) {
	q := connect(t)
	put(t, "job11", "tube1", 3, 0, q)
	put(t, "job21", "tube2", 1, 0, q)
	put(t, "job31", "tube3", 2, 0, q)
	put(t, "job22", "tube2", 1, 0, q)
	put(t, "job32", "tube3", 2, 0, q)
	put(t, "job12", "tube1", 3, 0, q)
	jobs := NewJobQueue(q)
	assertNextJob(t, jobs, "job21")
	assertNextJob(t, jobs, "job22")
	assertNextJob(t, jobs, "job31")
	assertNextJob(t, jobs, "job32")
	assertNextJob(t, jobs, "job11")
	assertNextJob(t, jobs, "job12")
}

func TestMinorPrioraties(t *testing.T) {
	q := connect(t)
	put(t, "job11", "tube1", 0, 1, q)
	put(t, "job21", "tube2", 0, 0, q)
	put(t, "job22", "tube2", 0, 0, q)
	put(t, "job12", "tube1", 0, 1, q)
	jobs := NewJobQueue(q)
	assertNextJob(t, jobs, "job21")
	assertNextJob(t, jobs, "job22")
	assertNextJob(t, jobs, "job11")
	assertNextJob(t, jobs, "job12")
}

func TestSamePriorityDifferentJobCount(t *testing.T) {
	q := connect(t)
	put(t, "job11", "tube1", 0, 0, q)
	put(t, "job12", "tube1", 0, 0, q)
	put(t, "job13", "tube1", 0, 0, q)
	put(t, "job21", "tube2", 0, 0, q)
	put(t, "job22", "tube2", 0, 0, q)
	jobs := NewJobQueue(q)
	assertNextJob(t, jobs, "job11")
	assertNextJob(t, jobs, "job21")
	assertNextJob(t, jobs, "job12")
	assertNextJob(t, jobs, "job22")
	assertNextJob(t, jobs, "job13")
}

func assertNextJob(t *testing.T, jobqueue *JobQueue, expected string) {
	jobinfo := make(map[string]string)
	job, e := jobqueue.Next()
	if e != nil {
		t.Error(e)
		return
	}
	json.Unmarshal(job.Body, &jobinfo)
	if jobinfo["name"] != expected {
		t.Errorf("%s != %s\n", expected, jobinfo["name"])
	}
	jobqueue.Delete(job.Id)
}

func put(t *testing.T, jobName, tube string, major, minor uint, q *lentil.Beanstalkd) {
	job := make(map[string]string)
	job["tube"] = tube
	job["name"] = jobName
	jobjson, _ := json.Marshal(job)
	e := q.Use(tube)
	if e != nil {
		t.Fatal(e)
	}
	_, e = q.Put(int((major<<16)|(minor)), 0, 60, jobjson)
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
