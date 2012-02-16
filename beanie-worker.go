package main

import "beanstalk"
import "fmt"
import "flag"
import "exec"
import "time"
import "strconv"
import "strings"
import "beanie"

var queue *string = flag.String("q", "", "host:port for Beanstalk queue")
var slurmpartition *string = flag.String("p", "", "Slurm partition")
var workerrank *uint = flag.Uint("r", 0, "Worker rank")
var beantube *string = flag.String("t", "", "Tube")

func slurm_start_worker(user, partition, queue string, tube string, rank uint) {
    
    e := exec.Command("bean-worker", "-t", tube, "-p", partition, "-q", queue, "-r", fmt.Sprintf("%d", rank+1)).Start()

//    e := exec.Command("sbatch", "-p", partition, "--wrap", fmt.Sprintf("bean-worker -t %s -p %s -q %s -r %d", tube, partition, queue, rank+1)).Run()
    if e != nil {
        panic(e.String())
    }
}

func monitor_tube(t *beanstalk.Tube, user, partition, queue string, rank uint, tube string, done chan int) {
   
    for true {
        // sleep for 10 seconds
        e := time.Sleep(10000000000)
        if e != nil {
            panic(e.String())
        }

        stats, err := t.Stats()
        if err != nil {
            panic("Failed to get stats for tube!")
        }

        count, err := strconv.Atoi(stats["current-jobs-ready"])
        if err != nil {
            panic("Failed to parse current-jobs-ready stat as int")
        }

/*        workers, err := strconv.Atoi(stats["current-jobs-reserved"])
        if err != nil {
            panic("Failed to parse current-jobs-reserved stat as int")
        }
*/
        if count > 1 {
            slurm_start_worker(user, partition, queue, tube, rank)
        }

        // break if signalled by main
        select {
        case <- done:
            break
        default:
            // do nothing
        }
    }
}

func main() {

    flag.Parse()

    var p string
    if slurmpartition == nil || *slurmpartition == "" {
        p = "medium"
    } else {
        p = *slurmpartition
    }

    var rank uint
    if workerrank == nil {
        rank = 0
    } else {
        rank = *workerrank
    }
 
    tube, err := beanie.GetQueue(*queue, *beantube)
    if err != nil {
        panic(err.String())
    }
    
    done := make(chan int)
    if rank == 0 {
        go monitor_tube(tube p, q, rank, tube, done)
    }

    tubeset, err := beanstalk.NewTubeSet(conn, []string{bean.GetTubeName(user, tube)})
    if err != nil {
        panic("Failed to create tubeset")
    }

    for {
        j, err := tubeset.ReserveWithTimeout(10000000)

        // no jobs in queue - shutdown
        if err == beanstalk.TimedOut {
            break
        }

        if err != nil {
            panic("failed to reserve job from tubeset!")
        }

        fmt.Printf("found job: %s.  Executing\n", j.Body)
        cmdargs := strings.Split(j.Body, " ")
        cmd := exec.Command(cmdargs[0], cmdargs[1:]...)
        err = cmd.Run()
        fmt.Printf("done\n")
        
        j.Delete()
    }

    if rank == 0 {
        done <- 1
    }
}
