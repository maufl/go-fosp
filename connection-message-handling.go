package main

import (
	_ "encoding/json"
	_ "errors"
	_ "github.com/gorilla/websocket"
	"log"
	_ "net"
	_ "sync/atomic"
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
	log.Println("Received new response")
	if ch, ok := c.pendingRequests[uint64(resp.seq)]; ok {
		log.Println("Returning response to caller")
		ch <- resp
	}
}

func (c *connection) handleNotification(ntf *Notification) {
	if user, ok := ntf.Head("User"); ok && user != "" {
		c.server.routeNotification(user, ntf)
	}
}
