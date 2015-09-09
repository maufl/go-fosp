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

package fosp

import (
	"encoding/json"
)

const (
	PermissionRead      string = "read"
	PermissionNotRead          = "not-read"
	PermissionWrite            = "write"
	PermissionNotWrite         = "not-write"
	PermissionDelete           = "delete"
	PermissionNotDelete        = "not-delete"
)

type PermissionSet struct {
	set []string
}

func NewPermissionSet(elements ...string) *PermissionSet {
	set := []string{}
Loop:
	for _, e := range elements {
		for _, s := range set {
			if e == s {
				continue Loop
			}
		}
		set = append(set, e)
	}
	return &PermissionSet{set: set}
}

func (ps *PermissionSet) Add(e string) bool {
	for _, s := range ps.set {
		if s == e {
			return false
		}
	}
	ps.set = append(ps.set, e)
	return true
}

func (ps *PermissionSet) Contains(e string) bool {
	for _, s := range ps.set {
		if e == s {
			return true
		}
	}
	return false
}

func (ps *PermissionSet) Size() int {
	return len(ps.set)
}

func (ps *PermissionSet) MarshalJSON() ([]byte, error) {
	bytes, err := json.Marshal(ps.set)
	return bytes, err
}
