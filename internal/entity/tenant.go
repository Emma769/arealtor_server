package entity

import (
	"strconv"
	"time"

	"github.com/google/uuid"

	funclib "github.com/emma769/a-realtor/internal/lib/func"
	"github.com/emma769/a-realtor/internal/validator"
)

type DateTime struct {
	time.Time
}

func (dt *DateTime) UnmarshalJSON(data []byte) error {
	d, err := strconv.Unquote(string(data))
	if err != nil {
		return err
	}

	t, err := time.Parse("2006-01-02", d)
	if err != nil {
		return err
	}

	dt.Time = t

	return nil
}

type Gender string

type Tenant struct {
	TenantID       uuid.UUID      `json:"tenantID"`
	FirstName      string         `json:"firstName"`
	LastName       string         `json:"lastName,omitempty"`
	Gender         Gender         `json:"gender"`
	DOB            time.Time      `json:"dob"`
	Image          string         `json:"image,omitempty"`
	Email          string         `json:"email,omitempty"`
	Phone          string         `json:"phone"`
	StateOfOrigin  string         `json:"stateOfOrigin"`
	Nationality    string         `json:"nationality"`
	Occupation     string         `json:"occupation"`
	AdditionalInfo map[string]any `json:"additionalInfo"`
	RegisteredBy   uuid.UUID      `json:"-"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      *time.Time     `json:"updatedAt,omitempty"`
	RentInfo       []*RentInfo    `json:"rentInfo,omitempty"`
}

type RentInfo struct {
	RentInfoID   int64     `json:"rentInfoID"`
	StartDate    time.Time `json:"startDate"`
	MaturityDate time.Time `json:"maturityDate"`
	RenewalDate  time.Time `json:"renewalDate"`
	LandlordID   uuid.UUID `json:"landlordID"`
	TenantID     uuid.UUID `json:"tenantID"`
	Address      string    `json:"address"`
	RentFee      float64   `json:"rentFee"`
}

type RentInfoIn struct {
	StartDate    DateTime  `json:"startDate"`
	MaturityDate DateTime  `json:"maturityDate"`
	RenewalDate  DateTime  `json:"renewalDate"`
	LandlordID   uuid.UUID `json:"landlordID"`
	Address      string    `json:"address"`
	RentFee      float64   `json:"rentFee"`
}

func ValidateRentInfoIn(v *validator.Validator, in RentInfoIn) {
	validator.Check(
		v,
		in,
		func(in RentInfoIn) (bool, validator.ValidationMsg) {
			return in.Address != "", validator.ValidationMsg{
				Prop: "address",
				Info: "cannot be blank",
			}
		},
		func(in RentInfoIn) (bool, validator.ValidationMsg) {
			return in.RentFee > 0, validator.ValidationMsg{
				Prop: "rentFee",
				Info: "must be greater than zero",
			}
		},
		func(in RentInfoIn) (bool, validator.ValidationMsg) {
			return in.LandlordID != uuid.UUID{}, validator.ValidationMsg{
				Prop: "landlordID",
				Info: "provide valid landlord information",
			}
		},
		func(in RentInfoIn) (bool, validator.ValidationMsg) {
			return in.RenewalDate != DateTime{}, validator.ValidationMsg{
				Prop: "renewalDate",
				Info: "provide a renewal date",
			}
		},
		func(in RentInfoIn) (bool, validator.ValidationMsg) {
			return in.MaturityDate != DateTime{}, validator.ValidationMsg{
				Prop: "maturityDate",
				Info: "provide a maturity date",
			}
		},
		func(in RentInfoIn) (bool, validator.ValidationMsg) {
			return in.StartDate != DateTime{}, validator.ValidationMsg{
				Prop: "startDate",
				Info: "provide a maturity date",
			}
		},
	)
}

type TenantOut struct {
	TenantID       uuid.UUID      `json:"tenantID"`
	FirstName      string         `json:"firstName"`
	LastName       string         `json:"lastName,omitempty"`
	Gender         Gender         `json:"gender"`
	DOB            time.Time      `json:"dob"`
	Image          string         `json:"image,omitempty"`
	Email          string         `json:"email,omitempty"`
	Phone          string         `json:"phone"`
	StateOfOrigin  string         `json:"stateOfOrigin"`
	Nationality    string         `json:"nationality"`
	RentFee        float64        `json:"rentFee"`
	Occupation     string         `json:"occupation"`
	AdditionalInfo map[string]any `json:"additionalInfo"`
	StartDate      time.Time      `json:"startDate"`
	MaturityDate   time.Time      `json:"maturityDate"`
	RenewalDate    time.Time      `json:"renewalDate"`
	Address        string         `json:"address"`
	LandlordID     uuid.UUID      `json:"landlordID"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      *time.Time     `json:"updatedAt,omitempty"`
}

