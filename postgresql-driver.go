package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	_ "github.com/lib/pq"
	"log"
	"strings"
	"io/ioutil"
	"path"
	"crypto/sha256"
)

type postgresqlDriver struct {
	db *sql.DB
	basepath string
}

func psqlError(err error) {
	println("Database error occured: " + err.Error())
}

func NewPostgresqlDriver(connectionString, basePath string) *postgresqlDriver {
	d := new(postgresqlDriver)
	d.basepath = path.Clean(basePath)
	var err error
	d.db, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal("Error occured when establishing db connection :: ", err)
	}
	d.db.SetMaxOpenConns(50)
	return d
}

func (d *postgresqlDriver) authenticate(name, password string) error {
	var pw string
	err := d.db.QueryRow("SELECT password FROM users WHERE name = $1", name).Scan(&pw)
	if err != nil {
		psqlError(err)
		return err
	} else if pw == password {
		return nil
	} else {
		println(pw + " != " + password)
		return errors.New("Password did not match")
	}
}

func (d *postgresqlDriver) register(name, password string) error {
	_, err := d.db.Exec("INSERT INTO users (name, password) VALUES ($1, $2)", name, password)
	if err != nil {
		psqlError(err)
		return InternalServerError
	} else {
		return nil
	}
}

func (d *postgresqlDriver) getNodeWithParents(url *Url) (Object, error) {
	urls := make([]string, 0, len(url.path))
	for !url.IsRoot() {
		urls = append(urls, `'`+url.String()+`'`)
		url = url.Parent()
	}
	urls = append(urls, `'`+url.String()+`'`)

	rows, err := d.db.Query("SELECT * FROM data WHERE uri IN (" + strings.Join(urls, ",") + ") ORDER BY uri ASC")
	if err != nil {
		psqlError(err)
		return Object{}, InternalServerError
	}
	defer rows.Close()
	var parent *Object
	for rows.Next() {
		var (
			id        uint64
			uri       string
			parent_id uint64
			content   string
		)
		if err := rows.Scan(&id, &uri, &parent_id, &content); err != nil {
			psqlError(err)
			return Object{}, errors.New("Internal database error")
		}
		obj, err := Unmarshal(content)
		if err != nil {
			log.Println("Error when unmarshaling json :: ", err)
			return Object{}, errors.New("Internal database error")
		}
		obj.Url, err = parseUrl(uri)
		obj.Parent = parent
		parent = obj
	}
	return *parent, nil
}

func (d *postgresqlDriver) createNode(url *Url, o *Object) error {
	var parent_id uint64
	if !url.IsRoot() {
		err := d.db.QueryRow("SELECT id FROM data WHERE uri = $1", url.Parent().String()).Scan(&parent_id)
		if err != nil {
			psqlError(err)
			return err
		}
	}
	content, err := json.Marshal(o)
	if err != nil {
		return err
	}
	_, err = d.db.Exec("INSERT INTO data (uri, parent_id, content) VALUES ($1, $2, $3)", url.String(), parent_id, content)
	if err != nil {
		psqlError(err)
		return err
	}
	return nil
}

func (d *postgresqlDriver) updateNode(url *Url, o *Object) error {
	content, err := json.Marshal(o)
	if err != nil {
		return err
	}
	_, err = d.db.Exec("UPDATE data SET content = $1 WHERE uri = $2", content, url.String())
	if err != nil {
		return err
	}
	return nil
}

func (d *postgresqlDriver) listNodes(url *Url) ([]string, error) {
	var parent_id uint64
	err := d.db.QueryRow("SELECT id FROM data WHERE uri = $1", url.String()).Scan(&parent_id)
	if err != nil {
		return []string{}, err
	}
	var rows *sql.Rows
	rows, err = d.db.Query("SELECT uri FROM data WHERE parent_id = $1", parent_id)
	if err != nil {
		return []string{}, err
	}
	defer rows.Close()
	uris := make([]string, 0, 25)
	parent := url.String()
	if !url.IsRoot() {
		parent += "/"
	}
	for rows.Next() {
		var uri string
		if err := rows.Scan(&uri); err != nil {
			log.Println("Error when reading row :: ", err)
			return nil, errors.New("Internal database error")
		}
		uris = append(uris, strings.TrimPrefix(uri, parent))
	}
	return uris, nil
}

func (d *postgresqlDriver) deleteNodes(url *Url) error {
	_, err := d.db.Exec("DELETE FROM data WHERE uri ~ $1", "^"+url.String())
	return err
}

func (d *postgresqlDriver) readAttachment(url *Url) ([]byte, error) {
	hash := sha256.Sum224([]byte(url.Path()))
	filename := string(hash[:sha256.Size224])
	path := d.basepath + "/" + filename
	return ioutil.ReadFile(path)
}

func (d *postgresqlDriver) writeAttachment(url *Url, data []byte) error {
	hash := sha256.Sum224([]byte(url.Path()))
	filename := string(hash[:sha256.Size224])
	path := d.basepath + "/" + filename
	return ioutil.WriteFile(path, data, 0660)
}
