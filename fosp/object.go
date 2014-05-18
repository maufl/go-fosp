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
	"encoding/json"
	"strings"
	"time"
)

// Object represents a FOSP object.
type Object struct {
	Parent        *Object                      `json:"omit"`
	URL           *URL                         `json:"omit"`
	Btime         time.Time                    `json:"btime,omitempty"`
	Mtime         time.Time                    `json:"mtime,omitempty"`
	Owner         string                       `json:"owner,omitempty"`
	Acl           *AccessControlList           `json:"acl,omitempty"`
	Subscriptions map[string]SubscriptionEntry `json:"subscriptions,omitempty"`
	Attachment    *Attachment                  `json:"attachment,omitempty"`
	Data          interface{}                  `json:"data,omitempty"`
}

// UnsaveObject represents a FOSP object but it's entries can be nil!
type UnsaveObject struct {
	Btime         time.Time                    `json:"btime,omitempty"`
	Mtime         time.Time                    `json:"mtime,omitempty"`
	Owner         string                       `json:"owner,omitempty"`
	Acl           *AccessControlList           `json:"acl,omitempty"`
	Subscriptions map[string]SubscriptionEntry `json:"subscriptions,omitempty"`
	Attachment    *Attachment                  `json:"attachment,omitempty"`
	Data          interface{}                  `json:"data,omitempty"`
}

// SubscriptionEntry represents an entry in the subscriptions list of an object.
type SubscriptionEntry struct {
	Depth  int      `json:"depth,omitempty"`
	Events []string `json:"events,omitempty"`
}

// Merge updates an Object with values of another Object.
func (o *Object) Merge(src *UnsaveObject) {
	if o.Acl == nil {
		o.Acl = NewAccessControlList()
	}
	if src.Acl != nil {
		if src.Acl.Owner != nil {
			o.Acl.Owner = src.Acl.Owner
		}
		if src.Acl.Others != nil {
			o.Acl.Others = src.Acl.Others
		}
		for user, rights := range src.Acl.Users {
			o.Acl.Users[user] = rights
		}
		for group, rights := range src.Acl.Groups {
			o.Acl.Groups[group] = rights
		}
	}
	for user, subscription := range src.Subscriptions {
		o.Subscriptions[user] = subscription
	}
	if src.Attachment != nil {
		if src.Attachment.Type != "" {
			o.Attachment.Type = src.Attachment.Type
		}
		if src.Attachment.Name != "" {
			o.Attachment.Name = src.Attachment.Name
		}
	}
	if src.Data != nil {
		if left, ok := o.Data.(map[string]interface{}); ok {
			if right, ok := src.Data.(map[string]interface{}); ok {
				o.Data = recursiveMerge(left, right)
				return
			}
		}
		o.Data = src.Data
	}
}

func recursiveMerge(left, right map[string]interface{}) map[string]interface{} {
	for key, rightValue := range right {
		if leftValue, exists := left[key]; exists {
			if newLeft, ok := leftValue.(map[string]interface{}); ok {
				if newRight, ok := rightValue.(map[string]interface{}); ok {
					left[key] = recursiveMerge(newLeft, newRight)
					continue
				}
			}
		}
		if rightValue != nil {
			left[key] = rightValue
		} else {
			delete(left, key)
		}
	}
	return left
}

func (o *Object) String() string {
	if str, err := json.Marshal(o); err == nil {
		return string(str)
	}
	return ""
}

// Bytes returns the JSON representation of the object as byte array.
func (o *Object) Bytes() []byte {
	return []byte(o.String())
}

// UserRights extracts the user rights for one user from this Object.
func (o *Object) UserRights(user string) []string {
	rights := []string{}
	if r, ok := o.Acl.Users[user]; ok {
		rights = r
	}
	if o.Parent != nil {
		pRights := o.Parent.UserRights(user)
		rights = overlayRights(rights, pRights)
	}
	return rights
}

