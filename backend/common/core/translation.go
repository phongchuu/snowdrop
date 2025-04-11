package core

import (
	"net/http"
	"slices"

	"golang.org/x/text/language"
	snowdrop "internal.snowdrop/framework"
)

func NewGetPreferredUserLanguageFn(
	config snowdrop.ConfigManager,
) snowdrop.GetPreferredUserLanguageFn {
	languageMatcher := language.NewMatcher(config.GetSupportedLanguages())

	return func(r *http.Request) language.Tag {
		var tags []language.Tag

		// Check for user-specified language in query parameter first
		if userLang := r.URL.Query().Get("lang"); userLang != "" {
			if tag, err := language.Parse(userLang); err == nil {
				if slices.Contains(config.GetSupportedLanguages(), tag) {
					tags = append(tags, tag)
				}
			}
		}

		// If no user-specified language or parsing failed, use Accept-Language header
		if len(tags) == 0 {
			acceptLang := r.Header.Get("Accept-Language")
			if parsedTags, _, err := language.ParseAcceptLanguage(acceptLang); err == nil {
				tags = parsedTags
			}
		}

		language, _, _ := languageMatcher.Match(tags...)

		return language
	}
}
