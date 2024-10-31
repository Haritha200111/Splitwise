package error

import "fmt"

func InvalidParameter(c string) error {
	return &Error{Code: "BAD_INPUT", Message: fmt.Sprintf("%v", c)}
}

var INTERNAL_ERROR = &Error{Code: "INTERNAL_ERROR", Message: "Internal Error"}
var LIMIT_EXCEEDS = &Error{Code: "LIMIT_EXCEEDS", Message: "Maximum Limit Reached"}
var ErrInvalidCredential = &Error{Code: "BAD_CREDENTIAL", Message: "Invalid Credential"}
var ErrInvalidToken = &Error{Code: "BAD_CREDENTIAL", Message: "Invalid Credential"}
var ErrInvalidRequest = &Error{Code: "BAD_INPUT", Message: "Invalid Request"}
var ErrInvalidTOTP = &Error{Code: "INVALID_TOTP", Message: "Invalid TOTP"}
var NOT_FOUND_USER = &Error{Code: "NOT_FOUND_USER", Message: "User not found"}
