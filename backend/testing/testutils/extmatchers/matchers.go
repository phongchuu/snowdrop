package extmatchers

import (
	"github.com/google/uuid"
	"github.com/onsi/gomega/gcustom"
	"github.com/onsi/gomega/types"
	"github.com/spf13/cast"
)

func BeValidUUID() types.GomegaMatcher {
	return gcustom.MakeMatcher(func(actual any) (bool, error) {
		value, err := cast.ToStringE(actual)
		if err != nil {
			return false, err
		}

		if err = uuid.Validate(value); err != nil {
			return false, err
		}

		return true, nil
	}).WithTemplate("to be a valid UUID")
}
