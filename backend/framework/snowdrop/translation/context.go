package translation

import (
	"context"
	"errors"
	"net/http"

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

type localizerCtxKey string

const localizerCtxID localizerCtxKey = "LocalizerCtxID"

var ErrNoLocalizer = errors.New("there is no *i18n.Localizer in the given context")

func WithLocalizer(r *http.Request, localizer *i18n.Localizer) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), localizerCtxID, localizer))
}

func GetLocalizer(r *http.Request) (*i18n.Localizer, error) {
	if localizer, ok := r.Context().Value(localizerCtxID).(*i18n.Localizer); ok {
		return localizer, nil
	}

	return nil, ErrNoLocalizer
}
