package utility_test

import (
	"testing"

	"github.com/robinmin/xally/shared/utility"
	"github.com/stretchr/testify/require"
)

func TestGetCurrPath(t *testing.T) {
	assertions := require.New(t)

	assertions.NotNil(utility.GetCurrPath(), "Get Current Path")
}
