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

package main

import (
	"github.com/maufl/go-fosp/fosp"
	"github.com/maufl/go-fosp/fosp/fospws"
)

func (c *ServerConnection) HandleMessage(inMsg *fospws.NumberedMessage) {
	msg := inMsg.Message
	if req, ok := msg.(*fosp.Request); ok {
		resp := c.handleRequest(req)
		if req.Method == fosp.READ {
			c.SendNumberedMessage(&fospws.NumberedMessage{Message: resp, Seq: inMsg.Seq, BinaryBody: true})
		} else {
			c.Send(resp, inMsg.Seq)
		}
	}
	if ntf, ok := msg.(*fosp.Notification); ok {
		c.handleNotification(ntf)
	}
}

func (c *ServerConnection) handleNotification(ntf *fosp.Notification) {
	// TODO This is not correct yet, user has to be local!
	if user := ntf.Header.Get("To"); user != "" {
		c.server.routeNotification(user, ntf)
	}
}
