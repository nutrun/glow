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

	jobWithNoDeps := make(map[string]string)
	jobWithNoDeps["tube"] = "rock"
	jobWithNoDeps["name"] = "2"
	jobWithNoDepsJson, _ := json.Marshal(jobWithNoDeps)

	jobWithDep := make(map[string]string)
	jobWithDep["tube"] = "metal"
	jobWithDep["name"] = "1"
	jobWithDep["depends"] = "rock"
	jobWithDepJson, _ := json.Marshal(jobWithDep)

	q.Use("metal")
	q.Put(0, 0, 60, jobWithDepJson)
	q.Use("rock")
	q.Put(0, 0, 60, jobWithNoDepsJson)

	job, e := jobqueue.Next()
	if e != nil {
		t.Fatal(e)
	}
	jobinfo := make(map[string]string)
	json.Unmarshal(job.Body, &jobinfo)
	if jobinfo["name"] != "2" {
		t.Error(jobinfo["name"])
	}
}
