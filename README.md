# glow

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

Needs golang go1 or recent weekly. 

```
$ go get github.com/nutrun/lentil && go get git/grid/glow
```

or, inside a $GOPATH

```
git clone git@git:grid/glow.git && cd glow && go install
```
