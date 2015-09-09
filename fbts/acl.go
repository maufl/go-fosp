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
	"github.com/maufl/go-fosp/fosp"
)

func testAcl() (success bool) {
	defer func() {
		if r := recover(); r != nil {
			println("Error in access control test")
			if s, ok := r.(string); ok {
				println(s)
			}
			success = false
		}
	}()

	userOne := "alice"
	passwordOne := "password"
	clientOne := &fosp.Client{}
	err := clientOne.OpenConnection(host)
	if err != nil {
		panic("Failed to open connection for client one.")
	}
	expectE(clientOne.Connect())
	expectE(clientOne.Register(userOne, passwordOne))
	expectE(clientOne.Authenticate(userOne, passwordOne))

	userTwo := "bob"
	passwordTwo := "password"
	clientTwo := &fosp.Client{}
	err = clientTwo.OpenConnection(host)
	if err != nil {
		panic("Falied to open connection for client two.")
	}
	expectE(clientTwo.Connect())
	expectE(clientTwo.Register(userTwo, passwordTwo))
	expectE(clientTwo.Authenticate(userTwo, passwordTwo))

	rootOne, err := fosp.ParseURL(userOne + "@" + host)
	if err != nil {
		panic("Error when parsing root URL of user one")
	}

	expect(E{Failed: true})(clientTwo.Select(rootOne))
	return true
}
