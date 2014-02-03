package main

import (
	"log"
	"time"
)

type database struct {
	driver DatabaseDriver
	server *server
}

var allRights = []string{"data-read", "data-write", "acl-read", "acl-write", "subscriptions-read", "subscriptions-write", "children-read", "children-write", "children-delete"}

func NewDatabase(driver DatabaseDriver, srv *server) *database {
	if driver == nil || srv == nil {
		panic("Cannot initialize database without server or driver")
	}
	db := new(database)
	db.driver = driver
	db.server = srv
	return db
}

func (d *database) Authenticate(user, password string) error {
	return d.driver.authenticate(user, password)
}

func (d *database) Register(user, password string) error {
	if err := d.driver.register(user, password); err != nil {
		return err
	}
	obj := new(Object)
	obj.Btime = time.Now().UTC()
	obj.Mtime = time.Now().UTC()
	obj.Owner = user + "@" + d.server.Domain()
	obj.Acl = map[string][]string{user + "@" + d.server.Domain(): allRights}
	obj.Data = "Foo"
	err := d.driver.createNode(&Url{user: user, domain: d.server.Domain()}, obj)
	return err
}

func (d *database) Select(user string, url *Url) (Object, error) {
	object, err := d.driver.getNodeWithParents(url)
	rights := object.UserRights(user)
	if err != nil {
		return Object{}, err
	}
	if !contains(rights, "data-read") {
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
	parent, err := d.driver.getNodeWithParents(url.Parent())
	if err != nil {
		return err
	}
	rights := parent.UserRights(user)
	log.Println("User rights %v+", rights)
	if !contains(rights, "children-write") {
		return NotAuthorizedError
	}

	o.Mtime = time.Now().UTC()
	o.Btime = time.Now().UTC()
	o.Owner = user
	err = d.driver.createNode(url, o)
	if err == nil {
		if object, err := d.driver.getNodeWithParents(url); err == nil {
			go d.notify(Created, object)
		}
	}
	return err
}

func (d *database) Update(user string, url *Url, o *Object) error {
	obj, err := d.driver.getNodeWithParents(url)
	if err != nil {
		return err
	}
	rights := obj.UserRights(user)
	if len(o.Acl) != 0 && !contains(rights, "acl-write") {
		return NotAuthorizedError
	}
	if len(o.Subscriptions) != 0 && !contains(rights, "subscriptions-write") {
		return NotAuthorizedError
	}
	if o.Data != nil && !contains(rights, "data-write") {
		return NotAuthorizedError
	}
	obj.Merge(o)
	obj.Mtime = time.Now().UTC()
	err = d.driver.updateNode(url, &obj)
	if err == nil {
		if object, err := d.driver.getNodeWithParents(url); err == nil {
			go d.notify(Updated, object)
		}
	}
	return err
}

func (d *database) List(user string, url *Url) ([]string, error) {
	obj, err := d.driver.getNodeWithParents(url)
	if err != nil {
		return []string{}, err
	}
	rights := obj.UserRights(user)
	if !contains(rights, "children-read") {
		return []string{}, NotAuthorizedError
	}
	list, err := d.driver.listNodes(url)
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
	object, err := d.driver.getNodeWithParents(url)
	if err != nil {
		return err
	}
	rights := object.Parent.UserRights(user)
	if !contains(rights, "children-delete") {
		return NotAuthorizedError
	}
	err = d.driver.deleteNodes(url)
	if err == nil {
		go d.notify(Deleted, object)
	}
	return err
}

func (d *database) Read(user string, url *Url) ([]byte, error) {
	object, err := d.driver.getNodeWithParents(url)
	if err != nil {
		return []byte{}, err
	}
	rights := object.UserRights(user)
	if !contains(rights, "attachment-read") {
		return []byte{}, NotAuthorizedError
	}
	return d.driver.readAttachment(url)
}

func (d *database) Write(user string, url *Url, data []byte) error {
	object, err := d.driver.getNodeWithParents(url)
	if err != nil {
		return err
	}
	rights := object.UserRights(user)
	if !contains(rights, "attachment-write") {
		return NotAuthorizedError
	}
	return d.driver.writeAttachment(url, data)
}
