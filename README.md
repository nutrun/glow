UNDER CONSTRUCTION

# glow

Distributed processing manager

## Setup 

- Install [beanstalkd](http://kr.github.com/beanstalkd/download.html)
- Download a glow binary and add it to `$PATH`

### Building from source

- Install [Go](http://golang.org/doc/install)
- Install the [lentil](https://github.com/nutrun/lentil) beanstalkd client library
- `go get github.com/nutrun/glow` or `cd <path-to-glow-source> && go install` 

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

## Listen

A listener connects to the beanstalk queue pointed to by the environment variable `GLOW_QUEUE`, waits for jobs and executes them as they become available. In order to achieve parallelism, a glow system will have many hosts and a number of listeners on each host. The number of listeners per host should depend on the type of job and number of available cores.

Listen options:

```
$ glow -h 2>&1 >/dev/null | grep listen
```

Start a listener:

```
$ GLOW_QUEUE=10.0.0.4:11300 glow -listen
```

Log not only errors:

```
$ glow -listen -v
```


### Tube Dependencies

A [beanstalk tube](https://github.com/kr/beanstalkd/blob/master/doc/protocol.txt#L105) is a priority based fifo queue of jobs. In glow, a tube can depend on one or more other tubes. Tube dependencies are specified in a JSON file:

```
$ cat glow-deps.json
{
 "foo": ["bar"],
 "baz": ["foo", "bar"]
}

$ glow -listen -deps=glow-deps.json
```

- Tube `foo` depends on tube bar: no jobs from `foo` will run while there are ready/delayed/reserved jobs in `bar`
- Tube `bar` does not have any dependencies. Jobs from `bar` will run whenever there are free workers available
- Tube `baz` depends on tube `bar` and `foo`. It will block until `bar` and `foo` are done
- Dependencies are not transitive. If `foo` depends on `bar` and `baz` depends on `foo`, `baz` doesn't depend on `bar`

### Excluding tubes

A listener will not reserve jobs from any of the tubes specified by the `exclude` flag:

```
$ glow -listen -exclude=foo,bar
```

### Email

The smtp server and email `FROM` field can be configured for glow's job failure email notifications:

```
$ glow -listen -smtp-server=smtp.example.com -mail-from=glow@example.com
```

Emails will only be sent when a list of recipients has been specified at job submission.

### Signals

`SIGTERM` kills a listener and its running job immediatly:

```
$ killall glow 
```

Shut down gracefully (wait for job to finish) with `SIGINT`:

```
$ killall -2 glow
```


## Submit

Submit options:

```
$ glow -h 2>&1 >/dev/null | grep submit
```

Send a job to a tube on the beanstalkd queue to be executed by a listener (`-tube` is required):

```
$ glow -tube=mytube mycommand arg1 arg2 # [...argn]
```

### Job delay

[Delay](https://github.com/kr/beanstalkd/blob/master/doc/protocol.txt#L136) is an integer number of seconds to wait before making the job avaible to run:

```
$ glow -tube=mytube -delay=60 mycommand arg1 arg2
```

### Failure emails

```
$ glow -tube=mytube -mailto=bob@example.com,alice@example.com mycommand arg1 arg2
```

### Job output

Job `stdout` and `stderr` can be redirected to a file:

```
$ glow -tube=mytube -out=/tmp/mycommand.log mycommand arg1 arg2
```

By default, a job's `stdout` and `stderr` are sent to `/dev/null`

### Job priority

[Priority](https://github.com/kr/beanstalkd/blob/master/doc/protocol.txt#L132) is an integer < 2**32. Jobs with smaller priority values will be scheduled before jobs with larger priorities:

```
$ glow -tube=mytube -pri=177 mycommand arg1 arg2
```

### Job working directory

Where to run the job from. Defaults to `/tmp`. The listener will `chdir` to `workdir` before executing the job's command:

```
$ glow -tube=mytube -workdir=/home/bob/scripts mycommand arg1 arg2
```

### Batch job submit
For improved performance when queueng up a lot of jobs at once, a json list of jobs can be piped to glow's stdin: 

```
echo '[{"cmd":"ls","arguments":["-l", "-a"],"pri":0,"tube":"foo","delay":0,"mailto":"example@example.com","out":"/tmp/glow.out","workdir":"/tmp/glow"},{"cmd":"ps","pri":1,"tube":"bar","delay":0,"mailto":"example@example.com","out":"/tmp/glow.out","workdir":"/tmp/glow"}]' | glow
```

## Errors

Every time a job exits with a non 0 exit status, glow sends a message to a tube on `GLOW_QUEUE` called `GLOW_ERRORS`. beanstalkd clients can listen on `GLOW_ERRORS` to implement custom error handling. 

Failure emails will also be sent to the list of recipients specified by the `-mailto` submit flag.

## Utilities

```
glow -h 2>&1 >/dev/null | grep -v 'submit\|listen'
```

### Drain tubes

Delete all jobs from a list tube, subsequently killing them:

```
$ glow -drain=tube1,tube2
```

The output of `drain` is json that can be used to requeue the jobs by piping to `glow`.

### Pause tubes

A list of tubes can be paused for a period of seconds specified by the `-pause-delay` int flag, during which jobs on those tubes will not be available to be reserved by listeners:

```
$ glow -pause=tube1,tube2 -pause-delay=600
```

### Queue stats

Show per tube beanstalkd queue statistics:

```
$ glow -stats
```

