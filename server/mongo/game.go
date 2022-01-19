package mongo

import (
	"context"
	"time"

	bingo "github.com/nohns/bingo-box/server"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DocGame struct {
	ID             primitive.ObjectID `bson:"_id"`
	Name           string             `bson:"name"`
	HostID         primitive.ObjectID `bson:"host_id"`
	NextCardNumber int                `bson:"next_card_number"`
	CalledNumbers  []int              `bson:"called_numbers"`
	UpdatedAt      time.Time          `bson:"updated_at"`
	CreatedAt      time.Time          `bson:"created_at"`
}

func (dg DocGame) ToAggregate() (*bingo.Game, error) {
	calledNums := make([]bingo.Ball, 0, len(dg.CalledNumbers))
	for _, n := range dg.CalledNumbers {
		calledNums = append(calledNums, bingo.Ball{Number: n})
	}
	g := &bingo.Game{
		ID:             dg.ID.Hex(),
		Name:           dg.Name,
		HostId:         dg.HostID.Hex(),
		NextCardNumber: dg.NextCardNumber,
		CalledNumbers:  calledNums,
		UpdatedAt:      dg.UpdatedAt,
		CreatedAt:      dg.CreatedAt,
	}
	if err := g.Validate(); err != nil {
		return nil, err
	}
	return g, nil
}

func DocFromGame(g *bingo.Game) (DocGame, error) {
	oid := primitive.NewObjectID()
	if g.ID != "" {
		var err error
		oid, err = primitive.ObjectIDFromHex(g.ID)
		if err != nil {
			return DocGame{}, ErrMalformedHexObjectID
		}
	}
	hOid, err := primitive.ObjectIDFromHex(g.HostId)
	if err != nil {
		return DocGame{}, ErrMalformedHexObjectID
	}
	nums := make([]int, 0, len(g.CalledNumbers))
	for _, cn := range g.CalledNumbers {
		nums = append(nums, cn.Number)
	}
	return DocGame{
		ID:             oid,
		Name:           g.Name,
		HostID:         hOid,
		NextCardNumber: g.NextCardNumber,
		CalledNumbers:  nums,
		UpdatedAt:      g.UpdatedAt,
		CreatedAt:      g.CreatedAt,
	}, nil
}

type GameRepository struct {
	db *DB
}

func (gr *GameRepository) Get(ctx context.Context, id string) (*bingo.Game, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrMalformedHexObjectID
	}
	var doc DocGame
	res := gr.db.Games.FindOne(ctx, bson.M{"_id": oid})
	if err := res.Decode(&doc); err != nil {
		return nil, err
	}

	aggr, err := doc.ToAggregate()
	if err != nil {
		return nil, err
	}

	return aggr, nil
}

func (gr *GameRepository) Save(ctx context.Context, g *bingo.Game) error {
	g.UpdatedAt = time.Now()
	doc, err := DocFromGame(g)
	if err != nil {
		return err
	}
	opts := options.Replace().SetUpsert(true)
	res, err := gr.db.Games.ReplaceOne(ctx, bson.M{"_id": doc.ID}, doc, opts)
	if err != nil {
		return err
	}
	if g.ID == "" {
		if res.UpsertedID == nil {
			return ErrNoUpsertedObjectID
		}
		oid, ok := res.UpsertedID.(primitive.ObjectID)
		if !ok {
			return ErrNoUpsertedObjectID
		}
		g.ID = oid.Hex()
	}

	return nil
}

func NewGameRepository(db *DB) *GameRepository {
	return &GameRepository{
		db: db,
	}
}
