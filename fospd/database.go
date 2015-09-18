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
	"github.com/maufl/go-fosp/fosp"
	"github.com/op/go-logging"
	"io"
	"net/url"
	"path"
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
func (d *Database) Authenticate(user, password string) bool {
	return d.driver.Authenticate(user, password)
}

// Get returns the object for the given url.
func (d *Database) Get(user string, url *url.URL) (fosp.Object, error) {
	object, err := d.driver.GetObjectWithParents(url)
	if err != nil {
		return fosp.Object{}, err
	}
	missingPermissions := 0
	if !object.PermissionsForData(user).Contain(fosp.PermissionRead) {
		object.Data = nil
		object.Type = nil
		missingPermissions += 1
	}
	if !object.PermissionsForAcl(user).Contain(fosp.PermissionRead) {
		object.Acl = nil
		missingPermissions += 1
	}
	if !object.PermissionsForSubscriptions(user).Contain(fosp.PermissionRead) {
		object.Subscriptions = nil
		missingPermissions += 1
	}
	if missingPermissions == 3 {
		return fosp.Object{}, NewFospError("Insufficent rights", fosp.StatusForbidden)
	}
	dbLog.Debug("Selected object is %v", object)
	return object, nil
}

// Create saves a new object at the given url.
func (d *Database) Create(user string, url *url.URL, o *fosp.Object) error {
	if url.Path == "/" {
		return BadRequest
	}
	parentUrl := *url
	parentUrl.Path = path.Dir(url.Path)
	parent, err := d.driver.GetObjectWithParents(&parentUrl)
	if err != nil {
		dbLog.Warning("Could not get parent %s for new object %s", parentUrl, url)
		return err
	}
	dbLog.Debug("Parent of to be created object is %v", parent)

	o.Mtime = time.Now().UTC()
	o.Btime = time.Now().UTC()
	o.Owner = user
	err = d.driver.CreateObject(url, o)
	if err == nil {
		if object, err := d.driver.GetObjectWithParents(url); err == nil {
			go d.notify(fosp.CREATED, &object)
		}
	}
	return err
}

// Update merges changes into the object at the given url.
func (d *Database) Patch(user string, url *url.URL, patch fosp.PatchObject) error {
	obj, err := d.driver.GetObjectWithParents(url)
	if err != nil {
		return err
	}
	dbLog.Debug("Before patching, object is %#v", obj)
	if err := obj.Patch(patch); err != nil {
		return err
	}
	dbLog.Debug("Patched object is now %#v", obj)
	obj.Mtime = time.Now().UTC()
	err = d.driver.UpdateObject(url, &obj)
	if err == nil {
		if object, err := d.driver.GetObjectWithParents(url); err == nil {
			go d.notify(fosp.UPDATED, &object)
		}
	}
	return err
}

// List returns all child objects for the given url.
func (d *Database) List(user string, url *url.URL) ([]string, error) {
	list, err := d.driver.ListObjects(url)
	if err != nil {
		return []string{}, err
	}
	return list, nil
}

// Delete removes the object for the given url.
func (d *Database) Delete(user string, url *url.URL) error {
	if path.Base(url.Path) == "/" {
		return BadRequest
	}
	obj, err := d.driver.GetObjectWithParents(url)
	if err != nil {
		return err
	}
	err = d.driver.DeleteObjects(url)
	if err == nil {
		go d.notify(fosp.DELETED, &obj)
	}
	return err
}

// Read returns the attached file for the given url.
func (d *Database) Read(user string, url *url.URL) ([]byte, error) {
	return d.driver.ReadAttachment(url)
}

// Write saves a file attachment at the givn url.
func (d *Database) Write(user string, url *url.URL, data io.Reader) error {
	object, err := d.driver.GetObjectWithParents(url)
	if err != nil {
		return err
	}
	bytesWritten, err := d.driver.WriteAttachment(url, data)
	if err != nil {
		return err
	}
	if object.Attachment == nil {
		object.Attachment = fosp.NewAttachment()
	}
	object.Attachment.Size = uint(bytesWritten)
	object.Mtime = time.Now().UTC()
	d.driver.UpdateObject(url, &object)
	return nil
}
