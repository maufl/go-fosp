package fosp

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
	c.lg.Info("Received new response: %s %d %d", resp.response, resp.status, resp.seq)
	if ch, ok := c.pendingRequests[uint64(resp.seq)]; ok {
		c.lg.Debug("Returning response to caller")
		ch <- resp
	}
}

func (c *connection) handleNotification(ntf *Notification) {
	if user, ok := ntf.Head("User"); ok && user != "" {
		c.server.routeNotification(user, ntf)
	}
}
