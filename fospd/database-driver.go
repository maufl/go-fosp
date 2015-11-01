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
	"io"
	"net/url"
)

// DatabaseDriver defines the interface of database drivers.
// A struct that implements this interface can be used by Database to fetch and store all data.
type DatabaseDriver interface {
	Authenticate(string, string) bool
	Register(string, string, *fosp.Object) bool
	GetObjectWithParents(*url.URL) (fosp.Object, error)
	CreateObject(*url.URL, *fosp.Object) error
	UpdateObject(*url.URL, *fosp.Object) error
	ListObjects(*url.URL) ([]string, error)
	DeleteObjects(*url.URL) error
	ReadAttachment(*url.URL) ([]byte, error)
	WriteAttachment(*url.URL, io.Reader) (int64, error)
}
