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
	// This import is needed to make the postgres driver available to database/sql.
	_ "github.com/lib/pq"
	"github.com/op/go-logging"
	"io/ioutil"
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
		psqlLog.Fatal("Error occured when establishing db connection :: ", err)
	}
	d.db.SetMaxOpenConns(50)
	return d
}

// Authenticate checks whether the name password pair is valid.
// On success nil is returned or an error otherwise.
func (d *PostgresqlDriver) Authenticate(name, password string) error {
	var passwordHash string
	err := d.db.QueryRow("SELECT password FROM users WHERE name = $1", name).Scan(&passwordHash)
	if err != nil {
		psqlLog.Error("Error when selecting record for authentication: ", err)
		return err
	} else if err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		return ErrAuthenticationFailed
	} else {
		return nil
	}
}

// Register creates a new user with the given name and password in the database.
// The password is hashed using the bcrypt library before it is stored in the database.
// If the user already exists or the hasing fails an error is returned.
// On success nil is returned.
func (d *PostgresqlDriver) Register(name, password string) error {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		psqlLog.Error("Error when generating password hash: ", err)
		return ErrInternalServerError
	}
	var exists bool
	d.db.QueryRow("SELECT TRUE FROM users WHERE name = $1", name).Scan(&exists)
	if exists {
		return ErrUserAlreadyExists
	}
	_, err = d.db.Exec("INSERT INTO users (name, password) VALUES ($1, $2)", name, string(passwordHash))
	if err != nil {
		psqlLog.Error("Error when adding new user: ", err)
		return ErrInternalServerError
	}
	return nil
}

// GetNodeWithParents returns an object and all it's parents from the database.
// The parents are stored recursively in the object.
func (d *PostgresqlDriver) GetNodeWithParents(url *URL) (Object, error) {
	urls := make([]string, 0, len(url.path))
	for !url.IsRoot() {
		urls = append(urls, `'`+url.String()+`'`)
		url = url.Parent()
	}
	urls = append(urls, `'`+url.String()+`'`)

	rows, err := d.db.Query("SELECT * FROM data WHERE uri IN (" + strings.Join(urls, ",") + ") ORDER BY uri ASC")
	if err != nil {
		psqlLog.Error("Error when fetching object and parents from database: ", err)
		return Object{}, ErrInternalServerError
	}
	defer rows.Close()
	var parent *Object
	var numObjects int
	for rows.Next() {
		var (
			id       uint64
			uri      string
			parentID uint64
			content  string
		)
		if err := rows.Scan(&id, &uri, &parentID, &content); err != nil {
			psqlLog.Error("Error when reading values from object row: ", err)
			return Object{}, ErrInternalServerError
		}
		obj, err := Unmarshal(content)
		if err != nil {
			psqlLog.Critical("Error when unmarshaling json :: ", err)
			return Object{}, ErrInternalServerError
		}
		obj.URL, err = ParseURL(uri)
		obj.Parent = parent
		parent = obj
		numObjects++
	}
	if numObjects != len(urls) {
		return Object{}, ErrObjectNotFound
	}
	return *parent, nil
}

// CreateNode saves a new object to the database under the given URL.
func (d *PostgresqlDriver) CreateNode(url *URL, o *Object) error {
	var parentID uint64
	if !url.IsRoot() {
		err := d.db.QueryRow("SELECT id FROM data WHERE uri = $1", url.Parent().String()).Scan(&parentID)
		if err != nil {
			psqlLog.Error("Error when fetching parent for new object: ", err)
			return err
		}
	}
	content, err := json.Marshal(o)
	if err != nil {
		return err
	}
	_, err = d.db.Exec("INSERT INTO data (uri, parent_id, content) VALUES ($1, $2, $3)", url.String(), parentID, content)
	if err != nil {
		psqlLog.Error("Error when adding new object: ", err)
		return err
	}
	return nil
}

// UpdateNode replaces the object at the given URL with a new object.
func (d *PostgresqlDriver) UpdateNode(url *URL, o *Object) error {
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

// ListNodes returns an array of child object names of the object at the given URL.
func (d *PostgresqlDriver) ListNodes(url *URL) ([]string, error) {
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
			psqlLog.Critical("Error when reading row :: ", err)
			return nil, ErrInternalServerError
		}
		uris = append(uris, strings.TrimPrefix(uri, parent))
	}
	return uris, nil
}

// DeleteNodes deletes the object at the given URL and all its children.
func (d *PostgresqlDriver) DeleteNodes(url *URL) error {
	_, err := d.db.Exec("DELETE FROM data WHERE uri ~ $1", "^"+url.String())
	return err
}

// ReadAttachment returns the content of the attached file of the object at the given URL.
func (d *PostgresqlDriver) ReadAttachment(url *URL) ([]byte, error) {
	hash := sha512.Sum512([]byte(url.Path()))
	filename := base32.StdEncoding.EncodeToString(hash[:sha512.Size])
	path := d.basepath + "/" + filename
	return ioutil.ReadFile(path)
}

// WriteAttachment stores the data as the attachment of the object at the given URL.
func (d *PostgresqlDriver) WriteAttachment(url *URL, data []byte) error {
	hash := sha512.Sum512([]byte(url.Path()))
	filename := base32.StdEncoding.EncodeToString(hash[:sha512.Size])
	path := d.basepath + "/" + filename
	return ioutil.WriteFile(path, data, 0660)
}
