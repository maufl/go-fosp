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
	"errors"
	"net/url"
	"time"
)

// Object represents a FOSP object.
type Object struct {
	Parent        *Object                       `json:"-"`
	URL           *url.URL                      `json:"-"`
	Btime         time.Time                     `json:"btime,omitempty"`
	Mtime         time.Time                     `json:"mtime,omitempty"`
	Owner         string                        `json:"owner,omitempty"`
	Acl           *AccessControlList            `json:"acl,omitempty"`
	Subscriptions map[string]*SubscriptionEntry `json:"subscriptions,omitempty"`
	Attachment    *Attachment                   `json:"attachment,omitempty"`
	Type          interface{}                   `json:"type,omitempty"`
	Data          interface{}                   `json:"data,omitempty"`
}

func NewObject() *Object {
	return &Object{
		Subscriptions: make(map[string]*SubscriptionEntry),
	}
}

func (o *Object) Patch(patch PatchObject) error {
	if newType, ok := patch["type"]; ok {
		o.Type = newType
	}
	if data, ok := patch["data"]; ok {
		if oldData, ok := o.Data.(map[string]interface{}); ok {
			if newData, ok := data.(map[string]interface{}); ok {
				o.Data = recursiveMerge(oldData, newData)
			} else {
				o.Data = newData
			}
		} else {
			o.Data = data
		}
	}
	if o.Acl == nil {
		o.Acl = NewAccessControlList()
	}
	if err := patch.PatchStruct(o.Acl, "acl"); err != nil {
		return err
	}
	if tmp, ok := patch["subscriptions"]; ok {
		if newSubscriptions, ok := tmp.(map[string]interface{}); ok {
			for user, subscription := range newSubscriptions {
				if subscriptionPatch, ok := subscription.(PatchObject); ok {
					if _, ok := o.Subscriptions[user]; !ok {
						o.Subscriptions[user] = NewSubscriptionEntry()
					}
					o.Subscriptions[user].Patch(subscriptionPatch)
				} else {
					return errors.New("Subscription field for " + user + " does not contain an object")
				}
			}
		} else {
			return errors.New("Subscription field does not contain an object")
		}
	}
	if o.Attachment == nil {
		o.Attachment = NewAttachment()
	}
	if err := patch.PatchStruct(o.Attachment, "attachment"); err != nil {
		return err
	}
	return nil
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

func (o *Object) ReducedACL() *AccessControlList {
	if o.Parent != nil {
		return o.Parent.ReducedACL().OverwriteWith(o.Acl)
	}
	if o.Acl == nil {
		return NewAccessControlList()
	}
	return o.Acl
}

func (o *Object) PermissionsForData(user string, groups ...string) *PermissionSet {
	acl := o.ReducedACL()
	perms := NewPermissionSet()
	if acl.Others != nil {
		perms = perms.OverwriteWith(acl.Others.Data)
	}
	for _, group := range groups {
		if _, ok := acl.Groups[group]; ok {
			perms = perms.OverwriteWith(acl.Groups[group].Data)
		}
	}
	if _, ok := acl.Users[user]; ok {
		perms = perms.OverwriteWith(acl.Users[user].Data)
	}
	if o.Owner == user && acl.Owner != nil {
		perms = perms.OverwriteWith(acl.Owner.Data)
	}
	return perms
}

func (o *Object) PermissionsForAcl(user string, groups ...string) *PermissionSet {
	acl := o.ReducedACL()
	perms := NewPermissionSet()
	if acl.Others != nil {
		perms = perms.OverwriteWith(acl.Others.Acl)
	}
	for _, group := range groups {
		if _, ok := acl.Groups[group]; ok {
			perms = perms.OverwriteWith(acl.Groups[group].Acl)
		}
	}
	if _, ok := acl.Users[user]; ok {
		perms = perms.OverwriteWith(acl.Users[user].Acl)
	}
	if o.Owner == user && acl.Owner != nil {
		perms = perms.OverwriteWith(acl.Owner.Acl)
	}
	return perms
}

func (o *Object) PermissionsForSubscriptions(user string, groups ...string) *PermissionSet {
	acl := o.ReducedACL()
	perms := NewPermissionSet()
	if acl.Others != nil {
		perms = perms.OverwriteWith(acl.Others.Subscriptions)
	}
	for _, group := range groups {
		if _, ok := acl.Groups[group]; ok {
			perms = perms.OverwriteWith(acl.Groups[group].Subscriptions)
		}
	}
	if _, ok := acl.Users[user]; ok {
		perms = perms.OverwriteWith(acl.Users[user].Subscriptions)
	}
	if o.Owner == user && acl.Owner != nil {
		perms = perms.OverwriteWith(acl.Owner.Subscriptions)
	}
	return perms
}

func (o *Object) PermissionsForChildren(user string, groups ...string) *PermissionSet {
	acl := o.ReducedACL()
	perms := NewPermissionSet()
	if acl.Others != nil {
		perms = perms.OverwriteWith(acl.Others.Children)
	}
	for _, group := range groups {
		if _, ok := acl.Groups[group]; ok {
			perms = perms.OverwriteWith(acl.Groups[group].Children)
		}
	}
	if _, ok := acl.Users[user]; ok {
		perms = perms.OverwriteWith(acl.Users[user].Children)
	}
	if o.Owner == user && acl.Owner != nil {
		perms = perms.OverwriteWith(acl.Owner.Children)
	}
	return perms
}
