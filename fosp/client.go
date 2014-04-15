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
)

// Client represents a FOSP client.
type Client struct {
	connection          *Connection
	notificationHandler func(Notification)
}

// OpenConnection connects to the given FOSP domain.
func (c *Client) OpenConnection(remoteDomain string) error {
	if c.connection != nil {
		c.connection.Close()
	}
	con, err := OpenConnection(remoteDomain)
	c.connection = con
	return err
}

// SendRequest sends a request using the current connection.
// If the client is not connected, an error is returned.
func (c *Client) SendRequest(rt RequestType, url *URL, headers map[string]string, body []byte) (*Response, error) {
	if c.connection == nil {
		return nil, errors.New("client is not connected")
	}
	return c.connection.SendRequest(rt, url, headers, body)
}

// SetNotificationHandler sets up the callback method for received notifications.
func (c *Client) SetNotificationHandler(handler func(Notification)) {
	c.notificationHandler = handler
}
