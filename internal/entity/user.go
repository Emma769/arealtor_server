package entity

import (
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	funclib "github.com/emma769/a-realtor/internal/lib/func"
	"github.com/emma769/a-realtor/internal/validator"
)

var UserCtxKey = struct{}{}

var AnonymousUser = new(User)

type User struct {
	UserID    uuid.UUID `json:"userID"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  []byte    `json:"password"`
	CreatedAt time.Time `json:"createdAt"`
}

func (u *User) IsAnonymous() bool {
	return u == AnonymousUser
}

type UserIn struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func ValidateUserIn(v *validator.Validator, in UserIn) {
	validator.Check(
		v,
		in,
		func(in UserIn) (bool, validator.ValidationMsg) {
			return strings.TrimSpace(in.Name) != "", validator.ValidationMsg{
				Prop: "name",
				Info: "cannot be blank",
			}
		},
		func(in UserIn) (bool, validator.ValidationMsg) {
			return strings.TrimSpace(in.Email) != "", validator.ValidationMsg{
				Prop: "email",
				Info: "cannot be blank",
			}
		},
		func(in UserIn) (bool, validator.ValidationMsg) {
			return utf8.RuneCountInString(
					strings.TrimSpace(in.Password),
				) >= 8, validator.ValidationMsg{
					Prop: "password",
					Info: "cannot be less than 8 characters",
				}
		},
		func(in UserIn) (bool, validator.ValidationMsg) {
			return funclib.ValidEmail(in.Email), validator.ValidationMsg{
				Prop: "email",
				Info: "provide valid email",
			}
		},
	)
}

type LoginIn struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func ValidateLoginIn(v *validator.Validator, in LoginIn) {
	validator.Check(
		v,
		in,
		func(in LoginIn) (bool, validator.ValidationMsg) {
			return strings.TrimSpace(in.Email) != "", validator.ValidationMsg{
				Prop: "email",
				Info: "cannot be blank",
			}
		},
		func(in LoginIn) (bool, validator.ValidationMsg) {
			return funclib.ValidEmail(in.Email), validator.ValidationMsg{
				Prop: "email",
				Info: "provide valid email",
			}
		},
		func(in LoginIn) (bool, validator.ValidationMsg) {
			return strings.TrimSpace(in.Password) != "", validator.ValidationMsg{
				Prop: "password",
				Info: "cannot be blank",
			}
		},
	)
}
