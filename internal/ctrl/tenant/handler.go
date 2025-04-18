package tenant

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"

	"github.com/emma769/a-realtor/internal/entity"
	funclib "github.com/emma769/a-realtor/internal/lib/func"
	handlerlib "github.com/emma769/a-realtor/internal/lib/handler"
	"github.com/emma769/a-realtor/internal/middleware"
	"github.com/emma769/a-realtor/internal/validator"
)

const timeout = 5 * time.Second

type Ctrl struct {
	*Service
	logger *slog.Logger
}

func New(store storer, logger *slog.Logger) *Ctrl {
	return &Ctrl{
		Service: &Service{
			store,
			timeout,
		},
		logger: logger,
	}
}

func (ctrl Ctrl) Routes(r chi.Router) {
	r.With(middleware.RequireAuth).Group(func(r chi.Router) {
		r.Post("/", ctrl.createTenant())
		r.Get("/", ctrl.findTenants())
		r.Get("/{id}", ctrl.findTenant())
		r.Get("/count", ctrl.tenantTotal())
		r.Delete("/{id}", ctrl.deleteTenant())
		r.Put("/{id}/info", ctrl.addRentInfo())
		r.Get("/xlsx", ctrl.tenantXlsx())
	})
}

func (ctrl *Ctrl) createTenant() http.HandlerFunc {
	return handlerlib.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		in, err := handlerlib.Bind[entity.TenantIn](w, r)
		if err != nil {
			return handlerlib.NewError(422, err.Error())
		}

		v := validator.New()

		if entity.ValidateTenantIn(v, in); !v.Valid() {
			return handlerlib.WriteJson(w, 422, v.Err())
		}

		tenant, err := ctrl.create(r.Context(), handlerlib.GetCtxUser(r), in)

		if err != nil && errors.Is(err, ErrDuplicateKey) {
			return handlerlib.NewError(409, "phone already in use")
		}

		if err != nil {
			return err
		}

		return handlerlib.WriteJson(w, 201, tenant)
	})
}

func (ctrl *Ctrl) findTenants() http.HandlerFunc {
	return handlerlib.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		filterParam := &FilterParam{
			firstName: handlerlib.GetQuery(r, "first_name", ""),
			phone:     handlerlib.GetQuery(r, "phone", ""),
			address:   handlerlib.GetQuery(r, "address", ""),
		}

		paginator := handlerlib.NewPaginator(
			handlerlib.GetQueryInt(r, "page", 1),
			handlerlib.GetQueryInt(r, "page_size", 10),
		)

		tenants, err := ctrl.findall(r.Context(), filterParam, paginator)
		if err != nil {
			return err
		}

		return handlerlib.WriteJson(w, 200, map[string]any{
			"data":     tenants,
			"metadata": paginator.GetMetadata(),
		})
	})
}

func (ctrl *Ctrl) findTenant() http.HandlerFunc {
	return handlerlib.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			return handlerlib.NewError(400, "invalid tenant id")
		}

		tenant, err := ctrl.findone(r.Context(), id)

		if err != nil && errors.Is(err, ErrNotFound) {
			return handlerlib.NewError(404, "tenant not found")
		}

		if err != nil {
			return err
		}

		return handlerlib.WriteJson(w, 200, tenant)
	})
}

func (ctrl *Ctrl) tenantTotal() http.HandlerFunc {
	return handlerlib.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		total, err := ctrl.total(r.Context())
		if err != nil {
			return err
		}

		return handlerlib.WriteJson(w, 200, map[string]int64{
			"total": total,
		})
	})
}

func (ctrl *Ctrl) deleteTenant() http.HandlerFunc {
	return handlerlib.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			return handlerlib.NewError(400, "invalid tenant id")
		}

		if err := ctrl.delete(r.Context(), id); err != nil {
			return err
		}

		return handlerlib.SendStatus(w, 204)
	})
}

func (ctrl *Ctrl) addRentInfo() http.HandlerFunc {
	return handlerlib.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			return handlerlib.NewError(400, "invalid tenant id")
		}

		in, err := handlerlib.Bind[entity.RentInfoIn](w, r)
		if err != nil {
			return handlerlib.NewError(422, err.Error())
		}

		v := validator.New()

		if entity.ValidateRentInfoIn(v, in); !v.Valid() {
			return handlerlib.WriteJson(w, 422, v.Err())
		}

		info, err := ctrl.createRentInfo(r.Context(), id, in)
		if err != nil {
			return err
		}

		return handlerlib.WriteJson(w, 200, info)
	})
}

func (ctrl *Ctrl) tenantXlsx() http.HandlerFunc {
	return handlerlib.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		tenants, err := ctrl.getAll(r.Context())
		if err != nil {
			return err
		}

		file := excelize.NewFile()

		headers := []string{
			"Firstname",
			"Lastname",
			"Phone",
			"Rent Amount",
			"Address",
			"Landlord/Investor",
			"Landlord/Investor Phone",
			"Duration",
			"Start Date",
			"Maturity Date",
			"Renewal Date",
		}

		for i, header := range headers {
			file.SetCellValue("Sheet1", fmt.Sprintf("%s%d", string(rune(65+i)), 1), header)
		}

		data := make([][]any, len(tenants))

		for i, tenant := range tenants {
			landlord, _ := ctrl.getLandlord(r.Context(), tenant.LandlordID)

			var landlordName string

			if landlord != nil {
				landlordName = fmt.Sprintf("%s %s", landlord.FirstName, landlord.LastName)
			}

			data[i] = []any{
				tenant.FirstName,
				tenant.LastName,
				tenant.Phone,
				tenant.RentFee,
				tenant.Address,
				landlordName,
				landlord.Phone,
				funclib.DaysBetween(tenant.StartDate, tenant.RenewalDate),
				tenant.StartDate.Format("02/01/2006"),
				tenant.MaturityDate.Format("02/01/2006"),
				tenant.RenewalDate.Format("02/01/2006"),
			}
		}

		for i, row := range data {
			dr := i + 2
			for j, col := range row {
				file.SetCellValue("Sheet1", fmt.Sprintf("%s%d", string(rune(65+j)), dr), col)
			}
		}

		temp, err := os.CreateTemp("", "tenants.xlsx")
		if err != nil {
			return err
		}

		defer func() {
			if err := os.Remove(temp.Name()); err != nil {
				ctrl.logger.ErrorContext(
					r.Context(),
					"could not remove temp file",
					"detail",
					err.Error(),
				)
			}
		}()

		defer func() {
			if err := temp.Close(); err != nil {
				ctrl.logger.ErrorContext(
					r.Context(),
					"could not close temp file",
					"detail",
					err.Error(),
				)
			}
		}()

		if _, err := file.WriteTo(temp); err != nil {
			return err
		}

		w.Header().
			Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		w.Header().Set("Content-Disposition", "attachment; filename="+temp.Name())

		http.ServeFile(w, r, temp.Name())

		return nil
	})
}
