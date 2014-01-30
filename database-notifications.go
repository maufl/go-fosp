package main

func (d *database) notify(event Event, object Object) {
	users := object.SubscribedUsers(event, 0)
	for _, user := range users {
		var notification *Notification
		if event != Deleted {
			ov := object.UserView(user)
			notification = &Notification{event: event, url: object.Url, body: ov.String()}
		} else {
			notification = &Notification{event: event, url: object.Url}
		}
		d.server.routeNotification(user, notification)
	}
}
