package main

import (
	"encoding/json"
	"lentil"
	"testing"
)

func TestDependencies(t *testing.T) {
	q := connect(t)
	put("job2", "tube2", []string{"tube1"}, q)
	put("job1", "tube1", make([]string, 0), q)
	jobqueue := NewJobQueue(q)
	assertNextJob(t, jobqueue, "job1")
}

func TestMoarDependencies(t *testing.T) {
	q := connect(t)
	nodeps := make([]string, 0)
	put("job11", "tube1", []string{"tube2", "tube3"}, q)
	put("job12", "tube1", []string{"tube2", "tube3"}, q)
	put("job21", "tube2", nodeps, q)
	put("job22", "tube2", nodeps, q)
	put("job31", "tube3", []string{"tube2"}, q)
	put("job32", "tube3", []string{"tube2"}, q)
	jobqueue := NewJobQueue(q)
	assertNextJob(t, jobqueue, "job21")
	assertNextJob(t, jobqueue, "job22")
	assertNextJob(t, jobqueue, "job31")	
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
	jobqueue.q.Delete(job.Id)
}

func put(jobName, tube string, depends []string, q *lentil.Beanstalkd) error {
	job := make(map[string]string)
	job["tube"] = tube
	job["name"] = jobName
	for _, dependency := range depends {
		job["depends"] = dependency
	}
	jobjson, _ := json.Marshal(job)
	e := q.Use(tube)
	if e != nil {
		return e
	}
	_, e = q.Put(0, 0, 60, jobjson)
	return e
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
