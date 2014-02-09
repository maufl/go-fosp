package fosp

import (
	"encoding/json"
	"log"
	"strings"
	"time"
)

type Object struct {
	Parent        *Object                      `json:"omit"`
	Url           *Url                         `json:"omit"`
	Btime         time.Time                    `json:"btime,omitempty"`
	Mtime         time.Time                    `json:"mtime,omitempty"`
	Owner         string                       `json:"owner,omitempty"`
	Acl           *AccessControlList          `json:"acl,omitempty"`
	Subscriptions map[string]SubscriptionEntry `json:"subscriptions,omitempty"`
	Attachment    *Attachment                  `json:"attachment,omitempty"`
	Data          interface{}                  `json:"data,omitempty"`
}

type AccessControlList struct {
	Owner []string `json:"owner,omitempty"`
	Users map[string][]string `json:"users,omitempty"`
	Groups map[string][]string `json:"groups,omitempty"`
	Others []string `json:"others,omitempty"`
}

type SubscriptionEntry struct {
	Depth  int      `json:"depth,omitempty"`
	Events []string `json:"events,omitempty"`
}

type Attachment struct {
	Name string `json:"name,omitempty"`
	Size uint   `json:"size,omitempty"`
	Type string `json:"type,omitempty"`
}

func (o *Object) Merge(src *Object) {
	if o.Acl == nil {
		o.Acl = new(AccessControlList)
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
		o.Attachment = src.Attachment
	}
	if src.Data != nil {
		o.Data = src.Data
	}
}

func (o *Object) String() string {
	if str, err := json.Marshal(o); err != nil {
		return ""
	} else {
		return string(str)
	}
}

func (o *Object) UserRights(user string) []string {
	rights := []string{}
	if r, ok := o.Acl.Users[user]; ok {
		rights = r
	}
	log.Println("Righst for user %s on this object are %v+", user, rights)
	if o.Parent != nil {
		pRights := o.Parent.UserRights(user)
		rights = overlayRights(rights, pRights)
	}
	return rights
}

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
	return ov
}

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

func Unmarshal(body string) (*Object, error) {
	var obj Object
	err := json.Unmarshal([]byte(body), &obj)
	if err != nil {
		return nil, err
	}
	if obj.Acl == nil {
		obj.Acl = new(AccessControlList)
	}
	if obj.Subscriptions == nil {
		obj.Subscriptions = make(map[string]SubscriptionEntry)
	}
	return &obj, nil
}
