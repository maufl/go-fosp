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
	"encoding/json"
)

// BUG: Seems like we are forwarding requests for other servers ...
func (c *ServerConnection) handleRequest(req *Request) *Response {
	var user string
	if c.user != "" {
		user = c.user + "@" + c.server.Domain()
	} else if reqUser, ok := req.Head("User"); ok {
		user = reqUser
	} else {
		servConnLog.Fatal("Received request but can't determin user!")
	}

	if req.URL().Domain() != c.server.Domain() {
		if c.user != "" {
			servConnLog.Info("Try to forward request for user " + user)
			if resp, err := c.server.forwardRequest(user, req.request, req.URL(), req.Headers(), req.Body()); err == nil {
				servConnLog.Debug("Response is %v+", resp)
				if resp.response == Succeeded {
					return req.SucceededWithBody(resp.status, resp.body)
				}
				return req.Failed(resp.status, string(resp.body))
			}
			return req.Failed(502, "Forwarding failed")
		}
		servConnLog.Fatal("Cannot forward request for non user")
	}

	switch req.request {
	case Select:
		return c.handleSelect(user, req)
	case Create:
		return c.handleCreate(user, req)
	case Update:
		return c.handleUpdate(user, req)
	case List:
		return c.handleList(user, req)
	case Delete:
		return c.handleDelete(user, req)
	case Read:
		return c.handleRead(user, req)
	case Write:
		return c.handleWrite(user, req)
	default:
		return req.Failed(500, "Cannot handle request type")
	}
}

func (c *ServerConnection) handleSelect(user string, req *Request) *Response {
	object, err := c.server.database.Select(user, req.url)
	if err != nil {
		if fe, ok := err.(FospError); ok {
			return req.Failed(fe.Code, fe.Message)
		}
		return req.Failed(500, "Internal database error")
	}
	body, err := json.Marshal(object)
	if err != nil {
		return req.Failed(500, "Internal server error")
	}
	return req.SucceededWithBody(200, body)
}

func (c *ServerConnection) handleCreate(user string, req *Request) *Response {
	o, err := req.BodyObject()
	if err != nil {
		return req.Failed(400, "Invalid body")
	}
	if err := c.server.database.Create(user, req.url, o); err != nil {
		return req.Failed(500, err.Error())
	}
	return req.Succeeded(200)
}

func (c *ServerConnection) handleUpdate(user string, req *Request) *Response {
	var (
		obj *UnsaveObject
		err error
	)
	if obj, err = UnmarshalUnsaveObject(req.Body()); err != nil {
		return req.Failed(400, "Invalid body :: "+err.Error())
	}
	if err = c.server.database.Update(user, req.url, obj); err != nil {
		return req.Failed(500, err.Error())
	}
	return req.Succeeded(200)
}

func (c *ServerConnection) handleList(user string, req *Request) *Response {
	list, err := c.server.database.List(user, req.url)
	if err != nil {
		return req.Failed(500, err.Error())
	}
	if body, err := json.Marshal(list); err == nil {
		return req.SucceededWithBody(200, body)
	}
	return req.Failed(500, "Internal server error")
}

func (c *ServerConnection) handleDelete(user string, req *Request) *Response {
	if err := c.server.database.Delete(user, req.url); err != nil {
		return req.Failed(500, err.Error())
	}
	return req.Succeeded(200)
}

func (c *ServerConnection) handleRead(user string, req *Request) *Response {
	data, err := c.server.database.Read(user, req.url)
	if err != nil {
		return req.Failed(500, err.Error())
	}
	resp := req.SucceededWithBody(200, data)
	resp.SetType(Binary)
	return resp
}

func (c *ServerConnection) handleWrite(user string, req *Request) *Response {
	if err := c.server.database.Write(user, req.url, []byte(req.Body())); err != nil {
		servConnLog.Warning("Write request failed: " + err.Error())
		return req.Failed(500, err.Error())
	}
	return req.Succeeded(200)
}
