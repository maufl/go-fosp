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

var dbLog = logging.MustGetLogger("go-fosp/fosp/database")

// Database is the database abtraction layer used by Server.
// It is mostly concered with access control and generating notifications.
// To access the actual database it uses a DatabaseDriver.
type Database struct {
	driver DatabaseDriver
	server *Server
}

var allRights = []string{"data-read", "data-write", "acl-read", "acl-write", "subscriptions-read", "subscriptions-write", "attachment-read", "attachment-write", "children-read", "children-write", "children-delete"}
var groupsPath = []string{"config", "groups"}

// NewDatabase creates a new Database struct and intializes the databaseDriver and server field.
func NewDatabase(driver DatabaseDriver, srv *Server) *Database {
	if driver == nil || srv == nil {
		panic("Cannot initialize database without server or driver")
	}
	db := new(Database)
	db.driver = driver
	db.server = srv
	return db
}

// Authenticate determins wether a user-password pair is valid or not.
// Returns nil on success and an error otherwise.
func (d *Database) Authenticate(user, password string) error {
	return d.driver.Authenticate(user, password)
}

// Register creates a new user in the Database.
// Returns nil on success and an error otherwise.
func (d *Database) Register(user, password string) error {
	if err := d.driver.Register(user, password); err != nil {
		return err
	}
	obj := new(Object)
	obj.Btime = time.Now().UTC()
	obj.Mtime = time.Now().UTC()
	obj.Owner = user + "@" + d.server.Domain()
	obj.Acl = &AccessControlList{Users: map[string][]string{user + "@" + d.server.Domain(): allRights}, Owner: allRights}
	obj.Data = "Foo"
	err := d.driver.CreateNode(&URL{user: user, domain: d.server.Domain()}, obj)
	return err
}

// Select returns the object for the given url.
func (d *Database) Select(user string, url *URL) (Object, error) {
	object, err := d.driver.GetNodeWithParents(url)
	if err != nil {
		return Object{}, err
	}
	dbLog.Debug("Selected object is %v", object.Acl)
	rights := d.userRights(user, &object)
	if !d.isUserAuthorized(user, &object, []string{"data-read"}) {
		return Object{}, ErrNotAuthorized
	}
	if !contains(rights, "acl-read") {
		object.Acl = nil
	}
	if !contains(rights, "subscriptions-read") {
		object.Subscriptions = nil
	}
	return object, nil
}

// Create saves a new object at the given url.
func (d *Database) Create(user string, url *URL, o *Object) error {
	if url.IsRoot() {
		return ErrInvalidRequest
	}
	parent, err := d.driver.GetNodeWithParents(url.Parent())
	if err != nil {
		return err
	}
	dbLog.Debug("Parent of to be created object is %v", parent)
	if !d.isUserAuthorized(user, &parent, []string{"children-write"}) {
		return ErrNotAuthorized
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

// Update merges changes into the object at the given url.
func (d *Database) Update(user string, url *URL, o *UnsaveObject) error {
	obj, err := d.driver.GetNodeWithParents(url)
	if err != nil {
		return err
	}
	rights := make([]string, 0)
	if o.Acl != nil && !o.Acl.Empty() {
		rights = append(rights, "acl-write")
	}
	if len(o.Subscriptions) != 0 {
		rights = append(rights, "subscriptions-write")
	}
	if o.Data != nil {
		rights = append(rights, "data-write")
	}
	if !d.isUserAuthorized(user, &obj, rights) {
		return ErrNotAuthorized
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

// List returns all child objects for the given url.
func (d *Database) List(user string, url *URL) ([]string, error) {
	obj, err := d.driver.GetNodeWithParents(url)
	if err != nil {
		return []string{}, err
	}
	if !d.isUserAuthorized(user, &obj, []string{"children-read"}) {
		return []string{}, ErrNotAuthorized
	}
	list, err := d.driver.ListNodes(url)
	if err != nil {
		return []string{}, err
	}
	return list, nil
}

// Delete removes the object for the given url.
func (d *Database) Delete(user string, url *URL) error {
	if url.IsRoot() {
		return ErrInvalidRequest
	}
	object, err := d.driver.GetNodeWithParents(url)
	if err != nil {
		return err
	}
	if !d.isUserAuthorized(user, object.Parent, []string{"children-delete"}) {
		return ErrNotAuthorized
	}
	err = d.driver.DeleteNodes(url)
	if err == nil {
		go d.notify(Deleted, object)
	}
	return err
}

// Read returns the attached file for the given url.
func (d *Database) Read(user string, url *URL) ([]byte, error) {
	object, err := d.driver.GetNodeWithParents(url)
	if err != nil {
		return []byte{}, err
	}
	if !d.isUserAuthorized(user, &object, []string{"attachment-read"}) {
		return []byte{}, ErrNotAuthorized
	}
	return d.driver.ReadAttachment(url)
}

// Write saves a file attachment at the givn url.
func (d *Database) Write(user string, url *URL, data []byte) error {
	object, err := d.driver.GetNodeWithParents(url)
	if err != nil {
		return err
	}
	if !d.isUserAuthorized(user, &object, []string{"attachment-write"}) {
		return ErrNotAuthorized
	}
	return d.driver.WriteAttachment(url, data)
}

func (d *Database) getGroups(url *URL) map[string][]string {
	groupsURL := &URL{url.UserName(), url.Domain(), groupsPath}
	object, err := d.driver.GetNodeWithParents(groupsURL)
	if err != nil {
		return make(map[string][]string)
	}
	if groups, ok := object.Data.(map[string][]string); ok {
		return groups
	}
	return make(map[string][]string)
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

func (d *Database) isUserAuthorized(user string, object *Object, rights []string) bool {
	dbLog.Debug("Authorizing user %s on object %s for rights %v", user, object.URL, rights)
	groups := groupsForUser(user, d.getGroups(object.URL))
	acl := object.AugmentedACL()
	dbLog.Debug("Augmented ACL is %v", acl)
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

func (d *Database) userRights(user string, object *Object) []string {
	rights := []string{}
	groups := groupsForUser(user, d.getGroups(object.URL))
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
