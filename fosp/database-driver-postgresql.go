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
	"code.google.com/p/go.crypto/bcrypt"
	"crypto/sha512"
	"database/sql"
	"encoding/base32"
	"encoding/json"
	"errors"
	// This import is needed to make the postgres driver available to database/sql.
	_ "github.com/lib/pq"
	"github.com/op/go-logging"
	"io/ioutil"
	"path"
	"strings"
)

type postgresqlDriver struct {
	lg       *logging.Logger
	db       *sql.DB
	basepath string
}

// NewPostgresqlDriver instanciates a new postgresqlDriver for the given connectionString.
func NewPostgresqlDriver(connectionString, basePath string) *postgresqlDriver {
	d := new(postgresqlDriver)
	d.basepath = path.Clean(basePath)
	d.lg = logging.MustGetLogger("go-fosp/fosp/postgresql-driver")
	var err error
	d.db, err = sql.Open("postgres", connectionString)
	if err != nil {
		d.lg.Fatal("Error occured when establishing db connection :: ", err)
	}
	d.db.SetMaxOpenConns(50)
	return d
}

func (d *postgresqlDriver) Authenticate(name, password string) error {
	var passwordHash string
	err := d.db.QueryRow("SELECT password FROM users WHERE name = $1", name).Scan(&passwordHash)
	if err != nil {
		d.lg.Error("Error when selecting record for authentication: ", err)
		return err
	} else if err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		return errors.New("error when comparing passwords")
	} else {
		return nil
	}
}

func (d *postgresqlDriver) Register(name, password string) error {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		d.lg.Error("Error when generating password hash: ", err)
		return InternalServerError
	}
	_, err = d.db.Exec("INSERT INTO users (name, password) VALUES ($1, $2)", name, string(passwordHash))
	if err != nil {
		d.lg.Error("Error when adding new user: ", err)
		return InternalServerError
	}
	return nil
}

func (d *postgresqlDriver) GetNodeWithParents(url *URL) (Object, error) {
	urls := make([]string, 0, len(url.path))
	for !url.IsRoot() {
		urls = append(urls, `'`+url.String()+`'`)
		url = url.Parent()
	}
	urls = append(urls, `'`+url.String()+`'`)

	rows, err := d.db.Query("SELECT * FROM data WHERE uri IN (" + strings.Join(urls, ",") + ") ORDER BY uri ASC")
	if err != nil {
		d.lg.Error("Error when fetching object and parents from database: ", err)
		return Object{}, InternalServerError
	}
	defer rows.Close()
	var parent *Object
	for rows.Next() {
		var (
			id        uint64
			uri       string
			parentID uint64
			content   string
		)
		if err := rows.Scan(&id, &uri, &parentID, &content); err != nil {
			d.lg.Error("Error when reading values from object row: ", err)
			return Object{}, errors.New("internal database error")
		}
		obj, err := Unmarshal(content)
		if err != nil {
			d.lg.Critical("Error when unmarshaling json :: ", err)
			return Object{}, errors.New("internal database error")
		}
		obj.URL, err = parseURL(uri)
		obj.Parent = parent
		parent = obj
	}
	return *parent, nil
}

func (d *postgresqlDriver) CreateNode(url *URL, o *Object) error {
	var parentID uint64
	if !url.IsRoot() {
		err := d.db.QueryRow("SELECT id FROM data WHERE uri = $1", url.Parent().String()).Scan(&parentID)
		if err != nil {
			d.lg.Error("Error when fetching parent for new object: ", err)
			return err
		}
	}
	content, err := json.Marshal(o)
	if err != nil {
		return err
	}
	_, err = d.db.Exec("INSERT INTO data (uri, parent_id, content) VALUES ($1, $2, $3)", url.String(), parentID, content)
	if err != nil {
		d.lg.Error("Error when adding new object: ", err)
		return err
	}
	return nil
}

func (d *postgresqlDriver) UpdateNode(url *URL, o *Object) error {
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

func (d *postgresqlDriver) ListNodes(url *URL) ([]string, error) {
	var parentID uint64
	err := d.db.QueryRow("SELECT id FROM data WHERE uri = $1", url.String()).Scan(&parentID)
	if err != nil {
		return []string{}, err
	}
	var rows *sql.Rows
	rows, err = d.db.Query("SELECT uri FROM data WHERE parent_id = $1", parentID)
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
			d.lg.Critical("Error when reading row :: ", err)
			return nil, errors.New("internal database error")
		}
		uris = append(uris, strings.TrimPrefix(uri, parent))
	}
	return uris, nil
}

func (d *postgresqlDriver) DeleteNodes(url *URL) error {
	_, err := d.db.Exec("DELETE FROM data WHERE uri ~ $1", "^"+url.String())
	return err
}

func (d *postgresqlDriver) ReadAttachment(url *URL) ([]byte, error) {
	hash := sha512.Sum512([]byte(url.Path()))
	filename := base32.StdEncoding.EncodeToString(hash[:sha512.Size])
	path := d.basepath + "/" + filename
	return ioutil.ReadFile(path)
}

func (d *postgresqlDriver) WriteAttachment(url *URL, data []byte) error {
	hash := sha512.Sum512([]byte(url.Path()))
	filename := base32.StdEncoding.EncodeToString(hash[:sha512.Size])
	path := d.basepath + "/" + filename
	return ioutil.WriteFile(path, data, 0660)
}
