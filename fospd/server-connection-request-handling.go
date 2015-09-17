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
	"bytes"
	"encoding/json"
	"github.com/maufl/go-fosp/fosp"
	"io/ioutil"
	"time"
)

// BUG: Seems like we are forwarding requests for other servers ...
func (c *ServerConnection) handleRequest(req *fosp.Request) *fosp.Response {
	servConnLog.Debug("Handeling request %#v", req)
	servConnLog.Debug("URL is %#v", req.URL)
	if req.URL != nil && req.URL.Host != c.server.Domain() {
		if c.User != "" {
			servConnLog.Info("Try to forward request for user " + c.User)
			if resp, err := c.server.forwardRequest(c.User, req); err == nil {
				servConnLog.Debug("Response is %v+", resp)
				return resp
			}
			return fosp.NewResponse(fosp.FAILED, fosp.StatusBadGateway)
		}
		servConnLog.Fatal("Cannot forward request for non user")
	}

	var user string
	if c.User != "" {
		user = c.User
	} else if reqUser := req.Header.Get("From"); reqUser != "" {
		user = reqUser
	}

	switch req.Method {
	case fosp.AUTH:
		return c.handleAuth(req)
	case fosp.GET:
		return c.handleGet(user, req)
	case fosp.CREATE:
		return c.handleCreate(user, req)
	case fosp.PATCH:
		return c.handlePatch(user, req)
	case fosp.LIST:
		return c.handleList(user, req)
	case fosp.DELETE:
		return c.handleDelete(user, req)
	case fosp.READ:
		return c.handleRead(user, req)
	case fosp.WRITE:
		return c.handleWrite(user, req)
	default:
		return fosp.NewResponse(fosp.FAILED, fosp.StatusBadRequest)
	}
}

func (c *ServerConnection) handleGet(user string, req *fosp.Request) *fosp.Response {
	defer timeTrack(time.Now(), "select request")
	object, err := c.server.database.Get(user, req.URL)
	if err != nil {
		if fe, ok := err.(FospError); ok {
			return fosp.NewResponse(fosp.FAILED, fe.Code)
		}
		return fosp.NewResponse(fosp.FAILED, fosp.StatusInternalServerError)
	}
	body, err := json.Marshal(object)
	if err != nil {
		return fosp.NewResponse(fosp.FAILED, fosp.StatusInternalServerError)
	}
	resp := fosp.NewResponse(fosp.SUCCEEDED, fosp.StatusOK)
	resp.Body = bytes.NewBuffer(body)
	return resp
}

func (c *ServerConnection) handleCreate(user string, req *fosp.Request) *fosp.Response {
	defer timeTrack(time.Now(), "create request")
	obj := fosp.NewObject()
	if err := json.NewDecoder(req.Body).Decode(obj); err != nil {
		servConnLog.Warning("Unable to decode CREATE body :: %s", err)
		return fosp.NewResponse(fosp.FAILED, fosp.StatusBadRequest)
	}
	if err := c.server.database.Create(user, req.URL, obj); err != nil {
		return fosp.NewResponse(fosp.FAILED, fosp.StatusInternalServerError)
	}
	return fosp.NewResponse(fosp.SUCCEEDED, fosp.StatusCreated)
}

func (c *ServerConnection) handlePatch(user string, req *fosp.Request) *fosp.Response {
	defer timeTrack(time.Now(), "update request")
	var obj fosp.PatchObject
	if err := json.NewDecoder(req.Body).Decode(&obj); err != nil {
		servConnLog.Warning("Unable to decode PATCH body :: %s", err)
		return fosp.NewResponse(fosp.FAILED, fosp.StatusBadRequest)
	}
	if err := c.server.database.Patch(user, req.URL, obj); err != nil {
		return fosp.NewResponse(fosp.FAILED, fosp.StatusInternalServerError)
	}
	return fosp.NewResponse(fosp.SUCCEEDED, fosp.StatusNoContent)
}

func (c *ServerConnection) handleList(user string, req *fosp.Request) *fosp.Response {
	defer timeTrack(time.Now(), "list request")
	list, err := c.server.database.List(user, req.URL)
	if err != nil {
		return fosp.NewResponse(fosp.FAILED, fosp.StatusInternalServerError)
	}
	if body, err := json.Marshal(list); err == nil {
		resp := fosp.NewResponse(fosp.SUCCEEDED, fosp.StatusOK)
		resp.Body = bytes.NewBuffer(body)
		return resp
	}
	return fosp.NewResponse(fosp.FAILED, fosp.StatusInternalServerError)
}

func (c *ServerConnection) handleDelete(user string, req *fosp.Request) *fosp.Response {
	defer timeTrack(time.Now(), "delete request")
	if err := c.server.database.Delete(user, req.URL); err != nil {
		return fosp.NewResponse(fosp.FAILED, fosp.StatusInternalServerError)
	}
	return fosp.NewResponse(fosp.SUCCEEDED, fosp.StatusNoContent)
}

func (c *ServerConnection) handleRead(user string, req *fosp.Request) *fosp.Response {
	defer timeTrack(time.Now(), "read request")
	data, err := c.server.database.Read(user, req.URL)
	if err != nil {
		return fosp.NewResponse(fosp.FAILED, fosp.StatusInternalServerError)
	}
	resp := fosp.NewResponse(fosp.SUCCEEDED, fosp.StatusOK)
	resp.Body = bytes.NewBuffer(data)
	return resp
}

func (c *ServerConnection) handleWrite(user string, req *fosp.Request) *fosp.Response {
	defer timeTrack(time.Now(), "write request")
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return fosp.NewResponse(fosp.FAILED, fosp.StatusInternalServerError)
	}
	if err := c.server.database.Write(user, req.URL, data); err != nil {
		servConnLog.Warning("Write request failed: " + err.Error())
		return fosp.NewResponse(fosp.FAILED, fosp.StatusInternalServerError)
	}
	return fosp.NewResponse(fosp.SUCCEEDED, fosp.StatusNoContent)
}
