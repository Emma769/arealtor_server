package funclib

import (
	"net/mail"
	"regexp"
	"time"
)

func ValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func ValidPhone(phone string) bool {
	return regexp.MustCompile(`^(0)([7-9][0-9]{9})$`).Match([]byte(phone))
}

func DaysBetween(a, b time.Time) int {
	if a.After(b) {
		a, b = b, a
	}

	a = a.Truncate(24 * time.Hour)
	b = b.Truncate(24 * time.Hour)
	duration := b.Sub(a)
	return int(duration.Hours() / 24)
}
