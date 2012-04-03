# glow

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

## INSTALLATION

Needs golang go1 or recent weekly. 

```
go get http://git/grid/glow
```

or, inside a $GOPATH

```
git clone git@git:grid/glow.git && cd glow && go install
```
