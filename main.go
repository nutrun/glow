package main

import (
	"flag"
	"log"
	"os"
	"strings"
)

var listener *bool = flag.Bool("listen", false, "Start listener")
var help *bool = flag.Bool("help", false, "Show help")
var mailto *string = flag.String("mailto", "", "Who to email on failure")
var workdir *string = flag.String("workdir", ".", "Directory to run job from")
var out *string = flag.String("out", "/dev/null", "File to send job's stdout and stderr")

func main() {
	flag.Parse()

	if *listener {
		l, e := NewListener()
		if e != nil {
			log.Fatalf("ERROR: %s", e.Error())
		}
		l.run()
	} else if *help || len(flag.Args()) == 0 {
		flag.Usage()
		os.Exit(1)
	} else {
		c, e := NewClient()
		if e != nil {
			log.Fatalf("ERROR: %s", e.Error())
		}
		cmd := strings.Join(flag.Args(), " ")
		e = c.put(cmd, *mailto, *workdir, *out)
		if e != nil {
			log.Fatalf("ERROR: %s", e.Error())
		}
		os.Exit(0)
	}
}
