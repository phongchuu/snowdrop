package middleware

import (
	"net/http"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	snowdrop "internal.snowdrop/framework"
	"internal.snowdrop/framework/translation"
)

// NewI18nMiddleware is a middleware function that adds internationalization support to an HTTP server.
// It uses the provided i18n.Bundle to create a localizer based on the language specified in the request.
func NewI18nMiddleware(
	bundle *i18n.Bundle,
	getPreferredUserLanguageFn snowdrop.GetPreferredUserLanguageFn,
) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tracer := snowdrop.MustGetTracer(r)

			spanCtx, span := tracer.Start(r.Context(), "I18nMiddleware")
			defer span.End()

			req := r.WithContext(spanCtx)
			language := getPreferredUserLanguageFn(req)
			localizer := i18n.NewLocalizer(bundle, language.String())
			next.ServeHTTP(w, translation.WithLocalizer(req, localizer))
		})
	}
}
