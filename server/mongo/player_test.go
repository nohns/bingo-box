package mongo_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	bingo "github.com/nohns/bingo-box/server"
	"github.com/nohns/bingo-box/server/mongo"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mongodb "go.mongodb.org/mongo-driver/mongo"
)

var implementsPlayerRepo bingo.PlayerRepository = &mongo.PlayerRepository{}

// Test that mongodb card player <-> player aggregate root conversion works
func TestDocPlayer(t *testing.T) {
	p := &bingo.Player{
		ID:           primitive.NewObjectID().Hex(),
		Name:         "test name",
		Email:        "test@test.com",
		InvitationID: primitive.NewObjectID().Hex(),
		Invitation:   nil,
		Cards:        nil,
		UpdatedAt:    time.Now(),
		CreatedAt:    time.Now(),
	}

	t.Run("test data out of date", func(t *testing.T) {
		pFieldsCount := reflect.Indirect(reflect.ValueOf(p)).NumField()
		expectedfc := 8
		require.Equal(t, expectedfc, pFieldsCount, "player test data missing one or more fields")
	})

	t.Run("bidirectional conversion", func(t *testing.T) {

		doc, err := mongo.DocFromPlayer(p)
		require.NoError(t, err, "no error expected from mongo.DocFromPlayer")

		cp, err := doc.ToAggregate(p.Invitation, p.Cards)
		require.NoError(t, err, "no error exptected from doc.ToAggregate()")
		require.EqualValues(t, p, cp, "Expected values of round-trip conversion to equal initial data")
	})
}

type valPlayerFunc func(context.Context, *testing.T, *bingo.Player)

func TestPlayerRepository_Get(t *testing.T) {
	playerRepo := mongo.NewPlayerRepository(sharedDB)
	insertInvDepDoc := mongo.DocInvitation{
		ID:             primitive.NewObjectID(),
		DeliveryMethod: string(bingo.InvitationDeliveryMethodDownload),
		MaxCardAmount:  3,
		Active:         true,
		GameID:         primitive.NewObjectID(),
		Criteria: []mongo.DocInvitationCriterion{
			{
				Kind:  string(bingo.InvitationCriterionKindRegex),
				Field: string(bingo.InvitationCriterionFieldEmail),
				Value: "regex",
			},
		},
	}
	insertDoc := mongo.DocPlayer{
		ID:           primitive.NewObjectID(),
		Name:         "test name",
		Email:        "test@test.com",
		InvitationID: insertInvDepDoc.ID,
		UpdatedAt:    time.Now(),
		CreatedAt:    time.Now(),
	}
	danglingDoc := insertDoc
	danglingDoc.ID = primitive.NewObjectID()
	danglingDoc.InvitationID = primitive.NewObjectID()
	MustInsertOneInvDoc(t, context.Background(), insertInvDepDoc)
	MustInsertOnePlayerDoc(t, context.Background(), insertDoc)
	MustInsertOnePlayerDoc(t, context.Background(), danglingDoc)

	cases := []struct {
		cn            string
		ctx           context.Context
		id            string
		expectErr     bool
		valFunc       valPlayerFunc
		expectedErrIs error
	}{
		{
			cn:            "success",
			ctx:           context.Background(),
			id:            insertDoc.ID.Hex(),
			expectErr:     false,
			expectedErrIs: nil,
			valFunc: func(ctx context.Context, t *testing.T, cp *bingo.Player) {
				p, err := insertDoc.ToAggregate(nil, []bingo.Card{})
				require.NoError(t, err, "expected no error from player aggregate conversion")
				MustComparePlayers(t, p, cp)
			},
		},
		{
			cn:            "fail no docs",
			ctx:           context.Background(),
			id:            primitive.NewObjectID().Hex(),
			expectErr:     true,
			expectedErrIs: mongodb.ErrNoDocuments,
			valFunc:       nil,
		},
		{
			cn:            "fail malformed hex",
			ctx:           context.Background(),
			id:            "non-hex id",
			expectErr:     true,
			expectedErrIs: mongo.ErrMalformedHexObjectID,
			valFunc:       nil,
		},
		{
			cn:            "fail no associated docs",
			ctx:           context.Background(),
			id:            danglingDoc.ID.Hex(),
			expectErr:     true,
			expectedErrIs: mongo.ErrNoAssociatedDocuments,
			valFunc:       nil,
		},
	}

	for _, c := range cases {
		t.Run(c.cn, func(t *testing.T) {
			ctx := c.ctx

			p, err := playerRepo.Get(ctx, c.id)
			if c.expectErr {
				require.Error(t, err, "expected error")
				if c.expectedErrIs != nil {
					require.ErrorIs(t, err, c.expectedErrIs, "expected different error")
				}
			} else {
				require.NoError(t, err, "expected no error")
				c.valFunc(ctx, t, p)
			}
		})
	}
}

