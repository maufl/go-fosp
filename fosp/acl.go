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

// AccessControlList represents the acl content of an Object.
type AccessControlList struct {
	Owner  *AccessControlEntry            `json:"owner,omitempty"`
	Users  map[string]*AccessControlEntry `json:"users,omitempty"`
	Groups map[string]*AccessControlEntry `json:"groups,omitempty"`
	Others *AccessControlEntry            `json:"others,omitempty"`
}

// NewAccessControlList creates a new AccessControlList and initializes fields to non-nil values.
func NewAccessControlList() *AccessControlList {
	return &AccessControlList{
		Owner:  NewAccessControlEntry(),
		Users:  make(map[string]*AccessControlEntry),
		Groups: make(map[string]*AccessControlEntry),
		Others: NewAccessControlEntry(),
	}
}
