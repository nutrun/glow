package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

var listener *bool = flag.Bool("listen", false, "Start listener")
var help *bool = flag.Bool("help", false, "Show help")
var mailto *string = flag.String("mailto", "", "Who to email on failure (comma separated) [submit]")
var workdir *string = flag.String("workdir", "/tmp", "Directory to run job from [submit]")
var stdout *string = flag.String("stdout", "/dev/null", "File to send job's stdout [submit]")
var stderr *string = flag.String("stderr", "/dev/null", "File to send job's stderr [submit]")
var tube *string = flag.String("tube", "", "Beanstalkd tube to send the job to [submit]")
var stats *bool = flag.Bool("stats", false, "Show queue stats")
var drain *string = flag.String("drain", "", "Empty tubes (comma separated)")
var verbose *bool = flag.Bool("v", false, "Increase verbosity")
var exclude *string = flag.String("exclude", "", "Tubes to ignore (comma separated) [listen]")
var priority *int = flag.Int("pri", 0, "Job priority (smaller runs first) [submit]")
var delay *int = flag.Int("delay", 0, "Job delay in seconds [submit]")
var local *bool = flag.Bool("local", false, "Run locally, reporting errors to the configured beanstalk")
var pause *string = flag.String("pause", "", "Pause tubes (comma separated)")
var pausedelay *int = flag.Int("pause-delay", 0, "How many seconds to pause tubes for")
var mailfrom *string = flag.String("mail-from", "glow@example.com", "Email 'from' field [listen]")
var smtpserver *string = flag.String("smtp-server", "", "Server to use for sending emails [listen]")
var deps *string = flag.String("deps", "", "Path to tube dependency config file [listen]")
var logpath *string = flag.String("log", "/dev/stderr", "Path to log file [listen]")
var mlen *int = flag.Int("mlen", 65536, "Max length of messeges sent to beanstalk [specified in bytes]")
var Config *Configuration

func main() {
	flag.Parse()
	Config = NewConfig(*deps, *smtpserver, *mailfrom)
	if *listener {
		include := false
		filter := make([]string, 0)
		if *exclude != "" {
			filter = strings.Split(*exclude, ",")
		}
		l, e := NewListener(*verbose, include, filter, *logpath, *mlen)
		if e != nil {
			fmt.Fprintln(os.Stderr, e)
			os.Exit(1)
		}
		l.run()
		return
	} else if *help {
		flag.Usage()
		os.Exit(1)
	}

	if *local {
		executable, arguments := parseCommand()
		// hack: local doesn't need tube, defaulting it to respect the Message API
		msg, e := NewMessage(executable, arguments, *mailto, *workdir, *stdout, *stderr, "localignore", 0, 0)
		if e != nil {
			fmt.Fprintln(os.Stderr, e)
			os.Exit(1)
		}
		logfile, e := os.OpenFile(*logpath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if e != nil {
			fmt.Fprintln(os.Stderr, e)
			os.Exit(1)
		}
		runner, e := NewRunner(*verbose, log.New(logfile, "", log.LstdFlags))
		if e != nil {
			fmt.Fprintln(os.Stderr, e)
			os.Exit(1)
		}
		e = runner.execute(msg)
		if e != nil {
			fmt.Fprintln(os.Stderr, e)
			os.Exit(1)
		}
		return
	}

	c, e := NewClient(*verbose)
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}

	if *drain != "" {
		e = c.drain(*drain)
		if e != nil {
			fmt.Fprintln(os.Stderr, e)
			os.Exit(1)
		}
	} else if *pause != "" {
		if *pausedelay == 0 {
			fmt.Fprintln(os.Stderr, "Usage: glow -pause=<tube1,tube2,...> -pause-delay=<seconds>")
			os.Exit(1)
		}
		e = c.pause(*pause, *pausedelay)
	} else if *stats {
		e = c.stats()
		if e != nil {
			fmt.Fprintln(os.Stderr, e)
			os.Exit(1)
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
				fmt.Fprintln(os.Stderr, e)
				os.Exit(1)
			}
			input = append(input, line...)
		}
		e = c.putMany(input)
		if e != nil {
			fmt.Fprintln(os.Stderr, e)
			os.Exit(1)
		}
	} else { // Queue up one job
		executable, arguments := parseCommand()
		msg, e := NewMessage(executable, arguments, *mailto, *workdir, *stdout, *stderr, *tube, *priority, *delay)
		if e != nil {
			fmt.Fprintln(os.Stderr, e)
			os.Exit(1)
		}
		e = c.put(msg)
		if e != nil {
			fmt.Fprintln(os.Stderr, e)
			os.Exit(1)
		}
	}
}

func parseCommand() (string, []string) {
	return flag.Args()[0], flag.Args()[1:len(flag.Args())]
}
