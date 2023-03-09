package cmd_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"robinmin.net/tools/xally/cmd"
)

func TestGetCurrPath(t *testing.T) {
	assertions := require.New(t)

	assertions.NotNil(cmd.GetCurrPath(), "Get Current Path")
}
