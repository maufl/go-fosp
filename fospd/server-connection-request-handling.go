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
	"encoding/json"
	"github.com/maufl/go-fosp/fosp"
	"time"
)

// BUG: Seems like we are forwarding requests for other servers ...
func (c *ServerConnection) handleRequest(req *fosp.Request) *fosp.Response {
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
			if resp, err := c.server.forwardRequest(user, req.RequestType(), req.URL(), req.Headers(), req.Body()); err == nil {
				servConnLog.Debug("Response is %v+", resp)
				if resp.ResponseType() == fosp.Succeeded {
					return req.SucceededWithBody(resp.Status(), resp.Body())
				}
				return req.Failed(resp.Status(), string(resp.Body()))
			}
			return req.Failed(502, "Forwarding failed")
		}
		servConnLog.Fatal("Cannot forward request for non user")
	}

	switch req.RequestType() {
	case fosp.Select:
		return c.handleSelect(user, req)
	case fosp.Create:
		return c.handleCreate(user, req)
	case fosp.Update:
		return c.handleUpdate(user, req)
	case fosp.List:
		return c.handleList(user, req)
	case fosp.Delete:
		return c.handleDelete(user, req)
	case fosp.Read:
		return c.handleRead(user, req)
	case fosp.Write:
		return c.handleWrite(user, req)
	default:
		return req.Failed(500, "Cannot handle request type")
	}
}

func (c *ServerConnection) handleSelect(user string, req *fosp.Request) *fosp.Response {
	defer timeTrack(time.Now(), "select request")
	object, err := c.server.database.Select(user, req.URL())
	if err != nil {
		if fe, ok := err.(fosp.FospError); ok {
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

func (c *ServerConnection) handleCreate(user string, req *fosp.Request) *fosp.Response {
	defer timeTrack(time.Now(), "create request")
	o, err := req.BodyObject()
	if err != nil {
		return req.Failed(400, "Invalid body")
	}
	if err := c.server.database.Create(user, req.URL(), o); err != nil {
		return req.Failed(500, err.Error())
	}
	return req.Succeeded(200)
}

func (c *ServerConnection) handleUpdate(user string, req *fosp.Request) *fosp.Response {
	defer timeTrack(time.Now(), "update request")
	var (
		obj *fosp.UnsaveObject
		err error
	)
	if obj, err = fosp.UnmarshalUnsaveObject(req.Body()); err != nil {
		return req.Failed(400, "Invalid body :: "+err.Error())
	}
	if err = c.server.database.Update(user, req.URL(), obj); err != nil {
		return req.Failed(500, err.Error())
	}
	return req.Succeeded(200)
}

func (c *ServerConnection) handleList(user string, req *fosp.Request) *fosp.Response {
	defer timeTrack(time.Now(), "list request")
	list, err := c.server.database.List(user, req.URL())
	if err != nil {
		return req.Failed(500, err.Error())
	}
	if body, err := json.Marshal(list); err == nil {
		return req.SucceededWithBody(200, body)
	}
	return req.Failed(500, "Internal server error")
}

func (c *ServerConnection) handleDelete(user string, req *fosp.Request) *fosp.Response {
	defer timeTrack(time.Now(), "delete request")
	if err := c.server.database.Delete(user, req.URL()); err != nil {
		return req.Failed(500, err.Error())
	}
	return req.Succeeded(200)
}

func (c *ServerConnection) handleRead(user string, req *fosp.Request) *fosp.Response {
	defer timeTrack(time.Now(), "read request")
	data, err := c.server.database.Read(user, req.URL())
	if err != nil {
		return req.Failed(500, err.Error())
	}
	resp := req.SucceededWithBody(200, data)
	resp.SetType(fosp.Binary)
	return resp
}

func (c *ServerConnection) handleWrite(user string, req *fosp.Request) *fosp.Response {
	defer timeTrack(time.Now(), "write request")
	if err := c.server.database.Write(user, req.URL(), []byte(req.Body())); err != nil {
		servConnLog.Warning("Write request failed: " + err.Error())
		return req.Failed(500, err.Error())
	}
	return req.Succeeded(200)
}
