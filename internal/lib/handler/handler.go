package handlerlib

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"strconv"

	"github.com/emma769/a-realtor/internal/entity"
)

type RespMsg struct {
	Message string `json:"message"`
}

type ErrResp struct {
	Detail string `json:"error"`
}

type HandlerError struct {
	msg  string
	code int
}

func (e HandlerError) Error() string {
	return e.msg
}

func NewError(code int, msg string) *HandlerError {
	return &HandlerError{
		msg,
		code,
	}
}

type HandlerFunc func(http.ResponseWriter, *http.Request) error

func Wrap(fn HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var he *HandlerError

		err := fn(w, r)

		if err != nil && errors.As(err, &he) {
			if err := WriteJson(w, he.code, ErrResp{Detail: he.msg}); err != nil {
				panic(err)
			}

			return
		}

		if err != nil {
			panic(err)
		}
	}
}

func WriteJson[T any](w http.ResponseWriter, code int, data T) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	return json.NewEncoder(w).Encode(data)
}

func Bind[T any](w http.ResponseWriter, r *http.Request) (T, error) {
	var t T

	defer func() {
		if err := r.Body.Close(); err != nil {
			log.Println(err)
		}
	}()

	body := http.MaxBytesReader(w, r.Body, 1_048_576)
	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(&t)

	var synerr *json.SyntaxError
	var typeerr *json.UnmarshalTypeError
	var invaliderr *json.InvalidUnmarshalError

	if err != nil && errors.As(err, &synerr) {
		return t, fmt.Errorf("invalid json at position %d", synerr.Offset)
	}

	if err != nil && errors.As(err, &typeerr) {
		if typeerr.Field != "" {
			return t, fmt.Errorf("invalid json at %s", typeerr.Field)
		}
		return t, fmt.Errorf("invalid json at position %d", typeerr.Offset)
	}

	if err != nil && errors.As(err, &invaliderr) {
		panic(err)
	}

	if err != nil && errors.Is(err, io.EOF) {
		return t, fmt.Errorf("request body has no content")
	}

	if err != nil && errors.Is(err, io.ErrUnexpectedEOF) {
		return t, fmt.Errorf("malformed json")
	}

	if err != nil {
		return t, err
	}

	return t, nil
}

func SendStatus(w http.ResponseWriter, code int) error {
	w.WriteHeader(code)
	return nil
}

type Paginator struct {
	page, size, total int
}

func NewPaginator(page, size int) *Paginator {
	return &Paginator{
		page: page,
		size: size,
	}
}

func (p *Paginator) SetTotal(total int) {
	p.total = total
}

func (p Paginator) Limit() int {
	return p.size
}

func (p Paginator) Offset() int {
	return (p.page - 1) * p.size
}

type PageMetadata struct {
	FirstPage   int  `json:"firstPage"`
	LastPage    int  `json:"lastPage"`
	CurrentPage int  `json:"currentPage"`
	Total       int  `json:"total"`
	PageSize    int  `json:"pageSize"`
	HasNextPage bool `json:"hasNextPage"`
}

func (p Paginator) GetMetadata() *PageMetadata {
	lastPage := int(math.Ceil(float64(p.total) / float64(p.size)))

	return &PageMetadata{
		FirstPage:   1,
		LastPage:    lastPage,
		CurrentPage: p.page,
		Total:       p.total,
		PageSize:    p.size,
		HasNextPage: hasNextPage(p.page, p.total, p.size, lastPage),
	}
}

func hasNextPage(currentPage, total, pageSize, lastPage int) bool {
	if total <= 0 || pageSize <= 0 {
		return false
	}
	return currentPage < lastPage
}

func GetQuery(r *http.Request, name string, fallback string) string {
	values, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return fallback
	}

	return cmp.Or(values.Get(name), fallback)
}

func GetQueryInt(r *http.Request, name string, fallback int) int {
	values, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return fallback
	}

	data, err := strconv.Atoi(values.Get(name))
	if err != nil {
		return fallback
	}

	return data
}

func SetCtxUser(r *http.Request, user *entity.User) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), entity.UserCtxKey, user))
}

func GetCtxUser(r *http.Request) *entity.User {
	user, ok := r.Context().Value(entity.UserCtxKey).(*entity.User)

	if !ok {
		panic("no user in request context")
	}

	return user
}
