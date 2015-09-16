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
	"code.google.com/p/go.crypto/bcrypt"
	"crypto/sha512"
	"database/sql"
	"encoding/base32"
	"encoding/json"
	// This import is needed to make the postgres driver available to database/sql.
	_ "github.com/lib/pq"
	"github.com/maufl/go-fosp/fosp"
	"github.com/op/go-logging"
	"io/ioutil"
	"net/url"
	"path"
	"strings"
)

var psqlLog = logging.MustGetLogger("go-fosp/fosp/postgresql-driver")

// PostgresqlDriver implements the database specific operations for storing the data in a Postgres database.
// PostgresqlDriver adheres to the DatabaseDriver interface and can be used by the Database object.
type PostgresqlDriver struct {
	db       *sql.DB
	basepath string
}

// NewPostgresqlDriver instanciates a new PostgresqlDriver for the given connectionString.
func NewPostgresqlDriver(connectionString, basePath string) *PostgresqlDriver {
	d := new(PostgresqlDriver)
	d.basepath = path.Clean(basePath)
	var err error
	d.db, err = sql.Open("postgres", connectionString)
	if err != nil {
		psqlLog.Fatal("Error occured when establishing db connection :: %s", err)
	}
	d.db.SetMaxOpenConns(50)
	return d
}

// Authenticate checks whether the name password pair is valid.
func (d *PostgresqlDriver) Authenticate(name, password string) bool {
	var passwordHash string
	err := d.db.QueryRow("SELECT password FROM users WHERE name = $1", name).Scan(&passwordHash)
	if err != nil {
		psqlLog.Error("Error when selecting record for authentication: %s", err)
		return false
	} else if err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		return false
	} else {
		return true
	}
}

// GetObjectWithParents returns an object and all it's parents from the database.
// The parents are stored recursively in the object.
func (d *PostgresqlDriver) GetObjectWithParents(url *url.URL) (fosp.Object, error) {
	urls := urlFamily(url)
	urlStrings := make([]string, len(urls))
	for i, u := range urls {
		urlStrings[i] = "'" + u.String() + "'"
	}
	psqlLog.Debug("Fetching objects for URLs %v from database", urlStrings)
	rows, err := d.db.Query("SELECT * FROM data WHERE uri IN (" + strings.Join(urlStrings, ",") + ") ORDER BY uri ASC")
	if err != nil {
		psqlLog.Error("Error when fetching object and parents from database: ", err)
		return fosp.Object{}, InternalServerError
	}
	defer rows.Close()
	var parent *fosp.Object
	var numObjects int
	for rows.Next() {
		var (
			id       uint64
			uri      string
			parentID uint64
			content  string
		)
		if err := rows.Scan(&id, &uri, &parentID, &content); err != nil {
			psqlLog.Error("Error when reading values from object row :: %s", err)
			return fosp.Object{}, InternalServerError
		}
		obj := fosp.NewObject()
		err := json.Unmarshal([]byte(content), obj)
		if err != nil {
			psqlLog.Critical("Error when unmarshaling json ::%T %s", err, err)
			return fosp.Object{}, InternalServerError
		}
		obj.URL, err = url.Parse(uri)
		if err != nil {
			psqlLog.Critical("Error while parsing URL %s :: %s", uri, err)
			return fosp.Object{}, InternalServerError
		}
		obj.Parent = parent
		parent = obj
		numObjects++
	}
	if numObjects != len(urls) {
		return fosp.Object{}, NewFospError("Object not found", fosp.StatusNotFound)
	}
	return *parent, nil
}

// CreateObject saves a new object to the database under the given URL.
func (d *PostgresqlDriver) CreateObject(url *url.URL, o *fosp.Object) error {
	var parentID uint64
	if path.Base(url.Path) != "/" {
		parentUrl := *url
		parentUrl.Path = path.Base(url.Path)
		err := d.db.QueryRow("SELECT id FROM data WHERE uri = $1", parentUrl.String()).Scan(&parentID)
		if err != nil {
			psqlLog.Error("Error when fetching parent for new object :: %s", err)
			return InternalServerError
		}
	}
	content, err := json.Marshal(o)
	if err != nil {
		psqlLog.Error("Error while marshaling object :: %s", err)
		return InternalServerError
	}
	_, err = d.db.Exec("INSERT INTO data (uri, parent_id, content) VALUES ($1, $2, $3)", url.String(), parentID, content)
	if err != nil {
		psqlLog.Error("Error when adding new object :: %s", err)
		return InternalServerError
	}
	return nil
}

// UpdateObject replaces the object at the given URL with a new object.
func (d *PostgresqlDriver) UpdateObject(url *url.URL, o *fosp.Object) error {
	content, err := json.Marshal(o)
	if err != nil {
		psqlLog.Error("Error while marshaling object :: %s", err)
		return InternalServerError
	}
	_, err = d.db.Exec("UPDATE data SET content = $1 WHERE uri = $2", content, url.String())
	if err != nil {
		psqlLog.Error("Error while updating object :: %s", err)
		return InternalServerError
	}
	return nil
}

// ListObjects returns an array of child object names of the object at the given URL.
func (d *PostgresqlDriver) ListObjects(url *url.URL) ([]string, error) {
	var parentID uint64
	err := d.db.QueryRow("SELECT id FROM data WHERE uri = $1", url.String()).Scan(&parentID)
	if err != nil {
		psqlLog.Error("Error while fetching object %s :: %s", url, err)
		return nil, InternalServerError
	}
	var rows *sql.Rows
	rows, err = d.db.Query("SELECT uri FROM data WHERE parent_id = $1", parentID)
	defer rows.Close()
	if err != nil {
		psqlLog.Error("Error while fetching children of %s :: %s", url, err)
		return nil, InternalServerError
	}
	uris := make([]string, 0, 25)
	for rows.Next() {
		var uri string
		if err := rows.Scan(&uri); err != nil {
			psqlLog.Error("Error when reading row :: %s", err)
			return nil, InternalServerError
		}
		u, err := url.Parse(uri)
		if err != nil {
			psqlLog.Error("Error while parsing URL %s :: %s", uri, err)
			return nil, InternalServerError
		}
		uris = append(uris, path.Base(u.Path))
	}
	return uris, nil
}

// DeleteObjects deletes the object at the given URL and all its children.
func (d *PostgresqlDriver) DeleteObjects(url *url.URL) error {
	_, err := d.db.Exec("DELETE FROM data WHERE uri ~ $1", "^"+url.String())
	if err != nil {
		psqlLog.Error("Error while deleting recorde for URL %s :: %s", url, err)
		return InternalServerError
	}
	return nil
}

// ReadAttachment returns the content of the attached file of the object at the given URL.
func (d *PostgresqlDriver) ReadAttachment(url *url.URL) ([]byte, error) {
	hash := sha512.Sum512([]byte(url.String()))
	filename := base32.StdEncoding.EncodeToString(hash[:sha512.Size])
	path := d.basepath + "/" + filename
	return ioutil.ReadFile(path)
}

// WriteAttachment stores the data as the attachment of the object at the given URL.
func (d *PostgresqlDriver) WriteAttachment(url *url.URL, data []byte) error {
	hash := sha512.Sum512([]byte(url.String()))
	filename := base32.StdEncoding.EncodeToString(hash[:sha512.Size])
	path := d.basepath + "/" + filename
	return ioutil.WriteFile(path, data, 0660)
}
