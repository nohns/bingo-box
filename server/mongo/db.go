package mongo

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
)

type DB struct {
	client      *mongo.Client
	Games       *mongo.Collection
	Cards       *mongo.Collection
	Users       *mongo.Collection
	Players     *mongo.Collection
	Invitations *mongo.Collection
}

func (db *DB) Close(ctx context.Context) error {
	return db.client.Disconnect(ctx)
}

func (db *DB) Client() *mongo.Client {
	return db.client
}

var (
	ErrInvalidConnUri        = errors.New("mongo: connection uri was invalid")
	ErrNoDatabase            = errors.New("mongo: no database name in connection uri")
	ErrNoUpsertedObjectID    = errors.New("mongo: returned upserted id was not of type primitive.ObjectID")
	ErrMalformedHexObjectID  = errors.New("mongo: hex formatted id strings is required")
	ErrNoAssociatedDocuments = errors.New("mongo: no associated documents found")
)

func New(ctx context.Context, connURI string) (*DB, error) {
	// Extract db name from conn uri
	u, err := connstring.Parse(connURI)
	if err != nil {
		return nil, ErrInvalidConnUri
	}
	if u.Database == "" {
		return nil, ErrNoDatabase
	}

	// Connect to mongodb
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(connURI))
	if err != nil {
		return nil, err
	}
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}
	db := client.Database(u.Database)

	return &DB{
		client:      client,
		Games:       db.Collection("games"),
		Cards:       db.Collection("cards"),
		Users:       db.Collection("users"),
		Players:     db.Collection("players"),
		Invitations: db.Collection("invitations"),
	}, nil
}
