package main

import (
  _ "github.com/gorilla/websocket"
  _ "log"
  _ "encoding/json"
  _ "errors"
  _ "sync/atomic"
  _ "net"
)

func (c *connection) handleMessage(msg Message) {
  if req, ok := msg.(*Request); ok {
    resp := c.handleRequest(req)
    c.send(resp)
  }
  if resp, ok := msg.(*Response); ok {
    c.handleResponse(resp)
  }
  if ntf, ok := msg.(*Notification); ok {
    c.handleNotification(ntf)
  }
}


func (c *connection) handleResponse(resp *Response) {
  if ch, ok := c.pendingRequests[uint64(resp.seq)]; ok {
    ch <- resp
  }
}

func (c *connection) handleNotification(ntf *Notification) {

}