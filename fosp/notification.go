// Copyright (C) 2015 Felix Maurer
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
	"fmt"
	"io"
	"net/textproto"
	"net/url"
)

// Notification is an object that represents a FOSP notification message.
type Notification struct {
	Event  string
	Header textproto.MIMEHeader
	Body   io.Reader

	URL *url.URL
}

// NewNotification creates a new Notification.
func NewNotification(event string, url *url.URL) *Notification {
	return &Notification{Event: event, Header: make(map[string][]string), Body: nil, URL: url}
}

func (n *Notification) String() string {
	return fmt.Sprintf("%s %s", n.Event, n.URL)
}

func (n *Notification) nop() {}
