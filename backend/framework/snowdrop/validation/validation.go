package validation

import (
	"errors"
	"fmt"

	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/vi"
	ut "github.com/go-playground/universal-translator"
	validation "github.com/go-playground/validator/v10"
	enTranslation "github.com/go-playground/validator/v10/translations/en"
	viTranslation "github.com/go-playground/validator/v10/translations/vi"
	"go.uber.org/fx"
	"golang.org/x/text/language"
)

type Result struct {
	fx.Out
	UniversalTranslator *ut.UniversalTranslator
	Validator           *validation.Validate
}

func NewValidator() (Result, error) {
	universalTranslator := createUniversalTranslator()
	validator := validation.New(validation.WithRequiredStructEnabled())

	err := errors.Join(
		registerTranslation(
			validator,
			universalTranslator,
			language.English.String(),
			enTranslation.RegisterDefaultTranslations,
		),
		registerTranslation(
			validator,
			universalTranslator,
			language.Vietnamese.String(),
			viTranslation.RegisterDefaultTranslations,
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
	enTranslator := en.New()
	viTranslator := vi.New()

	return ut.New(enTranslator, enTranslator, viTranslator)
}

func registerTranslation(
	v *validation.Validate,
	uni *ut.UniversalTranslator,
	lang string,
	registerFn func(*validation.Validate, ut.Translator) error,
) error {
	trans, found := uni.GetTranslator(lang)

	if !found {
		return fmt.Errorf("%s: %w", lang, ErrTranslatorNotFound)
	}

	return registerFn(v, trans)
}
