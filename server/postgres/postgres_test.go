package postgres

import (
	"testing"

	"github.com/pashagolub/pgxmock"
	"github.com/stretchr/testify/require"
)

func MustCreateDB(tb testing.TB) (pgxmock.PgxPoolIface, *DB) {
	tb.Helper()

	mock, err := pgxmock.NewPool()
	require.NoError(tb, err, "mock conn could not be made")

	return mock, &DB{
		conn: mock,
	}
}
