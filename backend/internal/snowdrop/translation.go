package snowdrop

import (
	"net/http"

	"golang.org/x/text/language"
)

type GetPreferredUserLanguageFn func(*http.Request) language.Tag
