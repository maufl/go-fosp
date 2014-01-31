package main

func (d *database) notify(event Event, object Object) {
	users := object.SubscribedUsers(event, 0)
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
