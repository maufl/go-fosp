package main

import (
	"encoding/json"
	"errors"
	_ "github.com/gorilla/websocket"
	"log"
	"net"
	_ "sync/atomic"
)

func (c *connection) bootstrap(req *Request) {
	log.Println("Bootstraping connection")
	if !c.negotiated {
		log.Println("Connection needs negotiation")
		if err := c.negotiate(req); err != nil {
			//c.ws.Close()
			//break
		}
	} else if req.request == Register {
		if err := c.register(req); err != nil {
			//c.ws.Close()
			//break
		}
	} else if !c.authenticated {
		log.Println("Connection needs authentication")
		if err := c.authenticate(req); err != nil {
			//c.ws.Close()
			//break
		} else {
			c.server.registerConnection(c, c.user+"@")
		}
	} else {
		//Invalid state
	}
}

func (c *connection) negotiate(req *Request) error {
	if req.request != Connect {
		log.Println("Recieved message on not negotiated connection")
		return errors.New("Recieved message on not negotiated connection")
	}
	var obj ConnectionNegotiationObject
	err := json.Unmarshal([]byte(req.body), &obj)
	if err != nil {
		log.Println("Error when unmarshaling object " + err.Error())
		return err
	} else if obj.Version != "0.1" {
		c.send(req.Failed(400, "Version not supported"))
		return errors.New("Unsupported FOSP version :: " + obj.Version)
	} else {
		c.negotiated = true
		c.send(req.Succeeded(200, ""))
		return nil
	}
}

func (c *connection) authenticate(req *Request) error {
	if req.request != Authenticate {
		log.Println("Recieved message on not authenticated connection")
		return errors.New("Recieved message on not authenticated connection")
	}
	var obj AuthenticationObject
	err := json.Unmarshal([]byte(req.body), &obj)
	if err != nil {
		log.Println("Error when unmarshaling object")
		return err
	} else if obj.Type == "server" {
		log.Printf("Authenticating server %v+\n", obj)
		remoteAddr := c.ws.RemoteAddr()
		if tcpAddr, ok := remoteAddr.(*net.TCPAddr); ok {
			log.Printf("Remote address is %v\n", tcpAddr.IP.String())
			resolvedNames, err := net.LookupAddr(tcpAddr.IP.String())
			if err != nil {
				log.Println("Reverse lookup failed ", err.Error())
				c.send(req.Failed(403, "Revers lookup did not succeed"))
				return nil
			}
			log.Printf("Reverse lookup found %v+\n", resolvedNames)
			for _, name := range resolvedNames {
				if name == obj.Domain || name == obj.Domain+"." {
					c.authenticated = true
					c.remote_domain = obj.Domain
					c.send(req.Succeeded(200, ""))
					return nil
				}
			}
		}
		c.send(req.Failed(403, "Revers lookup did not match or did not succeed"))
		return nil
	} else if obj.Name == "" || obj.Password == "" {
		c.send(req.Failed(400, "Name or password missing"))
		return errors.New("Name of password missing")
	} else {
		log.Printf("Authenticating user %v", obj)
		if err := c.server.database.Authenticate(obj.Name, obj.Password); err == nil {
			c.authenticated = true
			c.user = obj.Name
			c.send(req.Succeeded(200, ""))
			return nil
		} else {
			c.send(req.Failed(403, "Invalid user or password"))
			return nil
		}
	}
}

func (c *connection) register(req *Request) error {
	if req.request != Register {
		log.Fatal("Tried to register but request is not a REGISTER request")
	}
	var obj AuthenticationObject
	err := json.Unmarshal([]byte(req.body), &obj)
	if err != nil {
		return err
	} else if obj.Name == "" || obj.Password == "" {
		c.send(req.Failed(400, "Name or password missing"))
		return errors.New("Name of password missing")
	} else {
		if err := c.server.database.Register(obj.Name, obj.Password); err == nil {
			c.send(req.Succeeded(200, ""))
			return nil
		} else {
			c.send(req.Failed(500, err.Error()))
			return nil
		}
	}
}
