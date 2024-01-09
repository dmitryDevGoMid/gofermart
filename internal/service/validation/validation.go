package validation

import "net/mail"

func ValidationEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}
