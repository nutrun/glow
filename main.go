package main

import (
	"bufio"
	"flag"
	"log"
	"os"
	"strings"
)

var listener *bool = flag.Bool("listen", false, "Start listener")
var help *bool = flag.Bool("help", false, "Show help")
var mailto *string = flag.String("mailto", "", "Who to email on failure (comma separated) [sumbit]")
var workdir *string = flag.String("workdir", "/tmp", "Directory to run job from [submit]")
var out *string = flag.String("out", "/dev/null", "File to send job's stdout and stderr [submit]")
var tube *string = flag.String("tube", "", "Beanstalkd tube to send the job to [submit]")
var stats *bool = flag.Bool("stats", false, "Show queue stats")
var drain *string = flag.String("drain", "", "Empty tubes (comma separated)")
var verbose *bool = flag.Bool("v", false, "Increase verbosity")
var exclude *string = flag.String("exclude", "", "Tubes to ignore (comma separated) [listen]")
var priority *int = flag.Int("pri", 0, "Job priority (smaller runs first) [submit]")
var delay *int = flag.Int("delay", 0, "Job delay in seconds [submit]")
var local *bool = flag.Bool("local", false, "Run locally, reporting errors to the configured beanstalk")

func main() {
	log.SetFlags(0)
	flag.Parse()
	if *listener {
		include := false
		filter := make([]string, 0)
		if *exclude != "" {
			filter = strings.Split(*exclude, ",")
		}
		l, e := NewListener(*verbose, include, filter)
		if e != nil {
			log.Fatalf("ERROR: %s", e.Error())
		}
		l.run()
		return
	} else if *help {
		flag.Usage()
		os.Exit(1)
	}

	if *local {
        executable, arguments := parse_command()
        msg := &Message{executable, arguments, *mailto, *workdir, *out, "", 0, 0}
		runner, e := NewRunner()
		if e != nil {
			log.Fatal(e)
		}
		e = runner.execute(msg)
		if e != nil {
			log.Fatalf("ERROR: %s", e.Error())
		}
		return
	}

	c, e := NewClient(*verbose)
	if e != nil {
		log.Fatalf("ERROR: %s", e.Error())
	}

	if *drain != "" {
		e = c.drain(*drain)
		if e != nil {
			log.Fatalf("ERROR: %s", e.Error())
		}
	} else if *stats {
		e = c.stats()
		if e != nil {
			log.Fatalf("ERROR: %s", e.Error())
		}
	} else if len(flag.Args()) == 0 { // Queue up many jobs from STDIN
		in := bufio.NewReaderSize(os.Stdin, 1024*1024)
		input := make([]byte, 0)
		for {
			line, e := in.ReadSlice('\n')
			if e != nil {
				if e.Error() == "EOF" {
					break
				}
				log.Fatalf("ERROR: %s", e.Error())
			}
			input = append(input, line...)
		}
		e = c.putMany(input)
		if e != nil {
			log.Fatalf("ERROR: %s", e.Error())
		}
	} else { // Queue up one job
        executable, arguments := parse_command()
		e = c.put(&Message{executable, arguments, *mailto, *workdir, *out, *tube, *priority, *delay})
		if e != nil {
			log.Fatalf("ERROR: %s", e.Error())
		}
	}
}

func parse_command() (string, []string) {
    return flag.Args()[0], flag.Args()[1:len(flag.Args())]
}
