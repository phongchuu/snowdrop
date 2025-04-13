package validation

import (
	"errors"
	"fmt"

	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/vi"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	vi_translations "github.com/go-playground/validator/v10/translations/vi"
	"go.uber.org/fx"
	"golang.org/x/text/language"
)

type Result struct {
	fx.Out
	UniversalTranslator *ut.UniversalTranslator
	Validator           *validator.Validate
}

func NewValidator() (Result, error) {
	universalTranslator := createUniversalTranslator()
	validator := validator.New(validator.WithRequiredStructEnabled())

	err := errors.Join(
		registerTranslation(
			validator,
			universalTranslator,
			language.English.String(),
			en_translations.RegisterDefaultTranslations,
		),
		registerTranslation(
			validator,
			universalTranslator,
			language.Vietnamese.String(),
			vi_translations.RegisterDefaultTranslations,
		),
	)
	if err != nil {
		return Result{}, err
	}

	return Result{
		UniversalTranslator: universalTranslator,
		Validator:           validator,
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
		return fmt.Errorf("%s: %w", lang, ErrTranslatorNotFound)
	}

	if err := registerFn(v, trans); err != nil {
		return err
	}

	return nil
}
