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

type Url struct {
	user   string
	domain string
	path   []string
}

func parseUrl(s string) (*Url, error) {
	u := &Url{}
	at_i := strings.Index(s, "@")
	if at_i == -1 {
		return &Url{}, errors.New("Invalid url")
	}
	p_i := strings.Index(s, "/")
	u.user = s[:at_i]
	if p_i != -1 {
		u.domain = s[at_i+1 : p_i]
		path := s[p_i+1:]
		u.path = strings.Split(path, "/")
		if len(u.path) == 1 && u.path[0] == "" {
			u.path = []string{}
		}
	} else {
		u.domain = s[at_i+1:]
	}
	return u, nil
}

func (u Url) String() string {
	if u.user == "" {
		return "*"
	}
	return u.user + "@" + u.domain + "/" + strings.Join(u.path, "/")
}

func (u *Url) Parent() *Url {
	if u.IsRoot() {
		return u
	}
	p := &Url{user: u.user, domain: u.domain}
	p.path = u.path[:len(u.path)-1]
	return p
}

func (u *Url) IsRoot() bool {
	if len(u.path) == 0 {
		return true
	}
	return false
}

func (u *Url) Domain() string {
	return u.domain
}

func (u *Url) UserName() string {
	return u.user
}

func (u *Url) Path() string {
	return "/" + strings.Join(u.path, "/")
}
