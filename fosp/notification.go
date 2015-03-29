// Copyright (C) 2014 Felix Maurer
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>

package fosp

import (
	"errors"
	"fmt"
)

// ErrInvalidEventType is returned when a string is parsed that does not represenst an Event.
var ErrInvalidEventType = errors.New("not a valid event type")

// Event is the type of event of a notification
type Event uint

const (
	// Created denotes a notification for the CREATED event.
	Created Event = 1 << iota
	// Updated denotes a notification for the UPDATED event.
	Updated
	// Deleted denotes a notification for the DELETED event.
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

// ParseEvent returns the corresponding Event for a string.
// If the string is not an event 0 and an error is returned.
func ParseEvent(s string) (Event, error) {
	switch s {
	case "CREATED":
		return Created, nil
	case "UPDATED":
		return Updated, nil
	case "DELETED":
		return Deleted, nil
	default:
		return 0, ErrInvalidEventType
	}
}

// Notification is an object that represents a FOSP notification message.
type Notification struct {
	BasicMessage
	event Event
	url   *URL
}

// NewNotification creates a new Notification.
func NewNotification(ev Event, url *URL, headers map[string]string, body string) *Notification {
	return &Notification{BasicMessage{headers, []byte(body), Text}, ev, url}
}

func (n *Notification) String() string {
	result := fmt.Sprintf("%s %s\r\n", n.event, n.url)
	for k, v := range n.headers {
		result += k + ": " + v + "\r\n"
	}
	if string(n.body) != "" {
		result += "\r\n" + string(n.body)
	}
	return result
}

// Bytes returns the string representation of the Notification as byte array.
func (n *Notification) Bytes() []byte {
	return []byte(n.String())
}

func (n *Notification) URL() *URL {
	return n.url
}
