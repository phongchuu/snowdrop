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
	Timestamp  time.Time  `json:"timestamp"`
	Data       T          `json:"data,omitzero"`
	Pagination Pagination `json:"pagination,omitzero"`
}

type ResponseBuilder struct {
	w         http.ResponseWriter
	r         *http.Request
	localizer *i18n.Localizer
	httpCode  int
	result    Response[any]
}

func NewResponseBuilder(w http.ResponseWriter, r *http.Request) *ResponseBuilder {
	return &ResponseBuilder{
		w:         w,
		r:         r,
		localizer: trans.GetLocalizer(r),
		httpCode:  http.StatusOK,
		result:    Response[any]{},
	}
}

func (r *ResponseBuilder) setDefaults() {
	r.result.Timestamp = time.Now()

	if r.result.Status == 0 {
		r.result.Status = r.httpCode
	}

	if r.result.Message == "" {
		r.result.Message = r.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: strconv.Itoa(r.result.Status),
		})
	}
}

func (r *ResponseBuilder) Status(code int) *ResponseBuilder {
	r.httpCode = code
	r.result.Status = code

	return r
}

func (r *ResponseBuilder) Message(messageID string, templateDataMaps ...map[string]any) *ResponseBuilder {
	translatedMessge, err := r.localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    messageID,
		TemplateData: lo.Assign(templateDataMaps...),
	})

	if err != nil {
		r.result.Message = messageID
	} else {
		r.result.Message = translatedMessge
	}

	return r
}

func (r *ResponseBuilder) Data(data any) *ResponseBuilder {
	r.result.Data = data
	return r
}

func (r *ResponseBuilder) Pagination(totalRows int) *ResponseBuilder {
	pageSize := GetPageSize(r.r)

	r.result.Pagination = Pagination{
		Page:         GetPage(r.r),
		PageSize:     pageSize,
		TotalRecords: totalRows,
		TotalPages:   int(math.Ceil(float64(totalRows) / float64(pageSize))),
	}

	return r
}

func (r *ResponseBuilder) NoContent() {
	r.w.WriteHeader(http.StatusNoContent)
}

func (r *ResponseBuilder) JSON() {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)

	r.setDefaults()

	if err := enc.Encode(r.result); err != nil {
		http.Error(r.w, err.Error(), http.StatusInternalServerError)
		return
	}

	r.w.Header().Set("Content-Type", "application/json")
	r.w.Header().Set("X-Content-Type-Options", "nosniff")
	r.w.WriteHeader(r.result.Status)
	_, _ = r.w.Write(buf.Bytes())
}

func (r *ResponseBuilder) HTML(html string) {
	r.w.Header().Set("Content-Type", "text/html")
	r.w.Header().Set("X-Content-Type-Options", "nosniff")
	r.w.WriteHeader(r.httpCode)
	_, _ = r.w.Write([]byte(html))
}
