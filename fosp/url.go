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

// URL represents a FOSP URL in a message.
type URL struct {
	user   string
	domain string
	path   []string
}

func parseURL(s string) (*URL, error) {
	u := &URL{}
	atIndex := strings.Index(s, "@")
	if atIndex == -1 {
		return nil, errors.New("invalid url")
	}
	pathIndex := strings.Index(s, "/")
	u.user = s[:atIndex]
	if pathIndex != -1 {
		u.domain = s[atIndex+1 : pathIndex]
		path := s[pathIndex+1:]
		u.path = strings.Split(path, "/")
		if len(u.path) == 1 && u.path[0] == "" {
			u.path = []string{}
		}
	} else {
		u.domain = s[atIndex+1:]
	}
	return u, nil
}

func (u URL) String() string {
	if u.user == "" {
		return "*"
	}
	return u.user + "@" + u.domain + "/" + strings.Join(u.path, "/")
}

// Parent returns the parent URL or itself if it's a toplevel URL.
// The Parent URL is an URL with a path that is one segmet shorter.
func (u *URL) Parent() *URL {
	if u.IsRoot() {
		return u
	}
	p := &URL{user: u.user, domain: u.domain}
	p.path = u.path[:len(u.path)-1]
	return p
}

// IsRoot returns true if this URL is a toplevel URL.
func (u *URL) IsRoot() bool {
	if len(u.path) == 0 {
		return true
	}
	return false
}

// Domain returns the domain part of the URL.
func (u *URL) Domain() string {
	return u.domain
}

// UserName returns the username part of the URL.
func (u *URL) UserName() string {
	return u.user
}

// Path returns the string representation of the path part of the URL.
func (u *URL) Path() string {
	return "/" + strings.Join(u.path, "/")
}
