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

package main

import (
	"encoding/json"
	"errors"
	"github.com/maufl/go-fosp/fosp"
	"net"
	"sync/atomic"
)

// ErrInvalidConnectionState is returned when a request is recieved that is not allowed in the current connection state.
var ErrInvalidConnectionState = errors.New("invalid connection state")

// ErrUnsupportedFOSPVersion is returned when a client wants a FOSP version for the connection that is not supported by the server.
var ErrUnsupportedFOSPVersion = errors.New("unsupported FOSP version")

// ErrCredentialsMissing is returned when a client did not supply credentials for authentication.
var ErrCredentialsMissing = errors.New("credentials missing")

// AuthenticationObject represents the information sent in an AUTHENTICATE request.
type AuthenticationObject struct {
	Name     string
	Password string
	Type     string
	Domain   string
}

// ConnectionNegotiationObject represents the information sent in a CONNECT request.
type ConnectionNegotiationObject struct {
	Version string
}

func (c *ServerConnection) bootstrap(req *fosp.Request) {
	servConnLog.Info("Bootstraping connection")
	if atomic.CompareAndSwapUint32(&c.state, Opened, Opened) {
		servConnLog.Info("Connection needs negotiation")
		if err := c.negotiate(req); err != nil {
			//c.ws.Close()
			//break
		}
	} else if req.RequestType() == fosp.Register {
		if err := c.register(req); err != nil {
			//c.ws.Close()
			//break
		}
	} else if atomic.CompareAndSwapUint32(&c.state, Negotiated, Negotiated) {
		servConnLog.Info("Connection needs authentication")
		if err := c.authenticate(req); err != nil {
			//c.ws.Close()
			//break
		} else {
			c.server.registerConnection(c, c.user+"@")
		}
	} else {
		//Invalid state
		servConnLog.Critical("Connection should be bootstraped but is not in Opened or Negotiated state")
	}
}

func (c *ServerConnection) negotiate(req *fosp.Request) error {
	if req.RequestType() != fosp.Connect {
		servConnLog.Warning("Recieved message on not negotiated connection")
		return ErrInvalidConnectionState
	}
	var obj ConnectionNegotiationObject
	err := json.Unmarshal([]byte(req.Body()), &obj)
	if err != nil {
		servConnLog.Error("Error when unmarshaling object " + err.Error())
		return err
	}
	if obj.Version != "0.1" {
		c.Send(req.Failed(400, "Version not supported"))
		return ErrUnsupportedFOSPVersion
	}
	if atomic.CompareAndSwapUint32(&c.state, Opened, Negotiated) {
		c.Send(req.Succeeded(200))
		return nil
	}
	c.Send(req.Failed(4000, "Invalid connection state"))
	return ErrInvalidConnectionState
}

func (c *ServerConnection) authenticate(req *fosp.Request) error {
	if req.RequestType() != fosp.Authenticate {
		servConnLog.Warning("Recieved message on not authenticated connection")
		return ErrInvalidConnectionState
	}
	var obj AuthenticationObject
	err := json.Unmarshal([]byte(req.Body()), &obj)
	if err != nil {
		servConnLog.Error("Error when unmarshaling object")
		return err
	}
	if obj.Type == "server" {
		servConnLog.Info("Authenticating server %v+", obj)
		remoteAddr := c.Ws.RemoteAddr()
		if tcpAddr, ok := remoteAddr.(*net.TCPAddr); ok {
			servConnLog.Info("Remote address is %v", tcpAddr.IP.String())
			resolvedNames, err := net.LookupAddr(tcpAddr.IP.String())
			if err != nil {
				servConnLog.Error("Reverse lookup failed ", err.Error())
				c.Send(req.Failed(403, "Revers lookup did not succeed"))
				return nil
			}
			servConnLog.Info("Reverse lookup found %v+\n", resolvedNames)
			for _, name := range resolvedNames {
				if name == obj.Domain || name == obj.Domain+"." {
					if atomic.CompareAndSwapUint32(&c.state, Negotiated, Authenticated) {
						c.remoteDomain = obj.Domain
						c.Send(req.Succeeded(200))
						return nil
					}
					return ErrInvalidConnectionState
				}
			}
		}
		c.Send(req.Failed(403, "Revers lookup did not match or did not succeed"))
		return nil
	}
	if obj.Name == "" || obj.Password == "" {
		c.Send(req.Failed(400, "Name or password missing"))
		return ErrCredentialsMissing
	}
	servConnLog.Info("Authenticating user %v", obj)
	if err := c.server.database.Authenticate(obj.Name, obj.Password); err == nil {
		if atomic.CompareAndSwapUint32(&c.state, Negotiated, Authenticated) {
			c.user = obj.Name
			c.Send(req.Succeeded(200))
			return nil
		}
		c.Send(req.Failed(4000, "Invalid connection state"))
		return ErrInvalidConnectionState
	}
	c.Send(req.Failed(403, "Invalid user or password"))
	return nil
}

func (c *ServerConnection) register(req *fosp.Request) error {
	if req.RequestType() != fosp.Register {
		servConnLog.Fatal("Tried to register but request is not a REGISTER request")
	}
	var obj AuthenticationObject
	if err := json.Unmarshal([]byte(req.Body()), &obj); err != nil {
		return err
	}
	if obj.Name == "" || obj.Password == "" {
		c.Send(req.Failed(400, "Name or password missing"))
		return ErrCredentialsMissing
	}
	if err := c.server.database.Register(obj.Name, obj.Password); err != nil {
		c.Send(req.Failed(500, err.Error()))
		return nil
	}
	c.Send(req.Succeeded(200))
	return nil
}
