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
	"strings"
)

type SaslObject struct {
	Mechanism       string  `json:"mechanism,omitempty"`
	InitialResponse *string `json:"initial-response,omitempty"`
	Challenge       string  `json:"challende,omitempty"`
	Response        string  `json:"response,omitempty"`
	Outcome         string  `json:"outcome,omitempty"`
	AdditionalData  *string `json:"additional-data,omitempty"`
}

type AuthenticationObject struct {
	Sasl SaslObject `json:"sasl"`
}

func (c *ServerConnection) handleAuth(req *fosp.Request) *fosp.Response {
	authObj := &AuthenticationObject{Sasl: SaslObject{}}
	err := json.NewDecoder(req.Body).Decode(authObj)
	if err != nil {
		return fosp.NewResponse(fosp.FAILED, fosp.StatusBadRequest)
	}
	response := ""
	if authObj.Sasl.Mechanism != "" {
		if authObj.Sasl.Mechanism != "PLAIN" {
			return fosp.NewResponse(fosp.FAILED, fosp.StatusNotImplemented)
		}
		c.SaslMechanism = "PLAIN"
		if authObj.Sasl.InitialResponse == nil {
			content := AuthenticationObject{Sasl: SaslObject{Challenge: "Please provide your user name and password"}}
			encoded, err := json.Marshal(content)
			if err != nil {
				return fosp.NewResponse(fosp.FAILED, fosp.StatusInternalServerError)
			}
			resp := fosp.NewResponse(fosp.SUCCEEDED, fosp.StatusAdditionalDataNeeded)
			resp.Body = bytes.NewBuffer(encoded)
			return resp
		}
		response = *authObj.Sasl.InitialResponse
	} else {
		if c.SaslMechanism != "PLAIN" {
			return fosp.NewResponse(fosp.FAILED, fosp.StatusBadRequest)
		}
		if authObj.Sasl.Response == "" {
			return fosp.NewResponse(fosp.FAILED, fosp.StatusBadRequest)
		}
		response = authObj.Sasl.Response
	}
	parts := strings.Split(response, "\x00")
	if len(parts) != 3 {
		return fosp.NewResponse(fosp.FAILED, fosp.StatusBadRequest)
	}
	authorizationId := parts[0]
	authenticationId := parts[1]
	password := parts[2]

	if authorizationId != "" && authorizationId != authenticationId {
		content := AuthenticationObject{Sasl: SaslObject{Outcome: "Authorization ID and authentication ID must be the same"}}
		encoded, err := json.Marshal(content)
		if err != nil {
			return fosp.NewResponse(fosp.FAILED, fosp.StatusInternalServerError)
		}
		resp := fosp.NewResponse(fosp.FAILED, fosp.StatusUnauthorized)
		resp.Body = bytes.NewBuffer(encoded)
		return resp
	}
	servConnLog.Debug("Authenticating user %s", authenticationId)
	if c.server.database.Authenticate(authenticationId, password) {
		c.User = authenticationId
		return fosp.NewResponse(fosp.SUCCEEDED, fosp.StatusOK)
	}
	return fosp.NewResponse(fosp.FAILED, fosp.StatusUnauthorized)
}
