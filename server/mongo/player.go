package mongo

import (
	"context"
	"errors"
	"time"

	bingo "github.com/nohns/bingo-box/server"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DocPlayer struct {
	ID           primitive.ObjectID `bson:"_id"`
	Name         string             `bson:"name"`
	Email        string             `bson:"email"`
	InvitationID primitive.ObjectID `bson:"invitation_id"`
	UpdatedAt    time.Time          `bson:"updated_at"`
	CreatedAt    time.Time          `bson:"created_at"`
}

// Convert mongo document player struct to aggregate player.
func (dp DocPlayer) ToAggregate(inv *bingo.Invitation, cards []bingo.Card) (*bingo.Player, error) {
	p := &bingo.Player{
		ID:           dp.ID.Hex(),
		Name:         dp.Name,
		Email:        dp.Email,
		InvitationID: dp.InvitationID.Hex(),
		Invitation:   inv,
		Cards:        cards,
		UpdatedAt:    dp.UpdatedAt,
		CreatedAt:    dp.CreatedAt,
	}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return p, nil
}

// Create struct representing player mongo document from the aggregate player.
func DocFromPlayer(p *bingo.Player) (DocPlayer, error) {
	oid := primitive.NewObjectID()
	if p.ID != "" {
		var err error
		oid, err = primitive.ObjectIDFromHex(p.ID)
		if err != nil {
			return DocPlayer{}, ErrMalformedHexObjectID
		}
	}
	iOid, err := primitive.ObjectIDFromHex(p.InvitationID)
	if err != nil {
		return DocPlayer{}, ErrMalformedHexObjectID
	}
	return DocPlayer{
		ID:           oid,
		Name:         p.Name,
		Email:        p.Email,
		InvitationID: iOid,
		UpdatedAt:    p.UpdatedAt,
		CreatedAt:    p.CreatedAt,
	}, nil
}

type PlayerRepository struct {
	db *DB
}

func (pr *PlayerRepository) Get(ctx context.Context, id string) (*bingo.Player, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrMalformedHexObjectID
	}
	var doc DocPlayer
	res := pr.db.Players.FindOne(ctx, bson.M{"_id": oid})
	if err := res.Decode(&doc); err != nil {
		return nil, err
	}

	// Find associated invitation for player
	invDoc := &DocInvitation{}
	err = pr.db.Invitations.FindOne(ctx, bson.M{"_id": doc.InvitationID}).Decode(invDoc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNoAssociatedDocuments
	} else if err != nil {
		return nil, err
	}
	var inv *bingo.Invitation
	if invDoc != nil {
		inv, err = invDoc.ToAggregate(nil)
	}
	if err != nil {
		return nil, err
	}

	// Find associated cards for player
	cur, err := pr.db.Cards.Find(ctx, bson.M{"player_id": doc.ID})
	if err != nil {
		return nil, err
	}
	cards := make([]bingo.Card, 0, 3)
	for cur.Next(ctx) {
		var cDoc docCard
		if err = cur.Decode(&cDoc); err != nil {
			return nil, err
		}
		c, err := cDoc.ToAggregate(nil)
		if err != nil {
			return nil, err
		}
		cards = append(cards, *c)
	}

	aggr, err := doc.ToAggregate(inv, cards)
	if err != nil {
		return nil, err
	}

	return aggr, nil
}

func (pr *PlayerRepository) Save(ctx context.Context, p *bingo.Player) error {
	p.UpdatedAt = time.Now()
	doc, err := DocFromPlayer(p)
	if err != nil {
		return err
	}
	opts := options.Replace().SetUpsert(true)
	res, err := pr.db.Players.ReplaceOne(ctx, bson.M{"_id": doc.ID}, doc, opts)
	if err != nil {
		return err
	}
	if p.ID == "" {
		if res.UpsertedID == nil {
			return ErrNoUpsertedObjectID
		}
		oid, ok := res.UpsertedID.(primitive.ObjectID)
		if !ok {
			return ErrNoUpsertedObjectID
		}
		p.ID = oid.Hex()
	}

	return nil
}

func NewPlayerRepository(db *DB) *PlayerRepository {
	return &PlayerRepository{
		db: db,
	}
}
