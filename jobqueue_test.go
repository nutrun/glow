package main

import(
	"testing"
	"lentil"
	"encoding/json"
)

func TestDependencies(t *testing.T) {
	q, e := lentil.Dial("0.0.0.0:11300")
	if e != nil {
		t.Fatal(e)
	}
	// Clear queue
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
	jobqueue := NewJobQueue(q)
	jobinfo := make(map[string]string)
	jobinfo["tube"] = "rock"
	jobinfo["name"] = "1"
	jobjson, e := json.Marshal(jobinfo)
	if e != nil {
		t.Error(e)
	}
	e = q.Use("rock")
	if e != nil {
		t.Error(e)
	}
	q.Put(0, 0, 60, jobjson)
	job, e := jobqueue.Next()
	if e != nil {
		t.Error(e)
	}
	println(job.Id)
}
