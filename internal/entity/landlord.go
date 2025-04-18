package entity

import (
	"strings"
	"time"

	"github.com/google/uuid"

	funclib "github.com/emma769/a-realtor/internal/lib/func"
	"github.com/emma769/a-realtor/internal/validator"
)

type PropertyType int

const (
	_ PropertyType = iota
	BlocksOfFlat
	Bungalow
	Duplex
	Flat
	MiniFlat
	OneBedroomFlat
	Roomself
	TwoBedroomFlat
)

func (p PropertyType) String() string {
	return [...]string{
		"Blocks of flat",
		"Bungalow",
		"Duplex",
		"Flat",
		"Mini flat",
		"One bedroom flat",
		"Roomself",
		"Two bedroom flat",
	}[p-1]
}

type PropertyInfo struct {
	PropertyInfoID int64          `json:"propertyInfoID"`
	Address        string         `json:"address"`
	PropertyType   PropertyType   `json:"propertyType"`
	LeasePrice     float64        `json:"leasePrice"`
	LeasePeriod    int            `json:"leasePeriod"`
	StartDate      time.Time      `json:"startDate"`
	EndDate        time.Time      `json:"endDate"`
	AdditionalInfo map[string]any `json:"additionalInfo"`
	LandlordID     uuid.UUID      `json:"-"`
}

type PropertyInfoIn struct {
	Address        string         `json:"address"`
	PropertyType   PropertyType   `json:"propertyType"`
	LeasePrice     float64        `json:"leasePrice"`
	LeasePeriod    int            `json:"leasePeriod"`
	StartDate      DateTime       `json:"startDate"`
	EndDate        DateTime       `json:"endDate"`
	AdditionalInfo map[string]any `json:"additionalInfo"`
}

func ValidatePropertyInfoIn(v *validator.Validator, in PropertyInfoIn) {
	validator.Check(
		v,
		in,
		func(in PropertyInfoIn) (bool, validator.ValidationMsg) {
			return in.Address != "", validator.ValidationMsg{
				Prop: "address",
				Info: "cannot be blank",
			}
		},
		func(in PropertyInfoIn) (bool, validator.ValidationMsg) {
			return in.LeasePrice > 0, validator.ValidationMsg{
				Prop: "leasePrice",
				Info: "must be greater than 0",
			}
		},
		func(in PropertyInfoIn) (bool, validator.ValidationMsg) {
			return in.LeasePeriod > 0, validator.ValidationMsg{
				Prop: "leasePeriod",
				Info: "must be greater than 0",
			}
		},
		func(in PropertyInfoIn) (bool, validator.ValidationMsg) {
			return in.StartDate != DateTime{}, validator.ValidationMsg{
				Prop: "startDate",
				Info: "provide a valid start date",
			}
		},
		func(in PropertyInfoIn) (bool, validator.ValidationMsg) {
			return in.EndDate != DateTime{}, validator.ValidationMsg{
				Prop: "endDate",
				Info: "provide a valid end date",
			}
		},
	)
}

type Landlord struct {
	LandlordID   uuid.UUID       `json:"landlordID"`
	FirstName    string          `json:"firstName"`
	LastName     string          `json:"lastName,omitempty"`
	Email        string          `json:"email,omitempty"`
	Phone        string          `json:"phone"`
	RegisteredBy uuid.UUID       `json:"-"`
	CreatedAt    time.Time       `json:"createdAt"`
	UpdatedAt    *time.Time      `json:"updatedAt,omitempty"`
	PropertyInfo []*PropertyInfo `json:"propertyInfo"`
}

type LandlordOut struct {
	LandlordID     uuid.UUID      `json:"landlordID"`
	FirstName      string         `json:"firstName"`
	LastName       string         `json:"lastName,omitempty"`
	Email          string         `json:"email,omitempty"`
	Phone          string         `json:"phone"`
	RegisteredBy   uuid.UUID      `json:"-"`
	Address        string         `json:"address"`
	PropertyType   PropertyType   `json:"propertyType"`
	LeasePrice     float64        `json:"leasePrice"`
	LeasePeriod    int            `json:"leasePeriod"`
	StartDate      time.Time      `json:"startDate"`
	EndDate        time.Time      `json:"endDate"`
	AdditionalInfo map[string]any `json:"additionalInfo"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      *time.Time     `json:"updatedAt,omitempty"`
}

func UpdateLandlord(landlord *Landlord, in LandlordIn) {}

type LandlordIn struct {
	FirstName      string         `json:"firstName"`
	LastName       string         `json:"lastName"`
	Email          string         `json:"email"`
	Phone          string         `json:"phone"`
	Address        string         `json:"address"`
	PropertyType   PropertyType   `json:"propertyType"`
	AdditionalInfo map[string]any `json:"additionalInfo"`
	LeasePrice     float64        `json:"leasePrice"`
	LeasePeriod    int            `json:"leasePeriod"`
	StartDate      DateTime       `json:"startDate"`
	EndDate        DateTime       `json:"endDate"`
}

func ValidateLandlordIn(v *validator.Validator, in LandlordIn) {
	validator.Check(
		v,
		in,
		func(in LandlordIn) (bool, validator.ValidationMsg) {
			return strings.TrimSpace(in.FirstName) != "", validator.ValidationMsg{
				Prop: "firstName",
				Info: "cannot be blank",
			}
		},
		func(in LandlordIn) (bool, validator.ValidationMsg) {
			return funclib.ValidPhone(in.Phone), validator.ValidationMsg{
				Prop: "phone",
				Info: "provide a valid phone number",
			}
		},
		func(in LandlordIn) (bool, validator.ValidationMsg) {
			return strings.TrimSpace(in.Address) != "", validator.ValidationMsg{
				Prop: "propertyAdress",
				Info: "cannot be blank",
			}
		},
		func(in LandlordIn) (bool, validator.ValidationMsg) {
			return in.LeasePrice > 0, validator.ValidationMsg{
				Prop: "leasePrice",
				Info: "cannot be zero",
			}
		},
		func(in LandlordIn) (bool, validator.ValidationMsg) {
			return in.LeasePeriod > 0, validator.ValidationMsg{
				Prop: "leasePeriod",
				Info: "cannot be zero",
			}
		},
		func(in LandlordIn) (bool, validator.ValidationMsg) {
			return in.StartDate != DateTime{}, validator.ValidationMsg{
				Prop: "startDate",
				Info: "provide a start date",
			}
		},
		func(in LandlordIn) (bool, validator.ValidationMsg) {
			return in.EndDate != DateTime{}, validator.ValidationMsg{
				Prop: "endDate",
				Info: "provide an end date",
			}
		},
	)
}
