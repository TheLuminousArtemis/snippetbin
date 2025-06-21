package main

import (
	"strings"

	"github.com/go-playground/validator/v10"
)

func ValidateStruct(form any) (map[string]string, error) {
	err := validate.Struct(form)
	if err == nil {
		return nil, nil
	}

	fieldErrors := make(map[string]string)
	if ve, ok := err.(validator.ValidationErrors); ok {
		for _, fe := range ve {
			field := strings.ToLower(fe.Field())
			switch fe.Tag() {
			case "required":
				fieldErrors[field] = "This field cannot be blank"
			case "max":
				fieldErrors[field] = "This field cannot be more than 100 characters long"
			case "expires":
				fieldErrors[field] = "This field must equal to 1, 7 or 365"
			case "min":
				fieldErrors[field] = "This field cannot be less than 8"
			case "email":
				fieldErrors[field] = "This field must be a valid email address"
			case "unique":
				fieldErrors[field] = "This field must be unique"
			default:
				fieldErrors[field] = "This field is invalid"
			}
		}
	}
	return fieldErrors, err
}
