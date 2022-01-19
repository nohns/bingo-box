package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgconn"
	bingo "github.com/nohns/bingo-box/server"

	"github.com/georgysavva/scany/pgxscan"
)

type UserRepository struct {
	psql *DB
}

// Get user by its primary key, id, in the database
func (ur *UserRepository) Get(ctx context.Context, id string) (*bingo.User, error) {
	// Try to form uuid from id string
	uuid, err := uuid.FromString(id)
	if err != nil {
		return nil, err
	}

	// Find users with id
	users, err := ur.find(ctx, "id = $1", uuid)
	if err != nil {
		return nil, err
	}

	// Make sure we actually found any records
	if len(users) == 0 {
		return nil, bingo.ErrUserNotFound
	}

	return users[0], nil
}

// Get user by email string column in the database
func (ur *UserRepository) GetByEmail(ctx context.Context, email string) (*bingo.User, error) {
	users, err := ur.find(ctx, "email = $1", email)
	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, bingo.ErrUserNotFound
	}

	return users[0], nil
}

// Save the given user in the database, by updating or inserting it depending on the existance of a user id.
func (ur *UserRepository) Save(ctx context.Context, user *bingo.User) error {
	// If user id already exists, then update it
	if user.ID != "" {
		return ur.update(ctx, user)
	}

	// If no user id was found on user, then insert it
	return ur.insert(ctx, user)
}

func (ur *UserRepository) find(ctx context.Context, where string, param ...interface{}) ([]*bingo.User, error) {
	var users []*bingo.User

	sql := "SELECT id, name, email, hashed_password, updated_at, created_at FROM users WHERE " + where
	err := pgxscan.Select(ctx, ur.psql.conn, &users, sql, param...)
	if err != nil {
		return nil, ur.translatePSQLError(err)
	}

	return users, nil
}

func (ur *UserRepository) insert(ctx context.Context, u *bingo.User) error {
	// Start transaction and make sure it is valid
	tx, err := ur.psql.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var uuid uuid.UUID

	sql := "INSERT INTO users (name, email, updated_at, created_at) VALUES ($1, $2, $3, $4, $5) RETURNING id"
	err = ur.psql.conn.QueryRow(ctx, sql, u.Name, u.Email, u.UpdatedAt, u.CreatedAt).Scan(&uuid)
	if err != nil {
		return ur.translatePSQLError(err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	// Assign uuid to user
	u.ID = uuid.String()

	return nil
}

func (ur *UserRepository) update(ctx context.Context, u *bingo.User) error {
	// Try to form uuid from id string
	uuid, err := uuid.FromString(u.ID)
	if err != nil {
		return err
	}

	// Start transaction and make sure it is valid
	tx, err := ur.psql.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var newUpdatedAt time.Time

	// Try update user data
	sql := "UPDATE users SET name = $1, email = $2, updated_at = CURRENT_TIMESTAMP() WHERE id = $4 RETURNING updated_at"
	err = ur.psql.conn.QueryRow(ctx, sql, u.Name, u.Email, uuid).Scan(&newUpdatedAt)
	if err != nil {
		return ur.translatePSQLError(err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	// Update the updated at date of users
	u.UpdatedAt = newUpdatedAt

	return nil
}

func (ur *UserRepository) translatePSQLError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return err
	}

	// Handle postgres error codes
	switch pgErr.Code {
	case "23505": // Unique violation
		return bingo.ErrUserAlreadyExists
	default:
		return err
	}
}

func NewUserRepository(db *DB) *UserRepository {
	return &UserRepository{
		psql: db,
	}
}
