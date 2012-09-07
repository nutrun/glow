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

List what's available:

```
$ glow -h
```

### Listener
A listener connect to a specified beanstalk instsance configured by the enviormental varaible GLOW_QUEUE, reserves a job from beanstalk and executes it.

### Signals
 	- Kill listener immediatly
	- Shut down gracefully (wait for job to finish)

### Jobs
	- Required arguments
	- Optional arguments

### Tubes
	- Dependencies
	- Priorities

### Supermegamicrooptimization

For improved queueing performance, a json list of jobs can be piped to glow's stdin: 

```
echo '[{"cmd":"ls","pri":"0","tube":"foo","delay":"0","mailto":"example@example.com","out":"/tmp/glow.out","workdir":"/tmp/glow"},{"cmd":"ps","pri":"1","tube":"bar","delay":"0","mailto":"example@example.com","out":"/tmp/glow.out","workdir":"/tmp/glow"}]' | glow
```
