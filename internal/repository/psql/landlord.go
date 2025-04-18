package psql

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/emma769/a-realtor/internal/entity"
	"github.com/emma769/a-realtor/internal/repository"
)

type LandlordParam interface {
	FirstName() string
	LastName() string
	Email() string
	Phone() string
	RegisteredBy() uuid.UUID
}

type PropertyInfoParam interface {
	Address() string
	PropertyType() entity.PropertyType
	LeasePeriod() int
	LeasePrice() float64
	StartDate() time.Time
	EndDate() time.Time
	AdditionalInfo() []byte
}

type LandlordWithPropertyInfoParam interface {
	LandlordParam
	PropertyInfoParam
}

func (q *queries) CreateLandlord(
	ctx context.Context,
	param LandlordWithPropertyInfoParam,
) (*entity.Landlord, error) {
	landlord, err := q.createLandlord(ctx, param)
	if err != nil {
		return nil, err
	}

	propertyInfo, err := q.CreatePropertyInfo(ctx, landlord.LandlordID, param)
	if err != nil {
		return nil, err
	}

	landlord.PropertyInfo = append(landlord.PropertyInfo, propertyInfo)

	return landlord, nil
}

func (q *queries) createLandlord(
	ctx context.Context,
	param LandlordParam,
) (*entity.Landlord, error) {
	const query = `
    INSERT INTO landlords (
      first_name, last_name, email, phone, registered_by
    ) VALUES ($1, $2, $3, $4, $5) 
    RETURNING landlord_id, first_name, last_name, email, 
      phone, registered_by, created_at, updated_at;
  `
	row := q.db.QueryRowContext(
		ctx,
		query,
		param.FirstName(),
		param.LastName(),
		param.Email(),
		param.Phone(),
		param.RegisteredBy(),
	)

	var landlord entity.Landlord

	err := row.Scan(
		&landlord.LandlordID,
		&landlord.FirstName,
		&landlord.LastName,
		&landlord.Email,
		&landlord.Phone,
		&landlord.RegisteredBy,
		&landlord.CreatedAt,
		&landlord.UpdatedAt,
	)

	if err != nil && strings.Contains(err.Error(), "duplicate") {
		return nil, repository.ErrDuplicateKey
	}

	return &landlord, err
}

func (q *queries) CreatePropertyInfo(
	ctx context.Context,
	landlordID uuid.UUID,
	param PropertyInfoParam,
) (*entity.PropertyInfo, error) {
	const query = `
    INSERT INTO property_info (
      address, property_type, additional_info, lease_price, 
      lease_period, start_date, end_date, landlord_id
    ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    RETURNING 
      property_info_id, address, property_type, additional_info, 
      lease_price, lease_period, start_date, end_date, landlord_id;
  `
	row := q.db.QueryRowContext(
		ctx,
		query,
		param.Address(),
		param.PropertyType(),
		param.AdditionalInfo(),
		param.LeasePrice(),
		param.LeasePeriod(),
		param.StartDate(),
		param.EndDate(),
		landlordID,
	)

	var additionalInfo []byte
	var propertyInfo entity.PropertyInfo

	err := row.Scan(
		&propertyInfo.PropertyInfoID,
		&propertyInfo.Address,
		&propertyInfo.PropertyType,
		&additionalInfo,
		&propertyInfo.LeasePrice,
		&propertyInfo.LeasePeriod,
		&propertyInfo.StartDate,
		&propertyInfo.EndDate,
		&propertyInfo.LandlordID,
	)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(additionalInfo, &propertyInfo.AdditionalInfo); err != nil {
		return nil, err
	}

	return &propertyInfo, nil
}

func (q *queries) FindLandlord(
	ctx context.Context,
	id uuid.UUID,
) (*entity.Landlord, error) {
	const query = `
    SELECT l.landlord_id, l.first_name, l.last_name, l.email, 
      l.phone, l.registered_by, l.created_at, l.updated_at, 
      CASE WHEN count(p.property_info_id) = 0 THEN 
        '[]'::JSON
      ELSE
        json_agg(
          json_build_object(
            'propertyInfoID', p.property_info_id,
            'address', p.address,
            'propertyType', p.property_type,
            'additionalInfo', p.additional_info,
            'leasePrice', p.lease_price,
            'leasePeriod', p.lease_period,
            'startDate', p.start_date,
            'endDate', p.end_date,
            'landlordID', p.landlord_id
          )
        )
      END AS property_info 
    FROM landlords l LEFT JOIN property_info p ON l.landlord_id = p.landlord_id 
    WHERE l.landlord_id = $1 GROUP BY l.landlord_id, l.phone;
  `
	row := q.db.QueryRowContext(ctx, query, id)

	var propertyInfo []byte
	var landlord entity.Landlord

	err := row.Scan(
		&landlord.LandlordID,
		&landlord.FirstName,
		&landlord.LastName,
		&landlord.Email,
		&landlord.Phone,
		&landlord.RegisteredBy,
		&landlord.CreatedAt,
		&landlord.UpdatedAt,
		&propertyInfo,
	)

	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(propertyInfo, &landlord.PropertyInfo); err != nil {
		return nil, err
	}
	return &landlord, nil
}

