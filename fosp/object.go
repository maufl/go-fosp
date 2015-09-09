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
	"net/url"
	"time"
)

// Object represents a FOSP object.
type Object struct {
	Parent        *Object                      `json:"-"`
	URL           *url.URL                     `json:"-"`
	Btime         time.Time                    `json:",omitempty"`
	Mtime         time.Time                    `json:",omitempty"`
	Owner         string                       `json:",omitempty"`
	Acl           *AccessControlList           `json:",omitempty"`
	Subscriptions map[string]SubscriptionEntry `json:",omitempty"`
	Attachment    Attachment                   `json:",omitempty"`
	Type          interface{}                  `json:",omitempty"`
	Data          interface{}                  `json:",omitempty"`
}

func NewObject() *Object {
	return &Object{
		Acl:           NewAccessControlList(),
		Subscriptions: make(map[string]SubscriptionEntry),
	}
}

func (o *Object) Patch(patch PatchObject) {
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
	if _, ok := patch["acl"]; ok {
	}
	//TODO
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
