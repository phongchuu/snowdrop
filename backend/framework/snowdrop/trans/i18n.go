package trans

import (
	"context"
	"errors"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
	snowdrop "internal.snowdrop/framework"
)

type LocalizerCtxKey string

const localizerCtxID LocalizerCtxKey = "LocalizerCtxID"

var ErrNoLocalizer = errors.New("there is no *i18n.Localizer in the given context")

func NewI18nBundle(config snowdrop.ConfigManager) (*i18n.Bundle, error) {
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)

	embedResourcesFolder := config.GetEmbedResourceFolder()

	files, err := glob(embedResourcesFolder, "resources/trans/*.yaml")
	if err != nil {
		return nil, err
	}

	for i := range files {
		if _, err := bundle.LoadMessageFileFS(embedResourcesFolder, files[i]); err != nil {
			return nil, err
		}
	}

	return bundle, nil
}

func WithLocalizer(r *http.Request, localizer *i18n.Localizer) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), localizerCtxID, localizer))
}

func GetLocalizer(r *http.Request) (*i18n.Localizer, error) {
	if localizer, ok := r.Context().Value(localizerCtxID).(*i18n.Localizer); ok {
		return localizer, nil
	}

	return nil, ErrNoLocalizer
}

// glob is a function that performs a glob-style pattern matching on a given file system (fs.FS)
// and returns a list of matching file paths. It uses fs.WalkDir to traverse the file system
// and filepath.Match to check if each file path matches the provided pattern.
func glob(f fs.FS, pattern string) ([]string, error) {
	var matches []string

	err := fs.WalkDir(f, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() && !strings.HasSuffix(pattern, "/**") {
			return nil
		}

		matched, err := filepath.Match(pattern, path)
		if err != nil {
			return err
		}

		if matched {
			matches = append(matches, path)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return matches, nil
}
