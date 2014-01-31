package main

import (
	"errors"
	"fmt"
)

type Event uint

const (
	Created Event = 1 << iota
	Updated
	Deleted
)

func (ev Event) String() string {
	switch ev {
	case Created:
		return "CREATED"
	case Updated:
		return "UPDATED"
	case Deleted:
		return "DELETED"
	default:
		return "NA_EVENT_TYPE"
	}
}

func ParseEvent(s string) (Event, error) {
	switch s {
	case "CREATED":
		return Created, nil
	case "UPDATED":
		return Updated, nil
	case "DELETED":
		return Deleted, nil
	default:
		return 0, errors.New("Not a valid event type")
	}
}

type Notification struct {
	BasicMessage
	event Event
	url   *Url
}

func NewNotification(ev Event, url *Url, headers map[string]string, body string) *Notification {
	return &Notification{BasicMessage{headers, body}, ev, url}
}

func (n *Notification) String() string {
	result := fmt.Sprintf("%s %s\r\n", n.event, n.url)
	for k, v := range n.headers {
		result += k + ": " + v + "\r\n"
	}
	if n.body != "" {
		result += "\r\n" + n.body
	}
	return result
}
