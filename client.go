package main

import (
	beanstalk "github.com/nutrun/beanstalk.go"
	"encoding/json"
)

type Client struct {
	q *beanstalk.Conn
}

func NewClient(addr string) (*Client, error) {
	this := new(Client)
	q, err := beanstalk.Dial(addr)
	if err != nil {
		return nil, err
	}
	this.q = q
	return this, nil
}

func (this *Client) put(cmd, mailto string) error {
	msg := make(map[string]string)
	msg["cmd"] = cmd
	msg["mailto"] = mailto
	message, e := json.Marshal(msg)
	if e != nil {
		return e
	}
	_, e = this.q.Put(string(message), 0, 0, 1000*60*60) // An hour TTR?
	return e
}
