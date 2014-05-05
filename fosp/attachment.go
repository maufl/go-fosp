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

// Attachment represents a the content of the attachemnt field in an object.
type Attachment struct {
	Name string `json:"name,omitempty"`
	Size uint   `json:"size,omitempty"`
	Type string `json:"type,omitempty"`
}

// NewAttachment creates a new Attachment struct and returns it.
func NewAttachment() *Attachment {
	return &Attachment{}
}

// Empty returns true if no field is set in this attachment struct.
func (a *Attachment) Empty() bool {
	return a.Name == "" && a.Size == 0 && a.Type == ""
}
