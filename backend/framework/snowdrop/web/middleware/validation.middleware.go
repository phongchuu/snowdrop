package middleware

import (
	"net/http"

	ut "github.com/go-playground/universal-translator"
	snowdrop "internal.snowdrop/framework"
	"internal.snowdrop/framework/validation"
)

func NewUniversalTranslatorMiddleware(
	universalTranslator *ut.UniversalTranslator,
	getPreferredUserLanguageFn snowdrop.GetPreferredUserLanguageFn,
) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			language := getPreferredUserLanguageFn(r)
			translator, _ := universalTranslator.GetTranslator(language.String())
			next.ServeHTTP(w, validation.WithUniversalTranslator(r, translator))
		})
	}
}
