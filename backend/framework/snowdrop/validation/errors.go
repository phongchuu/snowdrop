package validation

import "errors"

var (
	ErrTranslatorNotFound    = errors.New("translator not found")
	ErrNoUniversalTranslator = errors.New("there is no ut.Translator in the given context")
)
