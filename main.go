package main

import (
	"flag"
	"log"
	"os"
	"strings"
)

var listener *bool = flag.Bool("listen", false, "Start listener")
var help *bool = flag.Bool("help", false, "Show help")
var mailto *string = flag.String("mailto", "", "Who to email on failure (comma separated)")
var workdir *string = flag.String("workdir", ".", "Directory to run job from")
var out *string = flag.String("out", "/dev/null", "File to send job's stdout and stderr")
var tube *string = flag.String("tube", "default", "Beanstalkd tube to send the job to")
var stats *bool = flag.Bool("stats", false, "Show queue stats")
var verbose *bool = flag.Bool("v", false, "Increase verbosity")
var depends *string = flag.String("depends", "", "List of tubes job depends on (comma separated)")

func main() {
	flag.Parse()

	if *listener {
		l, e := NewListener(*verbose)
		if e != nil {
			log.Fatalf("ERROR: %s", e.Error())
		}
		l.run()
	} else if *stats {
		c, e := NewClient(*verbose)
		if e != nil {
			log.Fatalf("ERROR: %s", e.Error())
		}
		e = c.stats()
		if e != nil {
			log.Fatalf("ERROR: %s", e.Error())
		}
	} else if *help || len(flag.Args()) == 0 {
		flag.Usage()
		os.Exit(1)
	} else {
		c, e := NewClient(*verbose)
		if e != nil {
			log.Fatalf("ERROR: %s", e.Error())
		}
		cmd := strings.Join(flag.Args(), " ")
		e = c.put(cmd, *mailto, *workdir, *out, *tube, *depends)
		if e != nil {
			log.Fatalf("ERROR: %s", e.Error())
		}
		os.Exit(0)
	}
}