func TestPlayerRepository_Save(t *testing.T) {
	playerRepo := mongo.NewPlayerRepository(sharedDB)
	commonSub := &bingo.Player{
		Name:         "test name",
		Email:        "test@test.com",
		InvitationID: primitive.NewObjectID().Hex(),
		Invitation:   nil,
		Cards:        []bingo.Card{},
		UpdatedAt:    time.Now(),
		CreatedAt:    time.Now(),
	}

	cases := []struct {
		cn            string
		ctx           context.Context
		p             *bingo.Player
		beforeFunc    valPlayerFunc
		valFunc       valPlayerFunc
		expectErr     bool
		expectedErrIs error
	}{
		{
			cn:         "insert success",
			ctx:        context.Background(),
			p:          commonSub,
			beforeFunc: nil,
			valFunc: func(ctx context.Context, t *testing.T, p *bingo.Player) {
				require.NotEmpty(t, p.ID, commonSub.ID, "expected assigned id for non-inserted player to be set")

				//  Make sure doc is inserted correctly
				doc := MustFindOnePlayerDoc(t, ctx, p.ID)
				cp, err := doc.ToAggregate(nil, []bingo.Card{})
				require.NoError(t, err, "expected no error from player aggregate conversion")
				MustComparePlayers(t, p, cp)

			},
			expectErr:     false,
			expectedErrIs: nil,
		},
		{
			cn:  "update success",
			ctx: context.Background(),
			p:   commonSub,
			beforeFunc: func(ctx context.Context, t *testing.T, p *bingo.Player) {
				// Mutate player
				p.Email = "test2@test.com"
				p.Name = "new name"
			},
			valFunc: func(ctx context.Context, t *testing.T, p *bingo.Player) {
				//  Make sure doc is inserted correctly
				doc := MustFindOnePlayerDoc(t, ctx, p.ID)
				cp, err := doc.ToAggregate(nil, []bingo.Card{})
				require.NoError(t, err, "expected no error from player aggregate conversion")
				MustComparePlayers(t, p, cp)
			},
			expectErr:     false,
			expectedErrIs: nil,
		},
		{
			cn:  "insert fail id not hex",
			ctx: context.Background(),
			p: &bingo.Player{
				ID: "non-hex value",
			},
			beforeFunc:    nil,
			valFunc:       nil,
			expectErr:     true,
			expectedErrIs: mongo.ErrMalformedHexObjectID,
		},
	}

	for _, c := range cases {
		t.Run(c.cn, func(t *testing.T) {
			ctx := c.ctx
			if c.beforeFunc != nil {
				c.beforeFunc(ctx, t, c.p)
			}

			err := playerRepo.Save(ctx, c.p)
			if c.expectErr {
				require.Error(t, err, "expected error")
				if c.expectedErrIs != nil {
					require.ErrorIs(t, err, c.expectedErrIs, "expected different error")
				}
			} else {
				require.NoError(t, err, "expected no error")
				c.valFunc(ctx, t, c.p)
			}
		})
	}
}

func MustInsertOnePlayerDoc(tb testing.TB, ctx context.Context, doc mongo.DocPlayer) {
	tb.Helper()

	_, err := sharedDB.Players.InsertOne(ctx, doc)
	require.NoError(tb, err, "expected no error from inserting player")
}

func MustFindOnePlayerDoc(tb testing.TB, ctx context.Context, id string) mongo.DocPlayer {
	tb.Helper()

	oid, err := primitive.ObjectIDFromHex(id)
	require.NoError(tb, err, "expected valid object id")
	var doc mongo.DocPlayer
	err = sharedDB.Players.FindOne(ctx, bson.M{"_id": oid}).Decode(&doc)
	require.NoError(tb, err, "expected no error when trying to find given players")
	return doc
}

func MustComparePlayers(tb testing.TB, exp, cmp *bingo.Player) {
	tb.Helper()

	// Workaround for differences in precision between go and mongodb time precision
	if cmp.CreatedAt.UnixMilli() == exp.CreatedAt.UnixMilli() {
		cmp.CreatedAt = exp.CreatedAt
	}
	if cmp.UpdatedAt.UnixMilli() == exp.UpdatedAt.UnixMilli() {
		cmp.UpdatedAt = exp.UpdatedAt
	}
	// Comparing invitation is not relevant when comparing players
	exp.Invitation = nil
	cmp.Invitation = nil
	require.EqualValues(tb, exp, cmp, "expected players in comparison to be equal")
}
