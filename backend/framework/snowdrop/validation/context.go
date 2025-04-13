package validation

import (
	"context"
	"net/http"

	ut "github.com/go-playground/universal-translator"
)

type universalTranslatorContextKey int

const universalTranslatorIdentifier universalTranslatorContextKey = iota

// GetValidationTranslator retrieves the universal translator from request context.
//
//nolint:ireturn
func GetValidationTranslator(r *http.Request) (ut.Translator, error) {
	if translator, ok := r.Context().Value(universalTranslatorIdentifier).(ut.Translator); ok {
		return translator, nil
	}

	return nil, ErrNoUniversalTranslator
}

// WithUniversalTranslator returns a new http.Request with the provided translator stored in its context.
// The translator can be later retrieved using the GetValidationTranslator function.
func WithUniversalTranslator(r *http.Request, translator ut.Translator) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), universalTranslatorIdentifier, translator))
}
