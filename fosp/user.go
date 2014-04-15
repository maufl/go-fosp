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

package fosp

import (
	"errors"
	"strings"
)

// Error that is returned when an invalid user identifier is supplied to a parse method.
var ErrorMalformedUserIdentifier = errors.New("malformed user identifier")

// User represents an user identfier.
type User struct {
	name   string
	domain string
}

// NewUser creates a new User object.
func NewUser(name, domain string) *User {
	return &User{name, domain}
}

// Name returns the name of the User.
func (u *User) Name() string {
	return u.name
}

// Domain returns the domain of the User.
func (u *User) Domain() string {
	return u.domain
}

// UnmarshalJSON parses a User from its JSON representation.
func (u *User) UnmarshalJSON(data []byte) error {
	return u.parse(data)
}

func (u *User) parse(data []byte) error {
	userString := string(data)
	parts := strings.Split(userString, "@")
	if len(parts) != 2 {
		return ErrorMalformedUserIdentifier
	}
	u.name = parts[0]
	u.domain = strings.TrimSuffix(parts[1], ".")
	return nil
}

// ParseUser parses a User from its string representation, but accepts a byte array.
func ParseUser(data []byte) (*User, error) {
	u := &User{}
	return u, u.parse(data)
}

// ParseUserString parses a User from its string representation.
func ParseUserString(data string) (*User, error) {
	return ParseUser([]byte(data))
}
