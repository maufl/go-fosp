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

// AccessControlList represents the acl content of an Object.
type AccessControlList struct {
	Owner  []string            `json:"owner,omitempty"`
	Users  map[string][]string `json:"users,omitempty"`
	Groups map[string][]string `json:"groups,omitempty"`
	Others []string            `json:"others,omitempty"`
}

// NewAccessControlList creates a new AccessControlList and initializes fields to non-nil values.
func NewAccessControlList() *AccessControlList {
	return &AccessControlList{make([]string, 0), make(map[string][]string), make(map[string][]string), make([]string, 0)}
}

// Clone creates a copy of the AccessControlList.
func (a *AccessControlList) Clone() *AccessControlList {
	acl := NewAccessControlList()
	acl.Owner = append(acl.Owner, a.Owner...)
	acl.Others = append(acl.Others, a.Others...)
	for user, rights := range a.Users {
		acl.Users[user] = rights
	}
	for group, rights := range a.Groups {
		acl.Groups[group] = rights
	}
	return acl
}

// Empty returns true if this ACL does not contain rights for the owner, others, groups or users
func (a *AccessControlList) Empty() bool {
	if len(a.Owner)+len(a.Users)+len(a.Groups)+len(a.Others) == 0 {
		return true
	}
	return false
}
