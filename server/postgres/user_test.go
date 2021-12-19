package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	bingo "github.com/nohns/bingo-box/server"
	"github.com/pashagolub/pgxmock"
	"github.com/stretchr/testify/require"
)

func TestUserRepository_Get(t *testing.T) {
	cases := []struct {
		name         string
		userId       string
		expectedErr  error
		expectedUser *bingo.User
	}{
		{
			name:        "should return user sucessfully",
			userId:      "41068cc1-d276-4ae3-9e94-d44a70f2b650",
			expectedErr: nil,
			expectedUser: &bingo.User{
				ID:             "41068cc1-d276-4ae3-9e94-d44a70f2b650",
				Name:           "John Doe",
				Email:          "john@doe.com",
				HashedPassword: []byte("$2a$12$J54obJsxm0vENbhUZcCiku5/bkhpK8pYhyMihIMdPS6WVCJ7H/SlG"),
				UpdatedAt:      time.Now().Add(-59 * time.Minute),
				CreatedAt:      time.Now().Add(-10 * time.Hour),
			},
		},
	}

	mock, repo := MustCreateUserRepo(t)
	defer mock.Close()

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {

			eu := c.expectedUser
			uuid, err := uuid.FromString(c.userId)
			require.NoError(t, err, "could not create uuid struct from userId string")

			rows := mock.NewRows([]string{"id", "name", "email", "hashed_password", "updated_at", "created_at"}).
				AddRow(uuid, eu.Name, eu.Email, eu.HashedPassword, eu.UpdatedAt, eu.CreatedAt)

			mock.ExpectBegin()
			mock.ExpectQuery("SELECT (.+) FROM users").WithArgs(uuid).WillReturnRows(rows)

			u, err := repo.Get(context.Background(), c.userId)
			require.Equal(t, c.expectedErr, err, "should return err = %v", c.expectedErr != nil)
			require.Equal(t, eu, u)

		})
	}
}

func MustCreateUserRepo(tb testing.TB) (pgxmock.PgxPoolIface, bingo.UserRepository) {
	tb.Helper()

	mock, db := MustCreateDB(tb)

	return mock, NewUserRepository(db)
}
