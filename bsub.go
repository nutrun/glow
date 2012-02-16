package main

import "beanstalk"
import "strings"
import "os"
import "fmt"
import "beanie"

var partition *string = flag.String("p", "", "Partition")
var workgroup *string = flag.String("g", "", "Name of work group for this job")
var begin     *string = flag.String("b", "", "Begin time")
var wait      *string = flag.String("w", "", "Wait on all jobs in the work group specified")
var esttime   *string = flag.String("t", "", "Estimated running time of this job")

func main() {
    conn, err := beanstalk.Dial("localhost:11300")
    if err != nil {
        panic(err.String())
    }

    btube, err := beanstalk.NewTube(conn, bean.GetTubeName(bean.GetUser(), "all"))
    id, err := btube.Put(strings.Join(os.Args[1:], " "), 0, 0, 10)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Job submission failed: %s\n", err.String())
    }
    fmt.Printf("Job submitted successfully. Beanie Id: %d\n", id)
}
