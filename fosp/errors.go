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
	message string
	code    uint
}

// Error returns the error message.
func (e FospError) Error() string {
	return e.message
}

// Code returns the error code.
func (e FospError) Code() uint {
	return e.code
}

// ObjectNotFoundError is returned when an requested object was not found.
var ObjectNotFoundError = FospError{"Object was not found", 404}
// NotAuhtorizedError is returned when the users rights do not suffice.
var NotAuthorizedError = FospError{"Not authorized", 403}
// InternalServerError is returned when the server encountered a unexpected error.
var InternalServerError = FospError{"Internal server error", 500}
// InvalidRequestError is returned when the request is not valid.
var InvalidRequestError = FospError{"Invalid request", 400}
// UserAlreadyExistsError is returned when a registration failed because a user with the same name already exists.
var UserAlreadyExistsError = FospError{"User already exist", 4001}
// ParentNotPresentError is returned when a request failed because a parent object does not exist.
var ParentNotPresentError = FospError{"Parent not present", 4002}
