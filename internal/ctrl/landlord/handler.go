package landlord

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
		r.Post("/", ctrl.createLandlord())
		r.Get("/", ctrl.findLandlords())
		r.Get("/{id}", ctrl.findLandlord())
		r.Delete("/{id}", ctrl.deleteLandlord())
		r.Get("/count", ctrl.landlordTotal())
		r.Put("/{id}/info", ctrl.addPropertyInfo())
		r.Get("/xlsx", ctrl.landlordXlsx())
	})
}

func (ctrl *Ctrl) createLandlord() http.HandlerFunc {
	return handlerlib.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		in, err := handlerlib.Bind[entity.LandlordIn](w, r)
		if err != nil {
			return handlerlib.NewError(422, err.Error())
		}

		v := validator.New()

		if entity.ValidateLandlordIn(v, in); !v.Valid() {
			return handlerlib.WriteJson(w, 422, v.Err())
		}

		landlord, err := ctrl.create(r.Context(), handlerlib.GetCtxUser(r), in)

		if err != nil && errors.Is(err, ErrDuplicateKey) {
			return handlerlib.NewError(409, "phone already in use")
		}

		if err != nil {
			return err
		}

		return handlerlib.WriteJson(w, 201, landlord)
	})
}

func (ctrl *Ctrl) findLandlord() http.HandlerFunc {
	return handlerlib.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		landlordID, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			return handlerlib.NewError(400, "invalid landlord id")
		}

		landlord, err := ctrl.findOne(r.Context(), landlordID)

		if err != nil && errors.Is(err, ErrNotFound) {
			return handlerlib.NewError(404, "landlord not found")
		}

		if err != nil {
			return err
		}

		return handlerlib.WriteJson(w, 200, landlord)
	})
}

func (ctrl *Ctrl) findLandlords() http.HandlerFunc {
	return handlerlib.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		filterParam := FilterParam{
			firstName: handlerlib.GetQuery(r, "first_name", ""),
			phone:     handlerlib.GetQuery(r, "phone", ""),
			address:   handlerlib.GetQuery(r, "address", ""),
		}

		paginator := handlerlib.NewPaginator(
			handlerlib.GetQueryInt(r, "page", 1),
			handlerlib.GetQueryInt(r, "page_size", 10),
		)

		landlords, err := ctrl.findAll(r.Context(), filterParam, paginator)
		if err != nil {
			return err
		}

		return handlerlib.WriteJson(w, 200, map[string]any{
			"metadata": paginator.GetMetadata(),
			"data":     landlords,
		})
	})
}

func (ctrl *Ctrl) deleteLandlord() http.HandlerFunc {
	return handlerlib.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			return handlerlib.NewError(400, "invalid landlord id")
		}

		if err := ctrl.deleteOne(r.Context(), id); err != nil {
			return err
		}

		return handlerlib.SendStatus(w, 204)
	})
}

func (ctrl *Ctrl) landlordTotal() http.HandlerFunc {
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

func (ctrl *Ctrl) addPropertyInfo() http.HandlerFunc {
	return handlerlib.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			return handlerlib.NewError(400, "invalid landord id")
		}

		in, err := handlerlib.Bind[entity.PropertyInfoIn](w, r)
		if err != nil {
			return handlerlib.NewError(422, err.Error())
		}

		v := validator.New()

		if entity.ValidatePropertyInfoIn(v, in); !v.Valid() {
			return handlerlib.WriteJson(w, 422, v.Err())
		}

		info, err := ctrl.createPropertyInfo(r.Context(), id, in)
		if err != nil {
			return err
		}

		return handlerlib.WriteJson(w, 200, info)
	})
}

func (ctrl *Ctrl) landlordXlsx() http.HandlerFunc {
	return handlerlib.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		landlords, err := ctrl.getAll(r.Context())
		if err != nil {
			return err
		}

		file := excelize.NewFile()

		headers := []string{
			"Firstname",
			"Lastname",
			"Email",
			"Phone",
			"Address",
			"Property Type",
			"Flat No",
			"Lease Price",
			"Lease Period",
			"Start Date",
			"End Date",
		}

		for i, header := range headers {
			file.SetCellValue("Sheet1", fmt.Sprintf("%s%d", string(rune(65+i)), 1), header)
		}

		data := make([][]any, len(landlords))

		for i, landlord := range landlords {
			data[i] = []any{
				landlord.FirstName,
				landlord.LastName,
				landlord.Email,
				landlord.Phone,
				landlord.Address,
				landlord.PropertyType.String(),
				landlord.AdditionalInfo["flatNo"],
				landlord.LeasePrice,
				landlord.LeasePeriod,
				landlord.StartDate.Format("02/01/2006"),
				landlord.EndDate.Format("02/01/2006"),
			}
		}

		for i, row := range data {
			dr := i + 2
			for j, col := range row {
				file.SetCellValue("Sheet1", fmt.Sprintf("%s%d", string(rune(65+j)), dr), col)
			}
		}

		temp, err := os.CreateTemp("", "landlords.xlsx")
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
