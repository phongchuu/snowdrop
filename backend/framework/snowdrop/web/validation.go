package web

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/vi"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	vi_translations "github.com/go-playground/validator/v10/translations/vi"
	"go.uber.org/fx"
	"golang.org/x/text/language"
	snowdrop "internal.snowdrop/framework"
)

type (
	UniversalTranslatorCtxKey     string
	UniversalTranslatorMiddleware func(next http.Handler) http.Handler
)

type ValidatorResult struct {
	fx.Out
	Validator *validator.Validate
	Uni       *ut.UniversalTranslator
}

const universalTranslatorCtxKey UniversalTranslatorCtxKey = "*ut.Translator"

var errTranslatorNotFound = errors.New("translator not found")

func NewValidator() (ValidatorResult, error) {
	uni := createUniversalTranslator()
	validator := validator.New(validator.WithRequiredStructEnabled())

	err := errors.Join(
		registerTranslation(
			validator,
			uni,
			language.English.String(),
			en_translations.RegisterDefaultTranslations,
		),
		registerTranslation(
			validator,
			uni,
			language.Vietnamese.String(),
			vi_translations.RegisterDefaultTranslations,
		),
	)
	if err != nil {
		return ValidatorResult{}, err
	}

	return ValidatorResult{
		Uni:       uni,
		Validator: validator,
	}, nil
}

func createUniversalTranslator() *ut.UniversalTranslator {
	en := en.New()
	vi := vi.New()

	return ut.New(en, en, vi)
}

func registerTranslation(
	v *validator.Validate,
	uni *ut.UniversalTranslator,
	lang string,
	registerFn func(*validator.Validate, ut.Translator) error,
) error {
	trans, found := uni.GetTranslator(lang)

	if !found {
		return fmt.Errorf("%s: %w", lang, errTranslatorNotFound)
	}

	if err := registerFn(v, trans); err != nil {
		return err
	}

	return nil
}

// GetValidationTranslator retrieves the universal translator from request context.
//
//nolint:ireturn
func GetValidationTranslator(r *http.Request) ut.Translator {
	translator, _ := r.Context().Value(universalTranslatorCtxKey).(ut.Translator)

	return translator
}

func WithUniversalTranslator(r *http.Request, translator ut.Translator) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), universalTranslatorCtxKey, translator))
}

func NewUniversalTranslatorMiddleware(
	universalTranslator *ut.UniversalTranslator,
	getPreferredUserLanguageFn snowdrop.GetPreferredUserLanguageFn,
) UniversalTranslatorMiddleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			language := getPreferredUserLanguageFn(r)
			translator, _ := universalTranslator.GetTranslator(language.String())
			next.ServeHTTP(w, WithUniversalTranslator(r, translator))
		})
	}
}
