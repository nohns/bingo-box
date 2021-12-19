package requiretest

import (
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"
)

func UUIDv4(t testing.TB) string {
	t.Helper()

	uuidv4, err := uuid.NewV4()
	require.NoError(t, err, "err when generating test uuid v4")
	return uuidv4.String()
}
