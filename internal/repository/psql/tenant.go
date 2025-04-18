package psql

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/emma769/a-realtor/internal/entity"
	"github.com/emma769/a-realtor/internal/repository"
)

type tenantParam interface {
	FirstName() string
	LastName() string
	Gender() entity.Gender
	DOB() time.Time
	Image() string
	Email() string
	Phone() string
	StateOfOrigin() string
	Nationality() string
	Occupation() string
	AdditionalInfo() []byte
	RegisteredBy() uuid.UUID
}

type RentInfoParam interface {
	Address() string
	RentFee() float64
	StartDate() time.Time
	MaturityDate() time.Time
	RenewalDate() time.Time
	LandlordID() uuid.UUID
}

type TenantParam interface {
	tenantParam
	RentInfoParam
}

func (repo *Repository) CreateTenant(
	ctx context.Context,
	param TenantParam,
) (*entity.Tenant, error) {
	tx, err := repo.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("beginTx error: %w", err)
	}

	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			repo.logger.LogAttrs(ctx, slog.LevelError, "rollbackTx err", slog.Attr{
				Key:   "detail",
				Value: slog.StringValue(err.Error()),
			})
		}
	}()

	tenant, err := repo.WithTX(tx).createTenant(ctx, param)
	if err != nil {
		return nil, err
	}

	rentInfo, err := repo.WithTX(tx).CreateRentInfo(ctx, tenant.TenantID, param)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commitTx err: %w", err)
	}

	tenant.RentInfo = append(tenant.RentInfo, rentInfo)

	return tenant, nil
}

func (q *queries) createTenant(ctx context.Context, param TenantParam) (*entity.Tenant, error) {
	const query = `
    INSERT INTO tenants (
      first_name, last_name, gender, dob, image, email, phone, state_of_origin, 
      nationality, occupation, additional_info, registered_by
    )
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) 
    RETURNING 
      tenant_id, first_name, last_name, gender, dob, image, email, phone, 
      state_of_origin, nationality, occupation, additional_info, 
      registered_by, created_at, updated_at;
  `
	row := q.db.QueryRowContext(
		ctx,
		query,
		param.FirstName(),
		param.LastName(),
		param.Gender(),
		param.DOB(),
		param.Image(),
		param.Email(),
		param.Phone(),
		param.StateOfOrigin(),
		param.Nationality(),
		param.Occupation(),
		param.AdditionalInfo(),
		param.RegisteredBy(),
	)

	var additionalInfo []byte
	var tenant entity.Tenant

	err := row.Scan(
		&tenant.TenantID,
		&tenant.FirstName,
		&tenant.LastName,
		&tenant.Gender,
		&tenant.DOB,
		&tenant.Image,
		&tenant.Email,
		&tenant.Phone,
		&tenant.StateOfOrigin,
		&tenant.Nationality,
		&tenant.Occupation,
		&additionalInfo,
		&tenant.RegisteredBy,
		&tenant.CreatedAt,
		&tenant.UpdatedAt,
	)

	if err != nil && strings.Contains(err.Error(), "duplicate") {
		return nil, repository.ErrDuplicateKey
	}

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(additionalInfo, &tenant.AdditionalInfo); err != nil {
		return nil, err
	}

	return &tenant, nil
}

func (q *queries) CreateRentInfo(
	ctx context.Context,
	id uuid.UUID,
	param RentInfoParam,
) (*entity.RentInfo, error) {
	const query = `
    INSERT INTO rent_info 
      (start_date, maturity_date, renewal_date, landlord_id, tenant_id, address, rent_fee) 
    VALUES ($1, $2, $3, $4, $5, $6, $7)
    RETURNING 
      rent_info_id, start_date, maturity_date, renewal_date, 
      landlord_id, tenant_id, address, rent_fee;
  `
	row := q.db.QueryRowContext(
		ctx,
		query,
		param.StartDate(),
		param.MaturityDate(),
		param.RenewalDate(),
		param.LandlordID(),
		id,
		param.Address(),
		param.RentFee(),
	)

	var rentInfo entity.RentInfo

	err := row.Scan(
		&rentInfo.RentInfoID,
		&rentInfo.StartDate,
		&rentInfo.MaturityDate,
		&rentInfo.RenewalDate,
		&rentInfo.LandlordID,
		&rentInfo.TenantID,
		&rentInfo.Address,
		&rentInfo.RentFee,
	)

	return &rentInfo, err
}

type TenantFilterParam interface {
	FirstName() string
	Phone() string
	Address() string
}

