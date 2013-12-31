package main

import (
  "time"
)

type Database struct {
  driver *PostgresqlDriver
  server *server
}

func (d* Database) Authenticate(user, password string) error {
  return d.driver.authenticate(user, password)
}

func (d* Database) Register(user, password string) error {
  if err := d.driver.register(user, password); err != nil {
    return err
  }
  obj := new(Object)
  obj.Btime = time.Now().UTC()
  obj.Mtime = time.Now().UTC()
  obj.Owner = user + "@" + d.server.GetDomain()
  obj.Acl = map[string][]string{ user + "@" + d.server.GetDomain(): []string{"data-read", "data-write", "acl-read", "acl-write", "subscriptions-read", "subscriptions-write", "children-read", "children-write", "children-list"} }
  obj.Data = "Foo"
  err := d.driver.createNode(&Url{ user: user, domain: d.server.GetDomain() }, obj)
  return err
}

func (d *Database) Select(user string, url *Url) (Object, error) {
  object, err := d.driver.getNodeWithParents(url)
  rights := object.UserRights(user)
  if err != nil {
    return Object{}, err
  }
  if ! contains(rights, "data-read") {
    object.Data = nil
  }
  if ! contains(rights, "acl-read") {
    object.Acl = nil
  }
  if ! contains(rights, "subscriptions-read") {
    object.Subscriptions = nil
  }
  return object, nil
}

func (d *Database) Create(user string, url *Url, o *Object) error {
  if url.IsRoot() {
    return InvalidRequestError
  }
  parent, err := d.driver.getNodeWithParents(url.Parent())
  if err != nil {
    return err
  }
  rights := parent.UserRights(user)
  if ! contains(rights, "children-write") {
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

func (d *Database) Update(user string, url *Url, o *Object) error {
  obj, err := d.driver.getNodeWithParents(url)
  if err != nil {
    return err
  }
  rights := obj.UserRights(user)
  if len(o.Acl) != 0 && ! contains(rights, "acl-write") {
    return NotAuthorizedError
  }
  if len(o.Subscriptions) != 0 && ! contains(rights, "subscriptions-write") {
    return NotAuthorizedError
  }
  if o.Data != nil && ! contains(rights, "data-write") {
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

func (d *Database) List(user string, url *Url) ([]string, error) {
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

func (d *Database) Delete(user string, url *Url) (error) {
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
