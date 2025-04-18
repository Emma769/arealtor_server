package tenant

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
	ErrDuplicateKey = errors.New("duplicate key")
	ErrNotFound     = errors.New("not found")
)

type storer interface {
	CreateTenant(context.Context, psql.TenantParam) (*entity.Tenant, error)
	FindTenants(
		context.Context,
		psql.TenantFilterParam,
		psql.PaginationParam,
	) ([]*entity.TenantOut, error)
	FindTenant(context.Context, uuid.UUID) (*entity.Tenant, error)
	TenantTotalCount(context.Context) (int64, error)
	DeleteTenant(context.Context, uuid.UUID) error
	CreateRentInfo(context.Context, uuid.UUID, psql.RentInfoParam) (*entity.RentInfo, error)
	FindAllTenants(context.Context) ([]*entity.TenantOut, error)
	FindLandlord(context.Context, uuid.UUID) (*entity.Landlord, error)
}

type Service struct {
	store   storer
	timeout time.Duration
}

type TenantParam struct {
	firstName      string
	lastName       string
	gender         entity.Gender
	dob            time.Time
	image          string
	email          string
	address        string
	phone          string
	stateOfOrigin  string
	nationality    string
	occupation     string
	additionalInfo map[string]any
	startDate      time.Time
	maturityDate   time.Time
	renewalDate    time.Time
	rentFee        float64
	landlordID     uuid.UUID
	registeredBy   uuid.UUID
}

func (param TenantParam) FirstName() string {
	return param.firstName
}

func (param TenantParam) LastName() string {
	return param.lastName
}

func (param TenantParam) Gender() entity.Gender {
	return param.gender
}

func (param TenantParam) DOB() time.Time {
	return param.dob
}

func (param TenantParam) Image() string {
	return param.image
}

func (param TenantParam) Email() string {
	return param.email
}

func (param TenantParam) Address() string {
	return param.address
}

func (param TenantParam) Phone() string {
	return param.phone
}

func (param TenantParam) StateOfOrigin() string {
	return param.stateOfOrigin
}

func (param TenantParam) Nationality() string {
	return param.nationality
}

func (param TenantParam) Occupation() string {
	return param.occupation
}

func (param TenantParam) AdditionalInfo() []byte {
	info, _ := json.Marshal(param.additionalInfo)
	return info
}

func (param TenantParam) StartDate() time.Time {
	return param.startDate
}

func (param TenantParam) MaturityDate() time.Time {
	return param.maturityDate
}

func (param TenantParam) RenewalDate() time.Time {
	return param.renewalDate
}

func (param TenantParam) RentFee() float64 {
	return param.rentFee
}

func (param TenantParam) LandlordID() uuid.UUID {
	return param.landlordID
}

func (param TenantParam) RegisteredBy() uuid.UUID {
	return param.registeredBy
}

func (s *Service) create(
	ctx context.Context,
	user *entity.User,
	in entity.TenantIn,
) (*entity.Tenant, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	param := TenantParam{
		firstName:      in.FirstName,
		lastName:       in.LastName,
		gender:         in.Gender,
		dob:            in.DOB.Time,
		image:          in.Image,
		email:          in.Email,
		phone:          in.Phone,
		stateOfOrigin:  in.StateOfOrigin,
		nationality:    in.Nationality,
		occupation:     in.Occupation,
		additionalInfo: in.AdditionalInfo,
		address:        in.Address,
		startDate:      in.StartDate.Time,
		maturityDate:   in.MaturityDate.Time,
		renewalDate:    in.RenewalDate.Time,
		rentFee:        in.RentFee,
		registeredBy:   user.UserID,
		landlordID:     in.LandlordID,
	}

	tenant, err := s.store.CreateTenant(ctx, param)

	if err != nil && errors.Is(err, repository.ErrDuplicateKey) {
		return nil, ErrDuplicateKey
	}

	if err != nil {
		return nil, err
	}

	return tenant, nil
}

type FilterParam struct {
	firstName,
	phone,
	address string
}

func (f FilterParam) FirstName() string {
	return f.firstName
}

func (f FilterParam) Phone() string {
	return f.phone
}

func (f FilterParam) Address() string {
	return f.address
}

func (s *Service) findall(
	ctx context.Context,
	filterParam *FilterParam,
	paginator *handlerlib.Paginator,
) ([]*entity.TenantOut, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	tenants, err := s.store.FindTenants(ctx, filterParam, paginator)
	if err != nil {
		return nil, err
	}

	return tenants, nil
}

func (s *Service) findone(ctx context.Context, id uuid.UUID) (*entity.Tenant, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	tenant, err := s.store.FindTenant(ctx, id)

	if err != nil && errors.Is(err, repository.ErrNotFound) {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	return tenant, nil
}

func (s *Service) total(ctx context.Context) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	return s.store.TenantTotalCount(ctx)
}

func (s *Service) delete(ctx context.Context, id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	return s.store.DeleteTenant(ctx, id)
}

type RentInfoParam struct {
	address      string
	startDate    time.Time
	maturityDate time.Time
	renewalDate  time.Time
	landlordID   uuid.UUID
	rentFee      float64
}

func (param RentInfoParam) Address() string {
	return param.address
}

func (param RentInfoParam) StartDate() time.Time {
	return param.startDate
}

func (param RentInfoParam) MaturityDate() time.Time {
	return param.maturityDate
}

func (param RentInfoParam) RenewalDate() time.Time {
	return param.renewalDate
}

func (param RentInfoParam) RentFee() float64 {
	return param.rentFee
}

func (param RentInfoParam) LandlordID() uuid.UUID {
	return param.landlordID
}

func (s *Service) createRentInfo(
	ctx context.Context,
	id uuid.UUID,
	in entity.RentInfoIn,
) (*entity.RentInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	param := RentInfoParam{
		address:      in.Address,
		startDate:    in.StartDate.Time,
		maturityDate: in.MaturityDate.Time,
		renewalDate:  in.RenewalDate.Time,
		landlordID:   in.LandlordID,
		rentFee:      in.RentFee,
	}

	return s.store.CreateRentInfo(ctx, id, param)
}

func (s *Service) getAll(ctx context.Context) ([]*entity.TenantOut, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	return s.store.FindAllTenants(ctx)
}

func (s *Service) getLandlord(ctx context.Context, id uuid.UUID) (*entity.Landlord, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	return s.store.FindLandlord(ctx, id)
}
