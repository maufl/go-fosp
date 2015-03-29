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

package main

import (
	"github.com/maufl/go-fosp/fosp"
)

func (d *Database) notify(event fosp.Event, object fosp.Object) {
	dbLog.Debug("Event %s on object %s occured", event, object.URL)
	users := object.SubscribedUsers(event, 0)
	dbLog.Debug("Users %v should be notified", users)
	for _, user := range users {
		var notification *fosp.Notification
		if event != fosp.Deleted {
			ov := object.UserView(user)
			notification = fosp.NewNotification(event, object.URL, map[string]string{}, ov.String())
		} else {
			notification = fosp.NewNotification(event, object.URL, map[string]string{}, "")
		}
		d.server.routeNotification(user, notification)
	}
}
