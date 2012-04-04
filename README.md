# glow

Distributed job scheduling via [beankstalkd](http://kr.github.com/beanstalkd/)

### RUNNING

Start beanstalkd:

```
$ beanstalkd
```

Start a listener:

```
$ glow -listen -v
```

Run a job:

```
$ glow -v -out=$HOME/glow.out ls
```

List what's available:

```
$ glow -h
```

### BUILDING

Needs golang go1 or recent weekly and [lentil](https://github.com/nutrun/lentil )

```
$ go get github.com/nutrun/lentil
```

Inside a $GOPATH/src

```
$ git clone git@git:grid/glow.git && cd glow && go install
```
