package fosp

import (
	_ "log"
)

func (d *database) notify(event Event, object Object) {
	//log.Printf("Event %s on object %s occured", event, object.Url)
	users := object.SubscribedUsers(event, 0)
	//log.Printf("Users %v should be notified", users)
	for _, user := range users {
		var notification *Notification
		if event != Deleted {
			ov := object.UserView(user)
			notification = NewNotification(event, object.Url, map[string]string{}, ov.String())
		} else {
			notification = NewNotification(event, object.Url, map[string]string{}, "")
		}
		d.server.routeNotification(user, notification)
	}
}
