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

func NegatePermission(perm string) string {
	switch perm {
	case PermissionRead:
		return PermissionNotRead
	case PermissionNotRead:
		return PermissionRead
	case PermissionWrite:
		return PermissionNotWrite
	case PermissionNotWrite:
		return PermissionWrite
	case PermissionDelete:
		return PermissionNotDelete
	case PermissionNotDelete:
		return PermissionDelete
	default:
		panic("Unknown permission " + perm)
	}
}

func IsPermission(perm string) bool {
	switch perm {
	case PermissionRead, PermissionNotRead, PermissionWrite, PermissionNotWrite, PermissionDelete, PermissionNotDelete:
		return true
	default:
		return false
	}
}

func NewPermissionSet(permissions ...string) *PermissionSet {
	ps := &PermissionSet{set: []string{}}
	for _, perm := range permissions {
		ps.Add(perm)
	}
	return ps
}

func (ps *PermissionSet) Add(newPerm string) bool {
	if !IsPermission(newPerm) {
		return false
	}
	neg := NegatePermission(newPerm)
	for i, oldPerm := range ps.set {
		if oldPerm == newPerm {
			return false
		}
		if oldPerm == neg {
			ps.set = append(ps.set[:i], ps.set[i+1:]...)
		}
	}
	ps.set = append(ps.set, newPerm)
	return true
}

func (ps *PermissionSet) Contains(needle string) bool {
	for _, perm := range ps.set {
		if needle == perm {
			return true
		}
	}
	return false
}

func (ps *PermissionSet) Size() int {
	return len(ps.set)
}

func (ps *PermissionSet) All() []string {
	return ps.set
}

func (ps *PermissionSet) OverwriteWith(newPS *PermissionSet) *PermissionSet {
	result := *ps
	if newPS == nil {
		return &result
	}
	for _, permission := range newPS.All() {
		result.Add(permission)
	}
	return &result
}

func (ps *PermissionSet) MarshalJSON() ([]byte, error) {
	bytes, err := json.Marshal(ps.set)
	return bytes, err
}

func (ps *PermissionSet) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &ps.set)
}
