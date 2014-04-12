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

type FospError struct {
	message string
	code    uint
}

func (e FospError) Error() string {
	return e.message
}

func (e FospError) Code() uint {
	return e.code
}

var ObjectNotFoundError = FospError{"Object was not found", 404}
var NotAuthorizedError = FospError{"Not authorized", 403}
var InternalServerError = FospError{"Internal server error", 500}
var InvalidRequestError = FospError{"Invalid request", 400}
var UserAlreadyExistsError = FospError{"User already exist", 4001}
var ParentNotPresentError = FospError{"Parent not present", 4002}
