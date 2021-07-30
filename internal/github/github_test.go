package github

import (
	"testing"

	"github.com/matryer/is"
)

func TestIsRateAboveLimit(t *testing.T) {
	is := is.New(t)

	is.Equal(false, isAboveTargetUsage(rate{
		Remaining: 4000,
		Limit:     5000,
	}, 50))

	is.Equal(false, isAboveTargetUsage(rate{
		Remaining: 2500,
		Limit:     5000,
	}, 50))

	is.Equal(true, isAboveTargetUsage(rate{
		Remaining: 2499,
		Limit:     5000,
	}, 50))

	is.Equal(true, isAboveTargetUsage(rate{
		Remaining: 500,
		Limit:     5000,
	}, 80))
}
