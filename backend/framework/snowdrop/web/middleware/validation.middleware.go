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
			tracer := snowdrop.MustGetTracer(r)

			spanCtx, span := tracer.Start(r.Context(), "UniversalTranslatorMiddleware")
			defer span.End()

			req := r.WithContext(spanCtx)
			language := getPreferredUserLanguageFn(req)
			translator, _ := universalTranslator.GetTranslator(language.String())
			next.ServeHTTP(w, validation.WithUniversalTranslator(req, translator))
		})
	}
}
