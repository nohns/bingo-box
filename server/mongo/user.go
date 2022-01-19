package mongo

import (
	"context"
	"time"

	bingo "github.com/nohns/bingo-box/server"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DocUser struct {
	ID             primitive.ObjectID `bson:"_id"`
	Name           string             `bson:"name"`
	Email          string             `bson:"email"`
	HashedPassword []byte             `bson:"password"`
	UpdatedAt      time.Time          `bson:"updated_at"`
	CreatedAt      time.Time          `bson:"created_at"`
}

func (du DocUser) ToAggregate() (*bingo.User, error) {
	u := &bingo.User{
		ID:             du.ID.Hex(),
		Name:           du.Name,
		Email:          du.Email,
		HashedPassword: du.HashedPassword,
		UpdatedAt:      du.UpdatedAt,
		CreatedAt:      du.CreatedAt,
	}
	if err := u.Validate(); err != nil {
		return nil, err
	}
	return u, nil
}

func DocFromUser(u *bingo.User) (DocUser, error) {
	var err error
	oid := primitive.NewObjectID()
	if u.ID != "" {
		oid, err = primitive.ObjectIDFromHex(u.ID)
	}
	if err != nil {
		return DocUser{}, ErrMalformedHexObjectID
	}
	doc := DocUser{
		ID:             oid,
		Name:           u.Name,
		Email:          u.Email,
		HashedPassword: u.HashedPassword,
		UpdatedAt:      u.UpdatedAt,
		CreatedAt:      u.CreatedAt,
	}
	return doc, nil
}

type UserRepository struct {
	db *DB
}

func NewUserRepository(db *DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (cr *UserRepository) Get(ctx context.Context, id string) (*bingo.User, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrMalformedHexObjectID
	}
	var doc DocUser
	res := cr.db.Users.FindOne(ctx, bson.M{"_id": oid})
	if err := res.Decode(&doc); err != nil {
		return nil, err
	}

	aggr, err := doc.ToAggregate()
	if err != nil {
		return nil, err
	}

	return aggr, nil
}

func (cr *UserRepository) GetByEmail(ctx context.Context, email string) (*bingo.User, error) {
	var doc DocUser
	res := cr.db.Users.FindOne(ctx, bson.M{"email": email})
	if err := res.Decode(&doc); err != nil {
		return nil, err
	}

	aggr, err := doc.ToAggregate()
	if err != nil {
		return nil, err
	}

	return aggr, nil
}

// Save card to mongodb
func (cr *UserRepository) Save(ctx context.Context, c *bingo.User) error {
	c.UpdatedAt = time.Now()
	doc, err := DocFromUser(c)
	if err != nil {
		return err
	}
	opts := options.Replace().SetUpsert(true)
	res, err := cr.db.Users.ReplaceOne(ctx, bson.M{"_id": doc.ID}, doc, opts)
	if err != nil {
		return err
	}
	if c.ID == "" {
		if res.UpsertedID == nil {
			return ErrNoUpsertedObjectID
		}
		oid, ok := res.UpsertedID.(primitive.ObjectID)
		if !ok {
			return ErrNoUpsertedObjectID
		}
		c.ID = oid.Hex()
	}

	return nil
}
