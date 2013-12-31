package main

type FospError struct {
  message string
  code uint
}

func (e FospError) Error() string {
  return e.message
}

var ObjectNotFoundError = FospError{"Object was not found", 404}
var NotAuthorizedError = FospError{"Not authorized", 403}
var InternalServerError = FospError{"Internal server error", 500}
var InvalidRequestError = FospError{"Invalid request", 400}
var UserAlreadyExistsError = FospError{"User already exist", 4001}
var ParentNotPresentError = FospError{"Parent not present", 4002}
