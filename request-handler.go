package main

import (
	"encoding/json"
	"log"
)

func (c *connection) handleRequest(req *Request) *Response {
	log.Println("Received new request %v+", req)
	var user string
	if c.user != "" {
		user = c.user + "@" + c.server.Domain()
	} else if reqUser, ok := req.GetHead("User"); ok {
		user = reqUser
	} else {
		panic("No user for this request!")
	}

	if req.Url().Domain() != c.server.Domain() {
		if c.user != "" {
			log.Println("Try to forward request")
			if resp, err := c.server.forwardRequest(user, req.request, req.Url(), req.Headers(), req.GetBody()); err == nil {
				return resp
			} else {
				return req.Failed(502, "Forwarding failed")
			}
		} else {
			panic("Cannot forward request for non user")
		}
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
	default:
		return req.Failed(500, "Cannot handle request type")
	}
}

func (c *connection) handleSelect(user string, req *Request) *Response {
	object, err := c.server.database.Select(user, req.url)
	if err != nil {
		return req.Failed(500, "Internal database error")
	}
	body, err := json.Marshal(object)
	if err != nil {
		return req.Failed(500, "Internal server error")
	}
	return req.Succeeded(200, string(body))
}

func (c *connection) handleCreate(user string, req *Request) *Response {
	o, err := req.GetBodyObject()
	if err != nil {
		return req.Failed(400, "Invalid body")
	}
	if err := c.server.database.Create(user, req.url, o); err != nil {
		return req.Failed(500, err.Error())
	}
	return req.Succeeded(200, "")
}

func (c *connection) handleUpdate(user string, req *Request) *Response {
	var (
		obj *Object
		err error
	)
	if obj, err = req.GetBodyObject(); err != nil {
		return req.Failed(400, "Invalid body :: "+err.Error())
	}
	if err = c.server.database.Update(user, req.url, obj); err != nil {
		return req.Failed(500, err.Error())
	}
	return req.Succeeded(200, "")
}

func (c *connection) handleList(user string, req *Request) *Response {
	if list, err := c.server.database.List(user, req.url); err != nil {
		return req.Failed(500, err.Error())
	} else {
		if body, err := json.Marshal(list); err != nil {
			return req.Failed(500, "Internal server error")
		} else {
			return req.Succeeded(200, string(body))
		}
	}
}

func (c *connection) handleDelete(user string, req *Request) *Response {
	if err := c.server.database.Delete(user, req.url); err != nil {
		return req.Failed(500, err.Error())
	} else {
		return req.Succeeded(200, "")
	}
}
