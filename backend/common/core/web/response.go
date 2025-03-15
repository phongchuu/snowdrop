package web

import (
	"bytes"
	"encoding/json"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/samber/lo"
	"internal.snowdrop/common/core/trans"
)

type Pagination struct {
	Page         int `json:"page"`
	PageSize     int `json:"pageSize"`
	TotalRecords int `json:"totalRecords"`
	TotalPages   int `json:"totalPages"`
}

type Response[T any] struct {
	Status     int        `json:"status"`
	Message    string     `json:"message"`
	Data       T          `json:"data,omitzero"`
	Pagination Pagination `json:"pagination,omitzero"`
	Timestamp  time.Time  `json:"timestamp"`
}

type ResponseBuilder[T any] struct {
	w         http.ResponseWriter
	r         *http.Request
	localizer *i18n.Localizer
	result    Response[T]
}

func NewResponseBuilder(w http.ResponseWriter, r *http.Request) *ResponseBuilder[any] {
	return &ResponseBuilder[any]{
		w:         w,
		r:         r,
		result:    Response[any]{},
		localizer: trans.GetLocalizer(r),
	}
}

func NewGenericResponseBuilder[T any](w http.ResponseWriter, r *http.Request) *ResponseBuilder[T] {
	return &ResponseBuilder[T]{
		w:      w,
		r:      r,
		result: Response[T]{},
	}
}

func (r *ResponseBuilder[T]) setDefaults() {
	r.result.Timestamp = time.Now()

	if r.result.Status == 0 {
		r.result.Status = http.StatusOK
	}

	if r.result.Message == "" {
		r.result.Message = r.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: strconv.Itoa(r.result.Status),
		})
	}
}

func (r *ResponseBuilder[T]) Status(code int) *ResponseBuilder[T] {
	r.w.WriteHeader(code)
	r.result.Status = code

	return r
}

func (r *ResponseBuilder[T]) Message(messageID string, templateDataMaps ...map[string]any) *ResponseBuilder[T] {
	r.result.Message = r.localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID:    messageID,
		TemplateData: lo.Assign(templateDataMaps...),
	})

	return r
}

func (r *ResponseBuilder[T]) Data(data T) *ResponseBuilder[T] {
	r.result.Data = data
	return r
}

func (r *ResponseBuilder[T]) Pagination(totalRows int) *ResponseBuilder[T] {
	pageSize := GetPageSize(r.r)

	r.result.Pagination = Pagination{
		Page:         GetPage(r.r),
		PageSize:     pageSize,
		TotalRecords: totalRows,
		TotalPages:   int(math.Ceil(float64(totalRows) / float64(pageSize))),
	}

	return r
}

func (r *ResponseBuilder[T]) JSON() {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)

	r.setDefaults()

	if err := enc.Encode(r.result); err != nil {
		http.Error(r.w, err.Error(), http.StatusInternalServerError)
		return
	}

	r.w.Header().Set("Content-Type", "application/json")
	r.w.Header().Set("X-Content-Type-Options", "nosniff")
	_, _ = r.w.Write(buf.Bytes())
}

func (r *ResponseBuilder[T]) HTML(html string) {
	r.w.Header().Set("Content-Type", "text/html")
	r.w.Header().Set("X-Content-Type-Options", "nosniff")
	_, _ = r.w.Write([]byte(html))
}
