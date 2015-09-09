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

func testSanityCheck() (success bool) {
	defer func() {
		if r := recover(); r != nil {
			println("Error in sanity-check test")
			if s, ok := r.(string); ok {
				println(s)
			}
			success = false
		}
	}()
	user := "test"
	password := "password"
	url, _ := fosp.ParseURL(user + "@" + host + "/")
	child, _ := fosp.ParseURL(user + "@" + host + "/foo")
	obj1, _ := fosp.UnmarshalObject(`{"data": "foo"}`)
	obj2, _ := fosp.UnmarshalObject(`{"data": "bar"}`)
	objNoWrite, _ := fosp.UnmarshalObject(`{ "acl": { "owner": ["not-data-write"], "users": { "` + user + "@" + host + `" : [ "not-data-write" ] } } }`)
	attachment := []byte("Hello World!")

	client := &fosp.Client{}
	err := client.OpenConnection(host)
	if err != nil {
		println("Failed to open connection")
		success = false
		return
	}
	expectE(client.Connect())
	expectE(client.Register(user, password))
	expectE(client.Authenticate(user, password))
	expectE(client.Select(url))
	expectE(client.List(url))
	expectE(client.Create(child, obj1))
	expectE(client.List(url))
	expectE(client.Update(child, obj2))
	expectE(client.Delete(child))
	expectFailed(client.Delete(child))
	expectE(client.Create(child, obj1))
	expectE(client.Update(child, objNoWrite))
	expectFailed(client.Update(child, obj2))
	expectE(client.Write(child, attachment))
	expect(E{Body: string(attachment)})(client.Read(child))
	success = true
	return
}
