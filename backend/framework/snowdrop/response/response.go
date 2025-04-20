package response

import (
	"encoding/json"
	"errors"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/samber/lo"
	"internal.snowdrop/framework/request"
	"internal.snowdrop/framework/translation"
	"internal.snowdrop/framework/validation"
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

type Builder struct {
	w                    http.ResponseWriter
	r                    *http.Request
	localizer            *i18n.Localizer
	validationTranslator ut.Translator
	httpCode             int
	result               Response[any]
}

func NewBuilder(w http.ResponseWriter, r *http.Request) *Builder {
	localizer, err := translation.GetLocalizer(r)
	if err != nil {
		panic(err)
	}

	validationTranslator, err := validation.GetValidationTranslator(r)
	if err != nil {
		panic(err)
	}

	return &Builder{
		w:                    w,
		r:                    r,
		localizer:            localizer,
		validationTranslator: validationTranslator,
		httpCode:             http.StatusOK,
		result:               Response[any]{},
	}
}

func (r *Builder) setDefaults() {
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

func (r *Builder) Status(code int) *Builder {
	r.httpCode = code
	r.result.Status = code

	return r
}

func (r *Builder) Message(
	messageID string,
	templateDataMaps ...map[string]any,
) *Builder {
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

func (r *Builder) Errors(errs error) *Builder {
	var validationErrs validator.ValidationErrors
	if errors.As(errs, &validationErrs) {
		r.result.Data = validationErrs.Translate(r.validationTranslator)
		r.Status(http.StatusBadRequest)

		return r
	}

	var maxBytesErr *http.MaxBytesError
	if errors.As(errs, &maxBytesErr) {
		r.Status(http.StatusRequestEntityTooLarge)

		return r
	}

	return r
}

func (r *Builder) Data(data any) *Builder {
	r.result.Data = data

	return r
}

func (r *Builder) Pagination(totalRows int) *Builder {
	pageSize := request.GetPageSize(r.r)

	r.result.Pagination = Pagination{
		Page:         request.GetPage(r.r),
		PageSize:     pageSize,
		TotalRecords: totalRows,
		TotalPages:   int(math.Ceil(float64(totalRows) / float64(pageSize))),
	}

	return r
}

func (r *Builder) NoContent() {
	r.w.WriteHeader(http.StatusNoContent)
}

func (r *Builder) JSON() {
	r.setDefaults()

	r.w.Header().Set("Content-Type", "application/json")
	r.w.Header().Set("X-Content-Type-Options", "nosniff")
	r.w.WriteHeader(r.result.Status)

	enc := json.NewEncoder(r.w)
	if err := enc.Encode(r.result); err != nil {
		log.Printf("failed to encode JSON response: %v", err)

		return
	}
}

func (r *Builder) HTML(html string) {
	r.w.Header().Set("Content-Type", "text/html")
	r.w.Header().Set("X-Content-Type-Options", "nosniff")
	r.w.WriteHeader(r.httpCode)
	_, _ = r.w.Write([]byte(html))
}
