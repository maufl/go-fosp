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

type AccessControlEntry struct {
	Data          *PermissionSet `json:"data,omitempty"`
	Acl           *PermissionSet `json:"acl,omitempty"`
	Subscriptions *PermissionSet `json:"subscriptions,omitempty"`
	Children      *PermissionSet `json:"children,omitempty"`
}

func NewAccessControlEntry() *AccessControlEntry {
	return &AccessControlEntry{
		Data:          NewPermissionSet(),
		Acl:           NewPermissionSet(),
		Subscriptions: NewPermissionSet(),
		Children:      NewPermissionSet(),
	}
}

func (ace *AccessControlEntry) OverwriteWith(newACE *AccessControlEntry) *AccessControlEntry {
	result := *ace
	if newACE == nil {
		return &result
	}
	if result.Data != nil {
		result.Data = result.Data.OverwriteWith(newACE.Data)
	} else {
		result.Data = newACE.Data
	}
	if result.Acl != nil {
		result.Acl = result.Acl.OverwriteWith(newACE.Acl)
	} else {
		result.Acl = newACE.Acl
	}
	if result.Subscriptions != nil {
		result.Subscriptions = result.Subscriptions.OverwriteWith(newACE.Subscriptions)
	} else {
		result.Subscriptions = newACE.Subscriptions
	}
	if result.Children != nil {
		result.Children = result.Children.OverwriteWith(newACE.Children)
	} else {
		result.Children = newACE.Children
	}
	return &result
}
