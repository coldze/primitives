package json_config

import "github.com/coldze/primitives/custom_error"

type Validatable interface {
	Validate() custom_error.CustomError
}
