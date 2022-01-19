package mongo

import (
	"context"
	"errors"

	bingo "github.com/nohns/bingo-box/server"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

type docCard struct {
	ID     primitive.ObjectID `bson:"_id"`
	Number int                `bson:"number"`

	// Game in which the bingo card is used
	GameID primitive.ObjectID `bson:"game_id"`

	// Player who owns the card, if it was generated via invitation
	PlayerID primitive.ObjectID `bson:"player_id"`

	GridNumbers []docCardGridNumber `bson:"grid_numbers"`
}

type docCardGridNumber struct {
	Row    int `bson:"row"`
	Col    int `bson:"col"`
	Number int `bson:"number"`
}

func (dc docCard) ToAggregate(p *bingo.Player) (*bingo.Card, error) {

	gridNums := make([]bingo.CardGridNumber, 0, len(dc.GridNumbers))
	for _, gn := range dc.GridNumbers {
		gridNums = append(gridNums, bingo.CardGridNumber{
			Row:    gn.Row,
			Col:    gn.Col,
			Number: gn.Number,
		})
	}

	c := &bingo.Card{
		ID:          dc.ID.Hex(),
		Number:      dc.Number,
		GameID:      dc.GameID.Hex(),
		PlayerID:    dc.PlayerID.Hex(),
		Player:      p,
		GridNumbers: gridNums,
	}
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return c, nil
}

func DocFromCard(c *bingo.Card) (docCard, error) {
	oid, err := primitive.ObjectIDFromHex(c.ID)
	if err != nil {
		return docCard{}, err
	}
	gOid, err := primitive.ObjectIDFromHex(c.GameID)
	if err != nil {
		return docCard{}, err
	}
	pOid, err := primitive.ObjectIDFromHex(c.PlayerID)
	if err != nil {
		return docCard{}, err
	}
	gridNums := make([]docCardGridNumber, 0, len(c.GridNumbers))
	for _, gn := range c.GridNumbers {
		gridNums = append(gridNums, docCardGridNumber{
			Row:    gn.Row,
			Col:    gn.Col,
			Number: gn.Number,
		})
	}
	return docCard{
		ID:          oid,
		Number:      c.Number,
		GameID:      gOid,
		PlayerID:    pOid,
		GridNumbers: gridNums,
	}, nil
}

type CardRepository struct {
	db *DB
}

func (cr *CardRepository) GetByNumber(ctx context.Context, cardNum int, gameId string) (*bingo.Card, error) {
	gOid, err := primitive.ObjectIDFromHex(gameId)
	if err != nil {
		return nil, err
	}
	var doc docCard
	res := cr.db.Cards.FindOne(ctx, bson.M{"number": cardNum, "game_id": gOid})
	if err := res.Decode(&doc); err != nil {
		return nil, err
	}

	// Find associated player to card
	var pDoc *DocPlayer
	err = cr.db.Players.FindOne(ctx, bson.M{"_id": doc.PlayerID}).Decode(pDoc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, err
	} else if err != nil {
		return nil, err
	}
	var p *bingo.Player
	if pDoc != nil {
		p, err = pDoc.ToAggregate(nil, make([]bingo.Card, 0))
	}
	if err != nil {
		return nil, err
	}

	aggr, err := doc.ToAggregate(p)
	if err != nil {
		return nil, err
	}

	return aggr, nil
}

// Save all given cards with Save() method but wrapped in a mongodb acid transaction
func (cr *CardRepository) SaveAll(ctx context.Context, cards []bingo.Card) error {
	wc := writeconcern.New(writeconcern.WMajority())
	rc := readconcern.Snapshot()
	txnOpts := options.Transaction().SetWriteConcern(wc).SetReadConcern(rc)
	sess, err := cr.db.client.StartSession()
	if err != nil {
		return err
	}
	err = mongo.WithSession(ctx, sess, func(sc mongo.SessionContext) error {
		if err = sess.StartTransaction(txnOpts); err != nil {
			return err
		}
		for _, c := range cards {
			if err := cr.Save(ctx, &c); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		if tErr := sess.AbortTransaction(ctx); tErr != nil {
			return tErr
		}

		return err
	}

	return nil
}

// Save card to mongodb
func (cr *CardRepository) Save(ctx context.Context, c *bingo.Card) error {
	doc, err := DocFromCard(c)
	if err != nil {
		return err
	}
	oid, err := primitive.ObjectIDFromHex(c.ID)
	if err != nil {
		return err
	}
	opts := options.Update().SetUpsert(true)
	res, err := cr.db.Cards.UpdateByID(ctx, oid, doc, opts)
	oid, ok := res.UpsertedID.(primitive.ObjectID)
	if !ok {
		return errors.New("mongo: returned inserted id was not of type primitive.ObjectID")
	}
	if c.ID == "" {
		c.ID = oid.Hex()
	}

	return nil
}

func NewCardRepository(db *DB) *CardRepository {
	return &CardRepository{
		db: db,
	}
}
