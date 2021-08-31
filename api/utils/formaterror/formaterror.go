package formaterror

import (
	"errors"
	"strings"
)

func LoginError(err string) error {
	return errors.New("incorrect credentials")
}

func FormatError(err string) error {

	if strings.Contains(err, "first_name") {
		return errors.New("FirstName Already Taken")
	}

	if strings.Contains(err, "email") {
		return errors.New("Email Already Taken")
	}

	if strings.Contains(err, "title") {
		return errors.New("Title Already Taken")
	}
	if strings.Contains(err, "hashedPassword") {
		return errors.New("Incorrect Password")
	}
	return errors.New("Incorrect Login")
}

type GeneralError struct {
	Errors []string `json:"errors"`
}
