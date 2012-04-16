package main

import (
	"encoding/json"
	"fmt"
	"github.com/nutrun/lentil"
	"testing"
)

func TestMajor(t *testing.T) {
	tubes := make(Tubes, 3)
	tubes[0] = &Tube{1, 2, 10, "one"}
	tubes[1] = &Tube{0, 2, 10, "two"}
	tubes[2] = &Tube{0, 2, 10, "three"}
	tubes = tubes.FirstMajor()
	if tubes.Len() != 2 {
		t.Errorf("Trim failed to remove tubes, the number of tubes is [%v]", tubes.Len())
	}
}

func TestMinor(t *testing.T) {
	tubes := make(Tubes, 3)
	tubes[0] = &Tube{1, 2, 10, "one"}
	tubes[1] = &Tube{0, 3, 10, "two"}
	tubes[2] = &Tube{0, 2, 10, "three"}
	tubes = tubes.FirstMinor()
	if tubes.Len() != 1 {
		t.Errorf("Trim failed to remove tubes, the number of tubes is [%v]", tubes.Len())
	}
}

func TestPriority(t *testing.T) {
	q := connect(t)
	put(t, "job1", "tube1", 2, 0, 0, q)
	put(t, "job2", "tube2", 0, 0, 0, q)
	jobs := NewJobQueue(q)
	assertNextJob(t, jobs, "job2")
	assertNextJob(t, jobs, "job1")
}

func TestMoarPriorities(t *testing.T) {
	q := connect(t)
	put(t, "job11", "tube1", 3, 0, 0, q)
	put(t, "job21", "tube2", 1, 0, 0, q)
	put(t, "job31", "tube3", 2, 0, 0, q)
	put(t, "job22", "tube2", 1, 0, 0, q)
	put(t, "job32", "tube3", 2, 0, 0, q)
	put(t, "job12", "tube1", 3, 0, 0, q)
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
	put(t, "job11", "tube1", 0, 1, 0, q)
	put(t, "job21", "tube2", 0, 0, 0, q)
	put(t, "job22", "tube2", 0, 0, 0, q)
	put(t, "job12", "tube1", 0, 1, 0, q)
	jobs := NewJobQueue(q)
	assertNextJob(t, jobs, "job21")
	assertNextJob(t, jobs, "job22")
	assertNextJob(t, jobs, "job11")
	assertNextJob(t, jobs, "job12")
}

func TestSleepWhenNoJobs(t *testing.T) {
	q := connect(t)
	jobs := NewJobQueue(q)
	no_job, err := reserveNextJob(t, jobs, "job11")
	if no_job != nil {
		t.Error(fmt.Sprintf("Reserved %v when should not have", no_job))
	}
	if err == nil {
		t.Error(fmt.Sprintf("Should have thrown a TIME_OUT, threw  %v instead", err))
	}

}

func TestTwoMinorsFromDiffQueues(t *testing.T) {
	q := connect(t)
	put(t, "job1", "tube1", 0, 1, 0, q)
	put(t, "job2", "tube2", 0, 1, 0, q)
	put(t, "job3", "tube3", 0, 1, 0, q)
	put(t, "job4", "tube4", 0, 1, 0, q)
	jobs := NewJobQueue(q)
	job1, err := reserveNextJob(t, jobs, "job1")
	if err != nil {
		t.Error(fmt.Sprintf("Could not resere job1 [%v]", err))
	}
	job2, err := reserveNextJob(t, jobs, "job2")
	if err != nil {
		t.Error(fmt.Sprintf("Could not resere job2 [%v]", err))
	}
	job3, err := reserveNextJob(t, jobs, "job3")
	if err != nil {
		t.Error(fmt.Sprintf("Could not resere job3 [%v]", err))
	}
	job4, err := reserveNextJob(t, jobs, "job4")
	if err != nil {
		t.Error(fmt.Sprintf("Could not resere job4 [%v]", err))
	}
	jobs.Delete(job1.Id)
	jobs.Delete(job2.Id)
	jobs.Delete(job3.Id)
	jobs.Delete(job4.Id)
}

func TestMajoarWorkingPrioraties(t *testing.T) {
	q := connect(t)
	put(t, "job11", "tube1", 0, 1, 0, q)
	put(t, "job21", "tube2", 0, 0, 0, q)
	put(t, "job22", "tube2", 0, 0, 0, q)
	put(t, "job12", "tube1", 0, 1, 0, q)
	jobs := NewJobQueue(q)
	job21, err := reserveNextJob(t, jobs, "job21")
	if err != nil {
		t.Error(fmt.Sprintf("Could not resere job21 [%v]", err))
	}
	job22, err := reserveNextJob(t, jobs, "job22")
	if err != nil {
		t.Error(fmt.Sprintf("Could not resere job22 [%v]", err))
	}
	no_job, err := reserveNextJob(t, jobs, "job11")
	if no_job != nil {
		t.Error(fmt.Sprintf("Reserved %v when should not have", no_job))
	}
	if err == nil {
		t.Error(fmt.Sprintf("Should have thrown a TIME_OUT, threw  %v instead", err))
	}
	jobs.Delete(job21.Id)
	jobs.Delete(job22.Id)
	assertNextJob(t, jobs, "job11")
	assertNextJob(t, jobs, "job12")
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

func put(t *testing.T, jobName, tube string, major, minor uint, delay int, q *lentil.Beanstalkd) {
	job := make(map[string]string)
	job["tube"] = tube
	job["name"] = jobName
	jobjson, _ := json.Marshal(job)
	e := q.Use(fmt.Sprintf("%s_%d", tube, int((major<<16)|minor)))
	if e != nil {
		t.Fatal(e)
	}
	_, e = q.Put(int((major<<16)|(minor)), delay, 60, jobjson)
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
