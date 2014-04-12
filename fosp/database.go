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
	"github.com/op/go-logging"
	"strings"
	"time"
)

type database struct {
	driver DatabaseDriver
	server *server
	lg     *logging.Logger
}

var allRights = []string{"data-read", "data-write", "acl-read", "acl-write", "subscriptions-read", "subscriptions-write", "attachment-read", "attachment-write", "children-read", "children-write", "children-delete"}
var groupsPath = []string{"config", "groups"}

func NewDatabase(driver DatabaseDriver, srv *server) *database {
	if driver == nil || srv == nil {
		panic("Cannot initialize database without server or driver")
	}
	db := new(database)
	db.driver = driver
	db.server = srv
	db.lg = logging.MustGetLogger("go-fosp/fosp/database")
	return db
}

func (d *database) Authenticate(user, password string) error {
	return d.driver.Authenticate(user, password)
}

func (d *database) Register(user, password string) error {
	if err := d.driver.Register(user, password); err != nil {
		return err
	}
	obj := new(Object)
	obj.Btime = time.Now().UTC()
	obj.Mtime = time.Now().UTC()
	obj.Owner = user + "@" + d.server.Domain()
	obj.Acl = &AccessControlList{Users: map[string][]string{user + "@" + d.server.Domain(): allRights}, Owner: allRights}
	obj.Data = "Foo"
	err := d.driver.CreateNode(&Url{user: user, domain: d.server.Domain()}, obj)
	return err
}

func (d *database) Select(user string, url *Url) (Object, error) {
	object, err := d.driver.GetNodeWithParents(url)
	if err != nil {
		return Object{}, err
	}
	d.lg.Debug("Selected object is %v", object.Acl)
	rights := d.userRights(user, &object)
	if !d.isUserAuthorized(user, &object, []string{"data-read"}) {
		return Object{}, NotAuthorizedError
	}
	if !contains(rights, "acl-read") {
		object.Acl = nil
	}
	if !contains(rights, "subscriptions-read") {
		object.Subscriptions = nil
	}
	return object, nil
}

func (d *database) Create(user string, url *Url, o *Object) error {
	if url.IsRoot() {
		return InvalidRequestError
	}
	parent, err := d.driver.GetNodeWithParents(url.Parent())
	if err != nil {
		return err
	}
	d.lg.Debug("Parent of to be created object is %v", parent)
	if !d.isUserAuthorized(user, &parent, []string{"children-write"}) {
		return NotAuthorizedError
	}

	o.Mtime = time.Now().UTC()
	o.Btime = time.Now().UTC()
	o.Owner = user
	err = d.driver.CreateNode(url, o)
	if err == nil {
		if object, err := d.driver.GetNodeWithParents(url); err == nil {
			go d.notify(Created, object)
		}
	}
	return err
}

func (d *database) Update(user string, url *Url, o *Object) error {
	obj, err := d.driver.GetNodeWithParents(url)
	if err != nil {
		return err
	}
	rights := make([]string, 0)
	if o.Acl != nil {
		rights = append(rights, "acl-write")
	}
	if len(o.Subscriptions) != 0 {
		rights = append(rights, "subscriptions-write")
	}
	if o.Data != nil && !contains(rights, "data-write") {
		rights = append(rights, "data-write")
	}
	if !d.isUserAuthorized(user, &obj, rights) {
		return NotAuthorizedError
	}
	obj.Merge(o)
	obj.Mtime = time.Now().UTC()
	err = d.driver.UpdateNode(url, &obj)
	if err == nil {
		if object, err := d.driver.GetNodeWithParents(url); err == nil {
			go d.notify(Updated, object)
		}
	}
	return err
}

func (d *database) List(user string, url *Url) ([]string, error) {
	obj, err := d.driver.GetNodeWithParents(url)
	if err != nil {
		return []string{}, err
	}
	if !d.isUserAuthorized(user, &obj, []string{"children-read"}) {
		return []string{}, NotAuthorizedError
	}
	list, err := d.driver.ListNodes(url)
	if err != nil {
		return []string{}, err
	} else {
		return list, nil
	}
}

func (d *database) Delete(user string, url *Url) error {
	if url.IsRoot() {
		return InvalidRequestError
	}
	object, err := d.driver.GetNodeWithParents(url)
	if err != nil {
		return err
	}
	if !d.isUserAuthorized(user, object.Parent, []string{"children-delete"}) {
		return NotAuthorizedError
	}
	err = d.driver.DeleteNodes(url)
	if err == nil {
		go d.notify(Deleted, object)
	}
	return err
}

func (d *database) Read(user string, url *Url) ([]byte, error) {
	object, err := d.driver.GetNodeWithParents(url)
	if err != nil {
		return []byte{}, err
	}
	if !d.isUserAuthorized(user, &object, []string{"attachment-read"}) {
		return []byte{}, NotAuthorizedError
	}
	return d.driver.ReadAttachment(url)
}

func (d *database) Write(user string, url *Url, data []byte) error {
	object, err := d.driver.GetNodeWithParents(url)
	if err != nil {
		return err
	}
	if !d.isUserAuthorized(user, &object, []string{"attachment-write"}) {
		return NotAuthorizedError
	}
	return d.driver.WriteAttachment(url, data)
}

func (d *database) getGroups(url *Url) map[string][]string {
	groupsUrl := &Url{url.UserName(), url.Domain(), groupsPath}
	object, err := d.driver.GetNodeWithParents(groupsUrl)
	if err != nil {
		return make(map[string][]string)
	}
	if groups, ok := object.Data.(map[string][]string); ok {
		return groups
	} else {
		return make(map[string][]string)
	}
}

func groupsForUser(user string, groups map[string][]string) []string {
	grps := make([]string, 0)
	for group, users := range groups {
		if contains(users, user) {
			grps = append(grps, group)
		}
	}
	return grps
}

func (d *database) isUserAuthorized(user string, object *Object, rights []string) bool {
	d.lg.Debug("Authorizing user %s on object %s for rights %v", user, object.Url, rights)
	groups := groupsForUser(user, d.getGroups(object.Url))
	acl := object.AugmentedACL()
	d.lg.Debug("Augmented ACL is %v", acl)
	for _, right := range rights {
		if contains(acl.Others, right) {
			break
		}
		for _, group := range groups {
			if groupRights, ok := acl.Groups[group]; ok && contains(groupRights, right) {
				break
			}
		}
		if userRights, ok := acl.Users[user]; ok && contains(userRights, right) {
			break
		}
		if user == object.Owner && contains(acl.Owner, right) {
			break
		}
		return false
	}
	return true
}

func (d *database) userRights(user string, object *Object) []string {
	rights := []string{}
	groups := groupsForUser(user, d.getGroups(object.Url))
	acl := object.AugmentedACL()
	rights = accRights(rights, acl.Others)
	for _, group := range groups {
		if groupRights, ok := acl.Groups[group]; ok {
			rights = accRights(rights, groupRights)
		}
	}
	if userRights, ok := acl.Users[user]; ok {
		rights = accRights(rights, userRights)
	}
	if object.Owner == user {
		rights = accRights(rights, acl.Owner)
	}
	return rights
}

func accRights(acc, rights []string) []string {
	for _, right := range rights {
		if !(strings.HasPrefix(right, "not-") || contains(acc, right)) {
			acc = append(acc, right)
		}
	}
	return acc
}
