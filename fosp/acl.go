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

func (acl *AccessControlList) OverwriteWith(newACL *AccessControlList) *AccessControlList {
	result := *acl
	if newACL == nil {
		return &result
	}
	if result.Owner != nil {
		result.Owner = result.Owner.OverwriteWith(newACL.Owner)
	} else {
		result.Owner = newACL.Owner
	}
	if result.Others != nil {
		result.Others = result.Others.OverwriteWith(newACL.Others)
	} else {
		result.Others = newACL.Others
	}
	for user, _ := range result.Users {
		if newACE, ok := newACL.Users[user]; ok {
			result.Users[user] = result.Users[user].OverwriteWith(newACE)
		}
	}
	for user, newACE := range newACL.Users {
		if _, ok := result.Users[user]; !ok {
			result.Users[user] = newACE
		}
	}
	for group, _ := range result.Groups {
		if newACE, ok := newACL.Groups[group]; ok {
			result.Groups[group] = result.Groups[group].OverwriteWith(newACE)
		}
	}
	for group, newACE := range newACL.Groups {
		if _, ok := result.Groups[group]; !ok {
			result.Groups[group] = newACE
		}
	}
	return &result
}

func (acl *AccessControlList) Patch(patch PatchObject) {
	if tmp, ok := patch["owner"]; ok {
		if tmp == nil {
			acl.Owner = NewAccessControlEntry()
		} else if ownerPatch, ok := tmp.(PatchObject); ok {
			acl.Owner.Patch(ownerPatch)
		}
	}
	if tmp, ok := patch["users"]; ok {
		if tmp == nil {
			acl.Users = make(map[string]*AccessControlEntry, 0)
		} else if entries, ok := tmp.(map[string]interface{}); ok {
			for user, entry := range entries {
				if entry == nil {
					acl.Users[user] = nil
				} else if acePatch, ok := entry.(PatchObject); ok {
					acl.Users[user].Patch(acePatch)
				}
			}
		}
	}
	if tmp, ok := patch["groups"]; ok {
		if tmp == nil {
			acl.Groups = make(map[string]*AccessControlEntry, 0)
		} else if entries, ok := tmp.(map[string]interface{}); ok {
			for group, entry := range entries {
				if entry == nil {
					acl.Groups[group] = nil
				} else if acePatch, ok := entry.(PatchObject); ok {
					acl.Groups[group].Patch(acePatch)
				}
			}
		}
	}
	if tmp, ok := patch["others"]; ok {
		if othersPatch, ok := tmp.(PatchObject); ok {
			acl.Others.Patch(othersPatch)
		}
	}
}