type LandlordFilterParam interface {
	FirstName() string
	Phone() string
	Address() string
}

func (q *queries) FindLandlords(
	ctx context.Context,
	filterParam LandlordFilterParam,
	paginator PaginationParam,
) ([]*entity.LandlordOut, error) {
	const query = `
    SELECT COUNT(*) OVER(), 
      l.landlord_id, l.first_name, l.last_name, l.email, l.phone, l.registered_by, 
      p.address, p.property_type, p.lease_price, p.lease_period, p.start_date, 
      p.end_date, p.additional_info, l.created_at, l.updated_at
    FROM landlords l LEFT JOIN property_info p ON l.landlord_id = p.landlord_id
    WHERE 
      (LOWER(l.first_name) = LOWER($1) OR $1 = '') 
      AND (l.phone = $2 OR $2 = '')
      AND (to_tsvector('simple', p.address) @@ plainto_tsquery('simple', $3) OR $3 = '')
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
	landlords := []*entity.LandlordOut{}

	for rows.Next() {
		var additionalInfo []byte
		var landlord entity.LandlordOut

		err := rows.Scan(
			&total,
			&landlord.LandlordID,
			&landlord.FirstName,
			&landlord.LastName,
			&landlord.Email,
			&landlord.Phone,
			&landlord.RegisteredBy,
			&landlord.Address,
			&landlord.PropertyType,
			&landlord.LeasePrice,
			&landlord.LeasePeriod,
			&landlord.StartDate,
			&landlord.EndDate,
			&additionalInfo,
			&landlord.CreatedAt,
			&landlord.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(additionalInfo, &landlord.AdditionalInfo); err != nil {
			return nil, err
		}

		landlords = append(landlords, &landlord)
	}

	paginator.SetTotal(total)

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if err := rows.Close(); err != nil {
		return nil, err
	}

	return landlords, nil
}

func (q *queries) UpdateLandlord(
	ctx context.Context,
	param LandlordParam,
) (*entity.Landlord, error) {
	return nil, nil
}

func (q *queries) DeleteLandlord(ctx context.Context, id uuid.UUID) error {
	const query = `DELETE FROM landlords WHERE landlord_id = $1;`

	if _, err := q.db.ExecContext(ctx, query, id); err != nil {
		return err
	}

	return nil
}

func (q *queries) TotalLandlordCount(ctx context.Context) (int64, error) {
	const query = "SELECT COUNT(*) FROM landlords;"

	row := q.db.QueryRowContext(ctx, query)

	var total int64

	err := row.Scan(&total)
	if err != nil {
		return 0, err
	}

	return total, nil
}

func (q *queries) GetAllLandlords(ctx context.Context) ([]*entity.LandlordOut, error) {
	const query = `
    SELECT  
      l.landlord_id, l.first_name, l.last_name, l.email, l.phone, l.registered_by, 
      p.address, p.property_type, p.lease_price, p.lease_period, p.start_date, 
      p.end_date, p.additional_info, l.created_at, l.updated_at
    FROM landlords l LEFT JOIN property_info p ON l.landlord_id = p.landlord_id;
  `

	rows, err := q.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	landlords := []*entity.LandlordOut{}

	for rows.Next() {
		var additionalInfo []byte
		var landlord entity.LandlordOut

		err := rows.Scan(
			&landlord.LandlordID,
			&landlord.FirstName,
			&landlord.LastName,
			&landlord.Email,
			&landlord.Phone,
			&landlord.RegisteredBy,
			&landlord.Address,
			&landlord.PropertyType,
			&landlord.LeasePrice,
			&landlord.LeasePeriod,
			&landlord.StartDate,
			&landlord.EndDate,
			&additionalInfo,
			&landlord.CreatedAt,
			&landlord.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(additionalInfo, &landlord.AdditionalInfo); err != nil {
			return nil, err
		}

		landlords = append(landlords, &landlord)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if err := rows.Close(); err != nil {
		return nil, err
	}

	return landlords, nil
}
