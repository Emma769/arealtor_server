package landlord

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/emma769/a-realtor/internal/entity"
	handlerlib "github.com/emma769/a-realtor/internal/lib/handler"
	"github.com/emma769/a-realtor/internal/repository"
	"github.com/emma769/a-realtor/internal/repository/psql"
)

var (
	ErrNotFound     = errors.New("not found")
	ErrDuplicateKey = errors.New("duplicate key")
)

type storer interface {
	CreateLandlord(context.Context, psql.LandlordWithPropertyInfoParam) (*entity.Landlord, error)
	FindLandlord(context.Context, uuid.UUID) (*entity.Landlord, error)
	FindLandlords(
		context.Context,
		psql.LandlordFilterParam,
		psql.PaginationParam,
	) ([]*entity.LandlordOut, error)
	UpdateLandlord(context.Context, psql.LandlordParam) (*entity.Landlord, error)
	DeleteLandlord(context.Context, uuid.UUID) error
	TotalLandlordCount(context.Context) (int64, error)
	CreatePropertyInfo(
		context.Context,
		uuid.UUID,
		psql.PropertyInfoParam,
	) (*entity.PropertyInfo, error)
	GetAllLandlords(context.Context) ([]*entity.LandlordOut, error)
}

type Service struct {
	store   storer
	timeout time.Duration
}

type LandlordParam struct {
	firstname      string
	lastname       string
	email          string
	phone          string
	address        string
	propertyType   entity.PropertyType
	additionalInfo map[string]any
	leasePrice     float64
	leasePeriod    int
	startDate      time.Time
	endDate        time.Time
	registeredBy   uuid.UUID
}

func (param LandlordParam) FirstName() string {
	return param.firstname
}

func (param LandlordParam) LastName() string {
	return param.lastname
}

func (param LandlordParam) Email() string {
	return param.email
}

func (param LandlordParam) Phone() string {
	return param.phone
}

func (param LandlordParam) Address() string {
	return param.address
}

func (param LandlordParam) PropertyType() entity.PropertyType {
	return param.propertyType
}

func (param LandlordParam) AdditionalInfo() []byte {
	b, _ := json.Marshal(param.additionalInfo)
	return b
}

func (param LandlordParam) LeasePrice() float64 {
	return param.leasePrice
}

func (param LandlordParam) LeasePeriod() int {
	return param.leasePeriod
}

func (param LandlordParam) StartDate() time.Time {
	return param.startDate
}

func (param LandlordParam) EndDate() time.Time {
	return param.endDate
}

func (param LandlordParam) RegisteredBy() uuid.UUID {
	return param.registeredBy
}

func (s *Service) create(
	ctx context.Context,
	user *entity.User,
	in entity.LandlordIn,
) (*entity.Landlord, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	param := LandlordParam{
		firstname:      in.FirstName,
		lastname:       in.LastName,
		email:          in.Email,
		phone:          in.Phone,
		address:        in.Address,
		propertyType:   in.PropertyType,
		additionalInfo: in.AdditionalInfo,
		leasePrice:     in.LeasePrice,
		leasePeriod:    in.LeasePeriod,
		startDate:      in.StartDate.Time,
		endDate:        in.EndDate.Time,
		registeredBy:   user.UserID,
	}

	landlord, err := s.store.CreateLandlord(ctx, param)

	if err != nil && errors.Is(err, repository.ErrDuplicateKey) {
		return nil, ErrDuplicateKey
	}

	if err != nil {
		return nil, err
	}

	return landlord, nil
}

func (s *Service) findOne(ctx context.Context, id uuid.UUID) (*entity.Landlord, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	landlord, err := s.store.FindLandlord(ctx, id)

	if err != nil && errors.Is(err, repository.ErrNotFound) {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	return landlord, nil
}

func (s *Service) update(_ context.Context, _ *entity.Landlord) (*entity.Landlord, error) {
	return nil, nil
}

type FilterParam struct {
	address,
	firstName,
	phone string
}

func (f FilterParam) Address() string {
	return f.address
}

func (f FilterParam) FirstName() string {
	return f.firstName
}

func (f FilterParam) Phone() string {
	return f.phone
}

func (s *Service) findAll(
	ctx context.Context,
	filterParam FilterParam,
	paginator *handlerlib.Paginator,
) ([]*entity.LandlordOut, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	landlords, err := s.store.FindLandlords(ctx, filterParam, paginator)
	if err != nil {
		return nil, err
	}

	return landlords, nil
}

func (s *Service) deleteOne(ctx context.Context, id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	return s.store.DeleteLandlord(ctx, id)
}

func (s *Service) total(ctx context.Context) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	return s.store.TotalLandlordCount(ctx)
}

type PropertyInfoParam struct {
	address        string
	propertyType   entity.PropertyType
	leasePrice     float64
	leasePeriod    int
	startDate      time.Time
	endDate        time.Time
	additionalInfo map[string]any
}

func (param PropertyInfoParam) Address() string {
	return param.address
}

func (param PropertyInfoParam) PropertyType() entity.PropertyType {
	return param.propertyType
}

func (param PropertyInfoParam) LeasePrice() float64 {
	return param.leasePrice
}

func (param PropertyInfoParam) LeasePeriod() int {
	return param.leasePeriod
}

func (param PropertyInfoParam) StartDate() time.Time {
	return param.startDate
}

func (param PropertyInfoParam) EndDate() time.Time {
	return param.endDate
}

func (param PropertyInfoParam) AdditionalInfo() []byte {
	b, _ := json.Marshal(param.additionalInfo)
	return b
}

func (s *Service) createPropertyInfo(
	ctx context.Context,
	id uuid.UUID,
	in entity.PropertyInfoIn,
) (*entity.PropertyInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	param := PropertyInfoParam{
		address:        in.Address,
		propertyType:   in.PropertyType,
		leasePrice:     in.LeasePrice,
		leasePeriod:    in.LeasePeriod,
		startDate:      in.StartDate.Time,
		endDate:        in.EndDate.Time,
		additionalInfo: in.AdditionalInfo,
	}

	info, err := s.store.CreatePropertyInfo(ctx, id, param)
	if err != nil {
		return nil, err
	}

	return info, nil
}

func (s *Service) getAll(ctx context.Context) ([]*entity.LandlordOut, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	return s.store.GetAllLandlords(ctx)
}
