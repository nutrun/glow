UNDER CONSTRUCTION, THE DOCS MAY NOT BE COMPLETE

# glow

Distributed job scheduling via [beankstalkd](http://kr.github.com/beanstalkd/), allows heterogenuis systems to execute a set of tasks includes:
-notification on faiure
-job dependencies
-resource isolation

### Setup 
### BUILDING

Needs golang go1 or recent weekly and [lentil](https://github.com/nutrun/lentil )

```
$ go get github.com/nutrun/lentil
```

Inside a $GOPATH/src

```
$ git clone git@git:grid/glow.git && cd glow && go install
```

### RUNNING

Start beanstalkd:

```
$ beanstalkd
```

Start a listener:

```
$ GLOW_QUEUE=localhost:11300 glow -listen -v
```

Run a job:

```
$ glow -v -tube=test -out=$HOME/glow.out ls
```

Look at output:
```
$ cat $HOME/glow.out
```

List what's available:

```
$ glow -h
```

### Listener
A listener connect to a specified beanstalk instsance configured by the enviormental varaible GLOW_QUEUE, reserves a job from beanstalk and executes it.

### Signals
Kill listener immediatly
```
$ killall glow 
```

Shut down gracefully (wait for job to finish)
```
$ killall -2 glow
```

### Jobs
Required arguments
```
cmd: Executable (string)
args: Arguments (string)
tube: Tube (string)
```
Defaulted
```
workdir: Workdir (string) default: /tmp
out: StdOut/Stderr (string)  default: /dev/null
```

Optional arguments
```
mailto: Mail error on Failure (string)
pri:  Beanstalk Job Priority   (int)
delay: Beanstalk Job Delay  (int)
```


### Tubes 
Dependencies
```

```
Priorities
```
<pri> is an integer < 2**32. Jobs with smaller priority values will be scheduled before jobs with larger priorities. 
```
Exclude
```
-exclude=<Tube,Tube> a listener will not reserve jobs from any of the specified tubes
```
### Supermegamicrooptimization

For improved queueing performance, a json list of jobs can be piped to glow's stdin: 

```
echo '[{"cmd":"ls","pri":"0","tube":"foo","delay":"0","mailto":"example@example.com","out":"/tmp/glow.out","workdir":"/tmp/glow"},{"cmd":"ps","pri":"1","tube":"bar","delay":"0","mailto":"example@example.com","out":"/tmp/glow.out","workdir":"/tmp/glow"}]' | glow
```
