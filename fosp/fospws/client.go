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

package fospws

import (
	"errors"
	"github.com/maufl/go-fosp/fosp"
)

// ErrClientNotConnected is returned when a request should be sent but the client is not connected.
var ErrClientNotConnected = errors.New("client is not connected")

// Client represents a FOSP client.
type Client struct {
	Connection *Connection
}

// OpenConnection connects to the given FOSP domain.
func (c *Client) OpenConnection(remoteDomain string) (err error) {
	if c.Connection != nil {
		c.Connection.Close()
	}
	c.Connection, err = OpenConnection(remoteDomain)
	return err
}

// SendRequest sends a request using the current connection.
// If the client is not connected, an error is returned.
func (c *Client) SendRequest(request *fosp.Request) (*fosp.Response, error) {
	if c.Connection == nil {
		return nil, ErrClientNotConnected
	}
	return c.Connection.SendRequest(request)
}
