package fosp

import (
	"errors"
	"strings"
)

var MalformedUserIdentifierError = errors.New("Malformed user identifier")

type User struct {
	name   string
	domain string
}

func NewUser(name, domain string) *User {
	return &User{name, domain}
}

func (u *User) Name() string {
	return u.name
}

func (u *User) Domain() string {
	return u.domain
}

func (u *User) UnmarshalJSON(data []byte) error {
	return u.parse(data)
}

func (u *User) parse(data []byte) error {
	userString := string(data)
	parts := strings.Split(userString, "@")
	if len(parts) != 2 {
		return MalformedUserIdentifierError
	}
	u.name = parts[0]
	u.domain = strings.TrimSuffix(parts[1], ".")
	return nil
}

func ParseUser(data []byte) (*User, error) {
	u := &User{}
	return u, u.parse(data)
}

func ParseUserString(data string) (*User, error) {
	return ParseUser([]byte(data))
}
