package main

import (
  "encoding/json"
  "log"
)

func (c *connection) handleRequest(req *Request) *Response {
  log.Println("Received new request %v+", req)
  switch req.request {
  case Select:
    return c.handleSelect(req)
  case Create:
    return c.handleCreate(req)
  case Update:
    return c.handleUpdate(req)
  case List:
    return c.handleList(req)
  case Delete:
    return c.handleDelete(req)
  default:
    return &Response{ response: Failed, status: 500, seq: req.seq, body: "Cannot handle request type"}
  }
}

func (c *connection) handleSelect(req *Request) *Response {
  object, err := c.server.database.Select(c.user + "@" + c.server.Domain(), req.url)
  if err != nil {
    return &Response{ response: Failed, status: 500, seq: req.seq, body: "Internal database error" }
  }
  body, err := json.Marshal(object)
  if err != nil {
    return &Response{ response: Failed, status: 500, seq: req.seq, body: "Internal server error" }
  }
  return &Response{ response: Succeeded, status: 200, seq: req.seq, body: string(body) }
}

func (c *connection) handleCreate(req *Request) *Response {
  o, err := req.GetBodyObject()
  if err != nil {
    return &Response{response: Failed, status: 400, seq: req.seq, body: "Invalid body" }
  }
  if err := c.server.database.Create(c.user + "@" + c.server.Domain(), req.url, o); err != nil {
    return &Response{response: Failed, status: 500, seq: req.seq, body: err.Error() }
  }
  return &Response{response: Succeeded, status: 200, seq: req.seq }
}

func (c *connection) handleUpdate(req *Request) *Response {
  var (
    obj *Object
    err error
  )
  if obj, err = req.GetBodyObject(); err != nil {
    return req.Failed(400, "Invalid body :: " + err.Error())
  }
  if err = c.server.database.Update(c.user + "@" + c.server.Domain(), req.url, obj); err != nil {
    return req.Failed(500, err.Error())
  }
  return req.Succeeded(200, "")
}

func (c *connection) handleList(req *Request) *Response {
  if list, err := c.server.database.List(c.user + "@" + c.server.Domain(), req.url); err != nil {
    return req.Failed(500, err.Error())
  } else {
    if body, err := json.Marshal(list); err != nil {
      return req.Failed(500, "Internal server error")
    } else {
      return req.Succeeded(200,string(body))
    }
  }
}

func (c *connection) handleDelete(req *Request) *Response {
  if err := c.server.database.Delete(c.user + "@" + c.server.Domain(), req.url); err != nil {
    return req.Failed(500, err.Error())
  } else {
    return req.Succeeded(200, "")
  }
}
