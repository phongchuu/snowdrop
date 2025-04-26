package web

import (
	"encoding/json"
	"net/http"

	validation "github.com/go-playground/validator/v10"
)

type Binder struct {
	validator *validation.Validate
}

type BinderOptFn func(*Binder)

func NewBinder(optFns ...BinderOptFn) *Binder {
	binder := &Binder{}

	for i := range optFns {
		optFns[i](binder)
	}

	return binder
}

func WithValidator(validator *validation.Validate) BinderOptFn {
	return func(b *Binder) {
		b.validator = validator
	}
}

func (b *Binder) JSON(r *http.Request, out any) error {
	if err := json.NewDecoder(r.Body).Decode(out); err != nil {
		return err
	}

	if b.validator != nil {
		return b.validator.StructCtx(r.Context(), out)
	}

	return nil
}