// AugmentedACL recursively combines the ACL from this Object with the ACL from its parent.
func (o *Object) AugmentedACL() *AccessControlList {
	acl := o.Acl.Clone()
	if o.Parent != nil {
		parentAcl := o.Parent.AugmentedACL()
		acl.Owner = overlayRights(acl.Owner, parentAcl.Owner)
		acl.Others = overlayRights(acl.Others, parentAcl.Others)
		for user, parentRights := range parentAcl.Users {
			if rights, ok := acl.Users[user]; ok {
				acl.Users[user] = overlayRights(rights, parentRights)
			} else {
				acl.Users[user] = parentRights
			}
		}
		for group, parentRights := range parentAcl.Groups {
			if rights, ok := acl.Groups[group]; ok {
				acl.Groups[group] = overlayRights(rights, parentRights)
			} else {
				acl.Groups[group] = parentRights
			}
		}
	}
	return acl
}

// UserView returns the Object as it can be seen with the rights of the given user.
func (o *Object) UserView(user string) Object {
	ov := Object{Owner: o.Owner, Btime: o.Btime, Mtime: o.Mtime}
	rights := o.UserRights(user)
	if contains(rights, "data-read") {
		ov.Data = o.Data
	}
	if contains(rights, "acl-read") {
		ov.Acl = o.Acl
	}
	if contains(rights, "subscriptions-read") {
		ov.Subscriptions = o.Subscriptions
	}
	if contains(rights, "attachment-read") {
		ov.Attachment = o.Attachment
	}
	return ov
}

// SubscribedUsers returns all users which have a subscriptions for the given Event on this Object.
func (o *Object) SubscribedUsers(event Event, depth int) []string {
	users := []string{}
	if o.Parent != nil {
		users = o.Parent.SubscribedUsers(event, depth+1)
	}
	for user, subscription := range o.Subscriptions {
		if !contains(users, user) && (subscription.Depth == -1 || subscription.Depth >= depth) {
			for _, ev := range subscription.Events {
				if strings.EqualFold(ev, event.String()) {
					users = append(users, user)
				}
			}
		}
	}
	return users
}

func overlayRights(bottom, top []string) []string {
	rights := []string{}
	rights = append(rights, bottom...)
	for _, t := range top {
		var positive, negative string
		if strings.HasPrefix(t, "not-") {
			positive = strings.TrimPrefix(t, "not-")
			negative = t
		} else {
			positive = t
			negative = "not-" + t
		}
		hit := false
		for _, b := range bottom {
			if b == negative || b == positive {
				hit = true
			}
		}
		if !hit {
			rights = append(rights, t)
		}
	}
	return rights
}

// UnmarshalObject parses an Object from its JSON representation.
func UnmarshalObject(body string) (*Object, error) {
	var obj Object
	err := json.Unmarshal([]byte(body), &obj)
	if err != nil {
		return nil, err
	}
	if obj.Acl == nil {
		obj.Acl = NewAccessControlList()
	}
	// FIXME(maufl): this is a HACK!
	// proper marshaling and unmarshaling should be implemented in the functions of the JSON Marshaler interface
	if obj.Acl.Owner == nil {
		obj.Acl.Owner = make([]string, 0)
	}
	if obj.Acl.Others == nil {
		obj.Acl.Others = make([]string, 0)
	}
	if obj.Acl.Users == nil {
		obj.Acl.Users = make(map[string][]string)
	}
	if obj.Acl.Groups == nil {
		obj.Acl.Groups = make(map[string][]string)
	}
	if obj.Subscriptions == nil {
		obj.Subscriptions = make(map[string]SubscriptionEntry)
	}
	if obj.Attachment == nil {
		obj.Attachment = NewAttachment()
	}
	return &obj, nil
}

// UnmarshalUnsaveObject parses an UnsaveObject from its JSON representation.
func UnmarshalUnsaveObject(data []byte) (*UnsaveObject, error) {
	obj := &UnsaveObject{}
	err := json.Unmarshal(data, &obj)
	return obj, err
}
