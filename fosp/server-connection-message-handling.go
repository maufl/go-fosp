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

func (c *ServerConnection) handleMessage(msg Message) {
	if req, ok := msg.(*Request); ok {
		resp := c.handleRequest(req)
		c.Send(resp)
	}
	if resp, ok := msg.(*Response); ok {
		c.handleResponse(resp)
	}
	if ntf, ok := msg.(*Notification); ok {
		c.handleNotification(ntf)
	}
}

func (c *ServerConnection) handleResponse(resp *Response) {
	servConnLog.Info("Received new response: %s %d %d", resp.response, resp.status, resp.seq)
	c.pendingRequestsLock.RLock()
	if ch, ok := c.pendingRequests[uint64(resp.seq)]; ok {
		servConnLog.Debug("Returning response to caller")
		ch <- resp
	}
	c.pendingRequestsLock.RUnlock()
}

func (c *ServerConnection) handleNotification(ntf *Notification) {
	if user, ok := ntf.Head("User"); ok && user != "" {
		c.server.routeNotification(user, ntf)
	}
}
