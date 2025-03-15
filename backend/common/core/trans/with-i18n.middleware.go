package trans

import (
	"net/http"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

// WithI18nMiddleware is a middleware function that adds internationalization support to an HTTP server.
// It uses the provided i18n.Bundle to create a localizer based on the language specified in the request.
func WithI18nMiddleware(bundle *i18n.Bundle) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			languages := []string{}
			lang := r.URL.Query().Get("lang")

			// Give the highest priority for the language specified in the query parameters
			if t, err := language.Parse(lang); err == nil {
				languages = append(languages, t.String())
			}

			// If no language is specified, fall back to English
			if len(languages) == 0 {
				languages = append(languages, language.English.String())
			}

			localizer := i18n.NewLocalizer(bundle, languages...)

			next.ServeHTTP(w, WithLocalizer(r, localizer))
		})
	}
}
