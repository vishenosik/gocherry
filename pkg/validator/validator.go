package validator

import (
	"github.com/go-playground/validator/v10"
)

var (
	valid = validator.New()
)

func Struct(Struct any) error {
	return valid.Struct(Struct)
}

func UUID4(uuid string) error {
	return valid.Var(uuid, "uuid4")
}

func Email(email string) error {
	return valid.Var(email, "email")
}
