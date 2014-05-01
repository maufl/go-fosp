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

// FospError represents an error that is known to FOSP.
type FospError struct {
	Message string
	Code    uint
}

// Error returns the error message.
func (e FospError) Error() string {
	return e.Message
}

// ErrInvalidRequest is returned when the request is not valid.
var ErrInvalidRequest = FospError{"Invalid request", 4000}

// ErrAuthenticationFailed is returned when the user provides invalid credentials.
var ErrAuthenticationFailed = FospError{"authentication failed", 4010}

// ErrNotAuhtorized is returned when the users rights do not suffice.
var ErrNotAuthorized = FospError{"Not authorized", 4030}

// ErrObjectNotFound is returned when an requested object was not found.
var ErrObjectNotFound = FospError{"Object was not found", 4040}

// ErrParentNotPresent is returned when a request failed because a parent object does not exist.
var ErrParentNotPresent = FospError{"Parent not present", 4041}

// ErrUserAlreadyExists is returned when a registration failed because a user with the same name already exists.
var ErrUserAlreadyExists = FospError{"User already exist", 4095}

// ErrInternalServerError is returned when the server encountered a unexpected error.
var ErrInternalServerError = FospError{"Internal server error", 5000}
