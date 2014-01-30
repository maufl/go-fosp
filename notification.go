package main

import (
	"errors"
	"fmt"
)

type Event uint

const (
	_             = iota
	Created Event = iota
	Updated
	Deleted
)

var eventToString = map[Event]string{
	Created: "CREATED",
	Updated: "UPDATED",
	Deleted: "DELETED",
}

var stringToEvent = map[string]Event{
	"CREATED": Created,
	"UPDATED": Updated,
	"DELETED": Deleted,
}

func (ev Event) String() string {
	if str, ok := eventToString[ev]; ok {
		return str
	} else {
		return "NA_EVENT_TYPE"
	}
}

func GetEvent(s string) (Event, error) {
	if ev, ok := stringToEvent[s]; ok {
		return ev, nil
	} else {
		return 0, errors.New("Not a valid event")
	}
}

type Notification struct {
	headers map[string]string
	body    string

	event Event
	url   *Url
}

func (r *Notification) SetHead(k, v string) {
	r.headers[k] = v
}

func (r Notification) GetHead(k string) (string, bool) {
	head, ok := r.headers[k]
	return head, ok
}

func (r *Notification) SetBody(b string) {
	r.body = b
}

func (r Notification) GetBody() string {
	return r.body
}

func (r Notification) String() string {
	result := fmt.Sprintf("%s %s\r\n", r.event, r.url)
	for k, v := range r.headers {
		result += k + ": " + v + "\r\n"
	}
	if r.body != "" {
		result += "\r\n" + r.body
	}
	return result
}

func (r Notification) Bytes() []byte {
	return []byte(r.String())
}
