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

package fosp

const (
	OPTIONS string = "OPTIONS"
	AUTH           = "AUTH"
	GET            = "GET"
	LIST           = "LIST"
	CREATE         = "CREATE"
	PATCH          = "PATCH"
	DELETE         = "DELETE"
	READ           = "READ"
	WRITE          = "WRITE"

	SUCCEEDED = "SUCCEEDED"
	FAILED    = "FAILED"

	CREATED = "CREATED"
	UPDATED = "UPDATED"
	DELETED = "DELETED"
)

// Message is the common interface of all FOSP messag objects.
type Message interface {
	String() string

	// This unexported method allows only structs of this package to implement the interface
	nop()
}
