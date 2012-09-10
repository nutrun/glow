UNDER CONSTRUCTION, THE DOCS MAY NOT BE COMPLETE

# glow

Distributed processing manager

## Setup 

- Install [beanstalkd](http://kr.github.com/beanstalkd/download.html)
- Download a glow binary and add it to `$PATH`

### Building from source

- Install [Go](http://golang.org/doc/install)
- Install the [lentil](https://github.com/nutrun/lentil) beanstalkd client library
- `go get github.com/nutrun/lentil` or `cd <path-to-glow-source> && go install` 

## Quickstart

Start beanstalkd:

```
$ beanstalkd
```

Start a glow listener:

```
$ glow -listen
```

Submit a job:

```
$ glow -tube=test -out=/dev/stdout ls
```

The job's output should appear on the terminal running the glow listener. Invoke `glow -h` to list all available options.


## Configuration

glow uses these environment variables:

- `GLOW_QUEUE`: beanstalkd queue to connect to, defaults to `0.0.0.0:11300`
- `GLOW_SMTP_SERVER`: server to use for sending emails [listener only]
- `GLOW_MAIL_FROM`: emails sent by glow will have this as the `from` field, defaults to `glow@example.com` [listener only]
- `GLOW_DEPS`: path to tube dependency configuration file (see below for details)

## Listen
A listener connects to the beanstalk queue the environment variable GLOW_QUEUE points to, listens for jobs and executes them.

```
$ GLOW_QUEUE=10.0.0.4:11300 glow -listen
```

Exclude

```
-exclude=<Tube,Tube> a listener will not reserve jobs from any of the specified tubes
```

### Submit
Required arguments

```
-tube= Tube (string)
```

Defaulted

```
-workdir= Workdir (string) default: /tmp
-out= StdOut/Stderr (string)  default: /dev/null
```

Optional arguments

```
-mailto= Mail error on Failure (string)
-pri=  Beanstalk Job Priority   (int)
-delay= Beanstalk Job Delay  (int)
```

### Tubes 
A beanstalk tube is a priority based fifo queue of jobs:
https://github.com/kr/beanstalkd/blob/master/doc/protocol.txt#L105

Dependency configuration

```
{
    "foo": [
        "bar"
    ],
    "baz": [
    	"foo",
    	"bar"
    ]
}
```

- Tube foo depends on tube bar: if there are any ready/delayed/working jobs in bar, no jobs from foo will run
- Tube bar does not have any dependencies, jobs from bar will run whenever there are free workers available in the priority or queueed order
- Tube baz depends on tube bar and foo, it will block until bar and foo are done

Priorities

```
-pri= is an integer < 2**32. Jobs with smaller priority values will be scheduled before jobs with larger priorities. 
```

### Signals
Kill listener immediatly

```
$ killall glow 
```

Shut down gracefully (wait for job to finish)

```
$ killall -2 glow
```

### Supermegamicrooptimization
For improved queueing performance, a json list of jobs can be piped to glow's stdin: 

```
echo '[{"cmd":"ls","pri":"0","tube":"foo","delay":"0","mailto":"example@example.com","out":"/tmp/glow.out","workdir":"/tmp/glow"},{"cmd":"ps","pri":"1","tube":"bar","delay":"0","mailto":"example@example.com","out":"/tmp/glow.out","workdir":"/tmp/glow"}]' | glow
```
