package main

import (
	"flag"
	"log"
	"os"
	"strings"
)

const beanstalkaddr = "0.0.0.0:11300"

var listener *bool = flag.Bool("listen", false, "Start listener")
var help *bool = flag.Bool("help", false, "Show help")
var mailto *string = flag.String("mailto", "", "Who to email on failure")

func main() {
	flag.Parse()

	if *listener {
		l, e := NewListener(beanstalkaddr)
		if e != nil {
			log.Fatalln(e.Error())
		}
		l.run()
	} else if *help || len(flag.Args()) == 0 {
		flag.Usage()
		os.Exit(1)
	} else {
		c, e := NewClient(beanstalkaddr)
		if e != nil {
			log.Fatalln(e.Error())
		}
		cmd := strings.Join(flag.Args(), " ")
		e = c.put(cmd, *mailto)
		if e != nil {
			log.Fatalln(e.Error())
		}
		os.Exit(0)
	}
}
