package validator

import (
	"fmt"
	"github.com/go-playground/validator/v10"
)

var (
	v *validator.Validate
)

func init() {
	v = validator.New()
}

func Validate(i interface{}) error {
	if i == nil {
		return fmt.Errorf("data to validate is nil")
	}

	return v.Struct(i)
}
