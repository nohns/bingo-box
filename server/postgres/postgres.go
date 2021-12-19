package postgres

import (
	"context"

	migratedb "github.com/golang-migrate/migrate/v4/database"
	migratepgx "github.com/golang-migrate/migrate/v4/database/pgx"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	pgtypeuuid "github.com/jackc/pgtype/ext/gofrs-uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/jackc/pgx/v4/stdlib"
)

type DB struct {
	conn     dbconn
	mgDriver migratedb.Driver
}

type dbconn interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, optionsAndArgs ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, optionsAndArgs ...interface{}) pgx.Row
}

func NewDB(ctx context.Context, dsn, migrationsPath string) (*DB, error) {

	// Register uuid type
	dbconfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	dbconfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		conn.ConnInfo().RegisterDataType(pgtype.DataType{
			Value: &pgtypeuuid.UUID{},
			Name:  "uuid",
			OID:   pgtype.UUIDOID,
		})
		return nil
	}

	// Connect to PostgreSQL database
	pool, err := pgxpool.ConnectConfig(ctx, dbconfig)
	if err != nil {
		return nil, err
	}

	// Create migrate conn
	stdDbconfig, _ := pgx.ParseConfig(dsn)
	mgConn := stdlib.OpenDB(*stdDbconfig)

	// Create migration driver
	mgDriver, err := migratepgx.WithInstance(mgConn, new(migratepgx.Config))
	if err != nil {
		return nil, err
	}

	return &DB{
		conn:     pool,
		mgDriver: mgDriver,
	}, nil
}