type TenantIn struct {
	FirstName      string         `json:"firstName"`
	LastName       string         `json:"lastName"`
	Gender         Gender         `json:"gender"`
	DOB            DateTime       `json:"dob"`
	Image          string         `json:"image"`
	Email          string         `json:"email"`
	Phone          string         `json:"phone"`
	StateOfOrigin  string         `json:"stateOfOrigin"`
	Nationality    string         `json:"nationality"`
	Occupation     string         `json:"occupation"`
	AdditionalInfo map[string]any `json:"additionalInfo"`
	StartDate      DateTime       `json:"startDate"`
	MaturityDate   DateTime       `json:"maturityDate"`
	RenewalDate    DateTime       `json:"renewalDate"`
	LandlordID     uuid.UUID      `json:"landlordID"`
	Address        string         `json:"address"`
	RentFee        float64        `json:"rentFee"`
}

func ValidateTenantIn(v *validator.Validator, in TenantIn) {
	validator.Check(
		v,
		in,
		func(in TenantIn) (bool, validator.ValidationMsg) {
			return in.FirstName != "", validator.ValidationMsg{
				Prop: "firstName",
				Info: "cannot be blank",
			}
		},
		func(in TenantIn) (bool, validator.ValidationMsg) {
			return in.LastName != "", validator.ValidationMsg{
				Prop: "lastName",
				Info: "cannot be blank",
			}
		},
		func(in TenantIn) (bool, validator.ValidationMsg) {
			return in.DOB != DateTime{}, validator.ValidationMsg{
				Prop: "dob",
				Info: "provide valid DoB",
			}
		},
		func(in TenantIn) (bool, validator.ValidationMsg) {
			return funclib.ValidPhone(in.Phone), validator.ValidationMsg{
				Prop: "phone",
				Info: "provide a valid phone number",
			}
		},
		func(in TenantIn) (bool, validator.ValidationMsg) {
			return in.StateOfOrigin != "", validator.ValidationMsg{
				Prop: "stateOfOrigin",
				Info: "cannot be blank",
			}
		},
		func(in TenantIn) (bool, validator.ValidationMsg) {
			return in.Nationality != "", validator.ValidationMsg{
				Prop: "nationality",
				Info: "cannot be blank",
			}
		},
		func(in TenantIn) (bool, validator.ValidationMsg) {
			return in.Occupation != "", validator.ValidationMsg{
				Prop: "occupation",
				Info: "cannot be blank",
			}
		},
		func(in TenantIn) (bool, validator.ValidationMsg) {
			return in.StartDate != DateTime{}, validator.ValidationMsg{
				Prop: "startDate",
				Info: "provide a start date",
			}
		},
		func(in TenantIn) (bool, validator.ValidationMsg) {
			return in.Address != "", validator.ValidationMsg{
				Prop: "address",
				Info: "cannot be blank",
			}
		},
		func(in TenantIn) (bool, validator.ValidationMsg) {
			return in.MaturityDate != DateTime{}, validator.ValidationMsg{
				Prop: "maturityDate",
				Info: "provide a maturity date",
			}
		},
		func(in TenantIn) (bool, validator.ValidationMsg) {
			return in.RenewalDate != DateTime{}, validator.ValidationMsg{
				Prop: "renewalDate",
				Info: "provide a renewal date",
			}
		},
		func(in TenantIn) (bool, validator.ValidationMsg) {
			return in.LandlordID != uuid.UUID{}, validator.ValidationMsg{
				Prop: "landlordID",
				Info: "provide valid landlord information",
			}
		},
		func(in TenantIn) (bool, validator.ValidationMsg) {
			return in.Address != "", validator.ValidationMsg{
				Prop: "address",
				Info: "cannot be blank",
			}
		},
		func(in TenantIn) (bool, validator.ValidationMsg) {
			return in.RentFee > 0, validator.ValidationMsg{
				Prop: "rentFee",
				Info: "must be greater than zero",
			}
		},
	)
}
