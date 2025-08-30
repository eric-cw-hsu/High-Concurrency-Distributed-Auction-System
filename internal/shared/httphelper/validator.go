package httphelper

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

func ParseValidationErrors(err error) map[string]string {
	errs := make(map[string]string)
	if ve, ok := err.(validator.ValidationErrors); ok {
		for _, fe := range ve {
			field := fe.Field()
			tag := fe.Tag()
			errs[field] = fmt.Sprintf("Field '%s' failed validation for tag '%s'", field, tag)
		}
	} else {
		errs["error"] = err.Error()
	}
	return errs
}
