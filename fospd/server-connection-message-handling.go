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

package main

import (
	"github.com/maufl/go-fosp/fosp"
)

func (c *ServerConnection) handleMessage(msg fosp.Message) {
	if req, ok := msg.(*fosp.Request); ok {
		resp := c.handleRequest(req)
		c.Send(resp)
	}
	if resp, ok := msg.(*fosp.Response); ok {
		c.handleResponse(resp)
	}
	if ntf, ok := msg.(*fosp.Notification); ok {
		c.handleNotification(ntf)
	}
}

func (c *ServerConnection) handleResponse(resp *fosp.Response) {
	servConnLog.Info("Received new response: %s %d %d", resp.ResponseType(), resp.Status(), resp.Seq())
	c.PendingRequestsLock.RLock()
	if ch, ok := c.PendingRequests[uint64(resp.Seq())]; ok {
		servConnLog.Debug("Returning response to caller")
		ch <- resp
	}
	c.PendingRequestsLock.RUnlock()
}

func (c *ServerConnection) handleNotification(ntf *fosp.Notification) {
	if user, ok := ntf.Head("User"); ok && user != "" {
		c.server.routeNotification(user, ntf)
	}
}
