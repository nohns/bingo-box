package mongo

import (
	"context"
	"errors"

	bingo "github.com/nohns/bingo-box/server"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DocInvitation struct {
	ID             primitive.ObjectID `bson:"_id"`
	DeliveryMethod string             `bson:"delivery_method"`
	MaxCardAmount  int                `bson:"max_card_amount"`
	Active         bool               `bson:"active"`

	GameID primitive.ObjectID `bson:"game_id"`

	Criteria []DocInvitationCriterion `json:"criteria,inline"`
}

type DocInvitationCriterion struct {
	Kind  string `bson:"kind"`
	Field string `bson:"field"`
	Value string `bson:"value"`
}

func (di DocInvitation) ToAggregate(g *bingo.Game) (*bingo.Invitation, error) {
	criteria := make([]bingo.InvitationCriterion, 0, len(di.Criteria))
	for _, c := range di.Criteria {
		criteria = append(criteria, bingo.InvitationCriterion{
			Kind:  bingo.InvitationCriterionKind(c.Kind),
			Field: bingo.InvitationCriterionField(c.Field),
			Value: c.Value,
		})
	}
	inv := &bingo.Invitation{
		ID:             di.ID.Hex(),
		DeliveryMethod: bingo.InvitationDeliveryMethod(di.DeliveryMethod),
		MaxCardAmount:  di.MaxCardAmount,
		Active:         di.Active,
		GameID:         di.GameID.Hex(),
		Game:           g,
		Criteria:       criteria,
	}
	if err := inv.Validate(); err != nil {
		return nil, err
	}
	return inv, nil
}

func DocFromInvitation(inv *bingo.Invitation) (DocInvitation, error) {
	oid := primitive.NewObjectID()
	if inv.ID != "" {
		var err error
		oid, err = primitive.ObjectIDFromHex(inv.ID)
		if err != nil {
			return DocInvitation{}, ErrMalformedHexObjectID
		}
	}
	gOid, err := primitive.ObjectIDFromHex(inv.GameID)
	if err != nil {
		return DocInvitation{}, ErrMalformedHexObjectID
	}
	criteria := make([]DocInvitationCriterion, 0, len(inv.Criteria))
	for _, c := range inv.Criteria {
		criteria = append(criteria, DocInvitationCriterion{
			Kind:  string(c.Kind),
			Field: string(c.Field),
			Value: c.Value,
		})
	}
	doc := DocInvitation{
		ID:             oid,
		DeliveryMethod: string(inv.DeliveryMethod),
		MaxCardAmount:  inv.MaxCardAmount,
		Active:         inv.Active,
		GameID:         gOid,
		Criteria:       criteria,
	}
	return doc, nil
}

type InvitationRepository struct {
	db *DB
}

func (ir *InvitationRepository) Get(ctx context.Context, id string) (*bingo.Invitation, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrMalformedHexObjectID
	}
	var doc DocInvitation
	res := ir.db.Invitations.FindOne(ctx, bson.M{"_id": oid})
	if err := res.Decode(&doc); err != nil {
		return nil, err
	}

	// Find associated game to invitation
	gDoc := &DocGame{}
	err = ir.db.Games.FindOne(ctx, bson.M{"_id": doc.GameID}).Decode(gDoc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		gDoc, err = nil, nil
	} else if err != nil {
		return nil, err
	}
	var g *bingo.Game
	if gDoc != nil {
		g, err = gDoc.ToAggregate()
	}
	if err != nil {
		return nil, err
	}

	aggr, err := doc.ToAggregate(g)
	if err != nil {
		return nil, err
	}

	return aggr, nil
}

func (ir *InvitationRepository) Save(ctx context.Context, inv *bingo.Invitation) error {
	doc, err := DocFromInvitation(inv)
	if err != nil {
		return err
	}
	opts := options.Replace().SetUpsert(true)
	res, err := ir.db.Invitations.ReplaceOne(ctx, bson.M{"_id": doc.ID}, doc, opts)
	if err != nil {
		return err
	}
	if inv.ID == "" {
		if res.UpsertedID == nil {
			return ErrNoUpsertedObjectID
		}
		oid, ok := res.UpsertedID.(primitive.ObjectID)
		if !ok {
			return ErrNoUpsertedObjectID
		}
		inv.ID = oid.Hex()
	}

	return nil
}

func NewInvitationRepository(db *DB) *InvitationRepository {
	return &InvitationRepository{
		db: db,
	}
}
