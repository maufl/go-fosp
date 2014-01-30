package main

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
