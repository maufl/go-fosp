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

package main

import (
	"bytes"
	"encoding/json"
	"github.com/maufl/go-fosp/fosp"
	"strings"
)

func (d *Database) notify(event string, object *fosp.Object) {
	dbLog.Debug("Event %s on object %s occured", event, object.URL)
	users := subscribedUsers(object, event, 0)
	dbLog.Debug("Users %v should be notified", users)
	for _, user := range users {
		notification := fosp.NewNotification(event, object.URL)
		if event != fosp.DELETED {
			if serialized, err := json.Marshal(object); err == nil {
				notification.Body = bytes.NewBuffer(serialized)
			} else {
				dbLog.Error("Unable to serialize object %s for sending notification :: %s", object.URL, err)
				continue
			}
		}
		d.server.routeNotification(user, notification)
	}
}

func subscribedUsers(obj *fosp.Object, event string, depth int) (users []string) {
	if obj.Parent != nil {
		users = subscribedUsers(obj.Parent, event, depth+1)
	}
	for user, subscription := range obj.Subscriptions {
		if !contains(users, user) && (subscription.Depth == -1 || subscription.Depth >= depth) {
			for _, ev := range subscription.Events {
				if strings.EqualFold(ev, event) {
					users = append(users, user)
				}
			}
		}
	}
	return users
}
