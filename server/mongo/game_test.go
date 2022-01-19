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

var implementsGameRepo bingo.GameRepository = &mongo.GameRepository{}

// Test that mongodb game doc <-> game aggregate root conversion works
func TestDocGame(t *testing.T) {
	g := &bingo.Game{
		ID:             primitive.NewObjectID().Hex(),
		Name:           "game name",
		HostId:         primitive.NewObjectID().Hex(),
		Host:           nil,
		NextCardNumber: 1,
		CalledNumbers: []bingo.Ball{
			{
				Number: 1,
			},
			{
				Number: 5,
			},
		},
		UpdatedAt: time.Now(),
		CreatedAt: time.Now(),
	}

	t.Run("test data out of date", func(t *testing.T) {
		gFieldsCount := reflect.Indirect(reflect.ValueOf(g)).NumField()
		expectedfc := 8
		require.Equal(t, expectedfc, gFieldsCount, "game test data missing one or more fields")
	})

	t.Run("bidirectional conversion", func(t *testing.T) {
		doc, err := mongo.DocFromGame(g)
		require.NoError(t, err, "no error expected from mongo.DocFromGame")

		cg, err := doc.ToAggregate()
		require.NoError(t, err, "no error exptected from doc.ToAggregate()")

		require.EqualValues(t, g, cg, "Expected these values")
	})
}

type valGameFunc func(context.Context, *testing.T, *bingo.Game)

func TestGameRepository_Get(t *testing.T) {
	gameRepo := mongo.NewGameRepository(sharedDB)
	insertDoc := mongo.DocGame{
		ID:             primitive.NewObjectID(),
		Name:           "test name",
		HostID:         primitive.NewObjectID(),
		NextCardNumber: 1,
		CalledNumbers: []int{
			1, 32, 65,
		},
		UpdatedAt: time.Now(),
		CreatedAt: time.Now(),
	}
	MustInsertOneGameDoc(t, context.Background(), insertDoc)

	cases := []struct {
		cn            string
		ctx           context.Context
		id            string
		expectErr     bool
		valFunc       valGameFunc
		expectedErrIs error
	}{
		{
			cn:            "success",
			ctx:           context.Background(),
			id:            insertDoc.ID.Hex(),
			expectErr:     false,
			expectedErrIs: nil,
			valFunc: func(ctx context.Context, t *testing.T, cg *bingo.Game) {
				g, err := insertDoc.ToAggregate()
				require.NoError(t, err, "expected no error from user aggregate conversion")
				MustCompareGames(t, g, cg)
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
	}

	for _, c := range cases {
		t.Run(c.cn, func(t *testing.T) {
			ctx := c.ctx

			g, err := gameRepo.Get(ctx, c.id)
			if c.expectErr {
				require.Error(t, err, "expected error")
				if c.expectedErrIs != nil {
					require.ErrorIs(t, err, c.expectedErrIs, "expected different error")
				}
			} else {
				require.NoError(t, err, "expected no error")
				c.valFunc(ctx, t, g)
			}
		})
	}
}

func TestGameRepository_Save(t *testing.T) {
	gameRepo := mongo.NewGameRepository(sharedDB)
	commonSub := &bingo.Game{
		Name:           "test name",
		HostId:         primitive.NewObjectID().Hex(),
		Host:           nil,
		NextCardNumber: 1,
		CalledNumbers: []bingo.Ball{
			{Number: 12},
			{Number: 34},
			{Number: 46},
		},
		UpdatedAt: time.Now(),
		CreatedAt: time.Now(),
	}

	cases := []struct {
		cn            string
		ctx           context.Context
		g             *bingo.Game
		beforeFunc    valGameFunc
		valFunc       valGameFunc
		expectErr     bool
		expectedErrIs error
	}{
		{
			cn:         "insert success",
			ctx:        context.Background(),
			g:          commonSub,
			beforeFunc: nil,
			valFunc: func(ctx context.Context, t *testing.T, g *bingo.Game) {
				require.NotEmpty(t, g.ID, commonSub.ID, "expected assigned id for non-inserted game to be set")

				//  Make sure doc is inserted correctly
				doc := MustFindOneGameDoc(t, ctx, g.ID)
				cg, err := doc.ToAggregate()
				require.NoError(t, err, "expected no error from game aggregate conversion")
				MustCompareGames(t, g, cg)

			},
			expectErr:     false,
			expectedErrIs: nil,
		},
		{
			cn:  "update success",
			ctx: context.Background(),
			g:   commonSub,
			beforeFunc: func(ctx context.Context, t *testing.T, g *bingo.Game) {
				// Mutate game
				g.Name = "new name"
				g.NextCardNumber = 2
				g.CalledNumbers = []bingo.Ball{
					{Number: 12},
					{Number: 34},
					{Number: 46},
					{Number: 56},
				}
			},
			valFunc: func(ctx context.Context, t *testing.T, g *bingo.Game) {
				//  Make sure doc is inserted correctly
				doc := MustFindOneGameDoc(t, ctx, g.ID)
				cu, err := doc.ToAggregate()
				require.NoError(t, err, "expected no error from game aggregate conversion")
				MustCompareGames(t, g, cu)
			},
			expectErr:     false,
			expectedErrIs: nil,
		},
		{
			cn:  "insert fail id not hex",
			ctx: context.Background(),
			g: &bingo.Game{
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
				c.beforeFunc(ctx, t, c.g)
			}

			err := gameRepo.Save(ctx, c.g)
			if c.expectErr {
				require.Error(t, err, "expected error")
				if c.expectedErrIs != nil {
					require.ErrorIs(t, err, c.expectedErrIs, "expected different error")
				}
			} else {
				require.NoError(t, err, "expected no error")
				c.valFunc(ctx, t, c.g)
			}
		})
	}
}

func MustInsertOneGameDoc(tb testing.TB, ctx context.Context, doc mongo.DocGame) {
	tb.Helper()

	_, err := sharedDB.Games.InsertOne(ctx, doc)
	require.NoError(tb, err, "expected no error from inserting game")
}

func MustFindOneGameDoc(tb testing.TB, ctx context.Context, id string) mongo.DocGame {
	tb.Helper()

	oid, err := primitive.ObjectIDFromHex(id)
	require.NoError(tb, err, "expected valid object id")
	var doc mongo.DocGame
	err = sharedDB.Games.FindOne(ctx, bson.M{"_id": oid}).Decode(&doc)
	require.NoError(tb, err, "expected no error when trying to find given games")
	return doc
}

func MustCompareGames(tb testing.TB, exp, cmp *bingo.Game) {
	tb.Helper()

	// Workaround for differences in precision between go and mongodb time precision
	if cmp.CreatedAt.UnixMilli() == exp.CreatedAt.UnixMilli() {
		cmp.CreatedAt = exp.CreatedAt
	}
	if cmp.UpdatedAt.UnixMilli() == exp.UpdatedAt.UnixMilli() {
		cmp.UpdatedAt = exp.UpdatedAt
	}
	require.EqualValues(tb, exp, cmp, "expected games in comparison to be equal")
}
