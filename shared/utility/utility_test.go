package utility_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"robinmin.net/tools/xally/shared/utility"
)

func TestGetCurrPath(t *testing.T) {
	assertions := require.New(t)

	assertions.NotNil(utility.GetCurrPath(), "Get Current Path")
}