func (q *queries) FindTenants(
	ctx context.Context,
	filterParam TenantFilterParam,
	paginator PaginationParam,
) ([]*entity.TenantOut, error) {
	const query = `
    SELECT COUNT(*) OVER(), t.tenant_id, t.first_name, t.last_name, t.gender, t.dob, 
      t.image, t.email, t.phone, t.state_of_origin, t.nationality, t.occupation,
      t.additional_info, r.start_date, r.maturity_date, r.renewal_date,
      r.address, r.rent_fee, r.landlord_id, t.created_at, t.updated_at
    FROM tenants t LEFT JOIN rent_info r ON t.tenant_id = r.tenant_id
    WHERE (lower(t.first_name) = lower($1) OR $1 = '') 
    AND (t.phone = $2 OR $2 = '')
    AND (to_tsvector('simple', r.address) @@ plainto_tsquery('simple', $3) OR $3 = '')
    LIMIT $4 OFFSET $5;
  `

	rows, err := q.db.QueryContext(
		ctx,
		query,
		filterParam.FirstName(),
		filterParam.Phone(),
		filterParam.Address(),
		paginator.Limit(),
		paginator.Offset(),
	)
	if err != nil {
		return nil, err
	}

	var total int
	tenants := []*entity.TenantOut{}

	for rows.Next() {
		var additionalInfo []byte
		var tenant entity.TenantOut

		err := rows.Scan(
			&total,
			&tenant.TenantID,
			&tenant.FirstName,
			&tenant.LastName,
			&tenant.Gender,
			&tenant.DOB,
			&tenant.Image,
			&tenant.Email,
			&tenant.Phone,
			&tenant.StateOfOrigin,
			&tenant.Nationality,
			&tenant.Occupation,
			&additionalInfo,
			&tenant.StartDate,
			&tenant.MaturityDate,
			&tenant.RenewalDate,
			&tenant.Address,
			&tenant.RentFee,
			&tenant.LandlordID,
			&tenant.CreatedAt,
			&tenant.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(additionalInfo, &tenant.AdditionalInfo); err != nil {
			return nil, err
		}

		tenants = append(tenants, &tenant)
	}

	paginator.SetTotal(total)

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if err := rows.Close(); err != nil {
		return nil, err
	}

	return tenants, nil
}

func (q *queries) FindTenant(ctx context.Context, id uuid.UUID) (*entity.Tenant, error) {
	const query = `
    SELECT t.tenant_id, t.first_name, t.last_name, t.gender, t.dob, t.image, t.email, 
      t.phone, t.state_of_origin, t.nationality, t.occupation, t.additional_info, 
      t.registered_by, t.created_at, t.updated_at,
      CASE WHEN count(r.rent_info_id) = 0 THEN
        '[]'::JSON
      ELSE
        json_agg(json_build_object(
          'rentInfoID', r.rent_info_id,
          'startDate', r.start_date,
          'maturityDate', r.maturity_date,
          'renewalDate', r.renewal_date,
          'landlordID', r.landlord_id,
          'tenantID', r.tenant_id,
          'address', r.address,
          'rentFee', r.rent_fee
        ))
      END AS rent_info
    FROM tenants t LEFT JOIN rent_info r ON t.tenant_id = r.tenant_id 
    WHERE t.tenant_id = $1 GROUP BY t.tenant_id, t.phone;
  `

	row := q.db.QueryRowContext(ctx, query, id)

	var rentInfo []byte
	var additionalInfo []byte
	var tenant entity.Tenant

	err := row.Scan(
		&tenant.TenantID,
		&tenant.FirstName,
		&tenant.LastName,
		&tenant.Gender,
		&tenant.DOB,
		&tenant.Image,
		&tenant.Email,
		&tenant.Phone,
		&tenant.StateOfOrigin,
		&tenant.Nationality,
		&tenant.Occupation,
		&additionalInfo,
		&tenant.RegisteredBy,
		&tenant.CreatedAt,
		&tenant.UpdatedAt,
		&rentInfo,
	)

	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(rentInfo, &tenant.RentInfo); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(additionalInfo, &tenant.AdditionalInfo); err != nil {
		return nil, err
	}

	return &tenant, nil
}

func (q *queries) TenantTotalCount(ctx context.Context) (int64, error) {
	const query = "SELECT COUNT(*) FROM tenants;"

	row := q.db.QueryRowContext(ctx, query)

	var total int64

	err := row.Scan(&total)
	if err != nil {
		return 0, err
	}

	return total, nil
}

func (q *queries) DeleteTenant(ctx context.Context, id uuid.UUID) error {
	const query = `DELETE FROM tenants WHERE tenant_id = $1;`

	if _, err := q.db.ExecContext(ctx, query, id); err != nil {
		return err
	}

	return nil
}

func (q *queries) FindAllTenants(ctx context.Context) ([]*entity.TenantOut, error) {
	const query = `
    SELECT t.tenant_id, t.first_name, t.last_name, t.gender, t.dob, 
      t.image, t.email, t.phone, t.state_of_origin, t.nationality, t.occupation,
      t.additional_info, r.start_date, r.maturity_date, r.renewal_date,
      r.address, r.rent_fee, r.landlord_id, t.created_at, t.updated_at
    FROM tenants t LEFT JOIN rent_info r ON t.tenant_id = r.tenant_id;
  `
	rows, err := q.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	tenants := []*entity.TenantOut{}

	for rows.Next() {
		var additionalInfo []byte
		var tenant entity.TenantOut

		err := rows.Scan(
			&tenant.TenantID,
			&tenant.FirstName,
			&tenant.LastName,
			&tenant.Gender,
			&tenant.DOB,
			&tenant.Image,
			&tenant.Email,
			&tenant.Phone,
			&tenant.StateOfOrigin,
			&tenant.Nationality,
			&tenant.Occupation,
			&additionalInfo,
			&tenant.StartDate,
			&tenant.MaturityDate,
			&tenant.RenewalDate,
			&tenant.Address,
			&tenant.RentFee,
			&tenant.LandlordID,
			&tenant.CreatedAt,
			&tenant.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(additionalInfo, &tenant.AdditionalInfo); err != nil {
			return nil, err
		}

		tenants = append(tenants, &tenant)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if err := rows.Close(); err != nil {
		return nil, err
	}

	return tenants, nil
}
