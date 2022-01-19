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

var implementsInvitationRepo bingo.InvitationRepository = &mongo.InvitationRepository{}

// Test that mongodb card invitation <-> invitation aggregate root conversion works
func TestDocInvitation(t *testing.T) {
	inv := &bingo.Invitation{
		ID:             primitive.NewObjectID().Hex(),
		DeliveryMethod: bingo.InvitationDeliveryMethodDownload,
		MaxCardAmount:  3,
		Active:         true,
		GameID:         primitive.NewObjectID().Hex(),
		Game:           nil,
		Criteria: []bingo.InvitationCriterion{
			{
				Kind:  bingo.InvitationCriterionKindRegex,
				Field: bingo.InvitationCriterionFieldEmail,
				Value: "email regex",
			},
		},
	}

	t.Run("test data out of date", func(t *testing.T) {
		invFieldsCount := reflect.Indirect(reflect.ValueOf(inv)).NumField()
		expectedfc := 7
		require.Equal(t, expectedfc, invFieldsCount, "invitation test data missing one or more fields")
	})

	t.Run("bidirectional conversion", func(t *testing.T) {
		doc, err := mongo.DocFromInvitation(inv)
		require.NoError(t, err, "no error expected from mongo.DocFromInvitation")

		ci, err := doc.ToAggregate(inv.Game)
		require.NoError(t, err, "no error exptected from doc.ToAggregate()")
		require.EqualValues(t, inv, ci, "Expected values of round-trip conversion to equal initial data")
	})
}

type valInvitationFunc func(context.Context, *testing.T, *bingo.Invitation)

func TestInvitationRepository_Get(t *testing.T) {
	invRepo := mongo.NewInvitationRepository(sharedDB)

	insertGameDepDoc := mongo.DocGame{
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
	insertDoc := mongo.DocInvitation{
		ID:             primitive.NewObjectID(),
		DeliveryMethod: string(bingo.InvitationDeliveryMethodDownload),
		MaxCardAmount:  3,
		Active:         true,
		GameID:         insertGameDepDoc.ID,
		Criteria: []mongo.DocInvitationCriterion{
			{
				Kind:  string(bingo.InvitationCriterionKindRegex),
				Field: string(bingo.InvitationCriterionFieldEmail),
				Value: "regex",
			},
		},
	}
	MustInsertOneGameDoc(t, context.Background(), insertGameDepDoc)
	MustInsertOneInvDoc(t, context.Background(), insertDoc)

	cases := []struct {
		cn            string
		ctx           context.Context
		id            string
		expectErr     bool
		valFunc       valInvitationFunc
		expectedErrIs error
	}{
		{
			cn:            "success",
			ctx:           context.Background(),
			id:            insertDoc.ID.Hex(),
			expectErr:     false,
			expectedErrIs: nil,
			valFunc: func(ctx context.Context, t *testing.T, cinv *bingo.Invitation) {
				require.NotNil(t, cinv.Game, "expected associated game to be loaded on invitation aggregate")

				inv, err := insertDoc.ToAggregate(nil)
				require.NoError(t, err, "expected no error from invitation aggregate conversion")
				MustCompareInvs(t, inv, cinv)
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

			inv, err := invRepo.Get(ctx, c.id)
			if c.expectErr {
				require.Error(t, err, "expected error")
				if c.expectedErrIs != nil {
					require.ErrorIs(t, err, c.expectedErrIs, "expected different error")
				}
			} else {
				require.NoError(t, err, "expected no error")
				c.valFunc(ctx, t, inv)
			}
		})
	}
}

func TestInvitationRepository_Save(t *testing.T) {
	invRepo := mongo.NewInvitationRepository(sharedDB)
	commonSub := &bingo.Invitation{
		DeliveryMethod: bingo.InvitationDeliveryMethodDownload,
		MaxCardAmount:  3,
		Active:         true,
		GameID:         primitive.NewObjectID().Hex(),
		Game:           nil,
		Criteria: []bingo.InvitationCriterion{
			{
				Kind:  bingo.InvitationCriterionKindRegex,
				Field: bingo.InvitationCriterionFieldEmail,
				Value: "regex",
			},
		},
	}

	cases := []struct {
		cn            string
		ctx           context.Context
		inv           *bingo.Invitation
		beforeFunc    valInvitationFunc
		valFunc       valInvitationFunc
		expectErr     bool
		expectedErrIs error
	}{
		{
			cn:         "insert success",
			ctx:        context.Background(),
			inv:        commonSub,
			beforeFunc: nil,
			valFunc: func(ctx context.Context, t *testing.T, inv *bingo.Invitation) {
				require.NotEmpty(t, inv.ID, commonSub.ID, "expected assigned id for non-inserted invitation to be set")

				//  Make sure doc is inserted correctly
				doc := MustFindOneInvDoc(t, ctx, inv.ID)
				cinv, err := doc.ToAggregate(nil)
				require.NoError(t, err, "expected no error from invitation aggregate conversion")
				MustCompareInvs(t, inv, cinv)

			},
			expectErr:     false,
			expectedErrIs: nil,
		},
		{
			cn:  "update success",
			ctx: context.Background(),
			inv: commonSub,
			beforeFunc: func(ctx context.Context, t *testing.T, inv *bingo.Invitation) {
				// Mutate invitation
				inv.Active = false
				inv.DeliveryMethod = bingo.InvitationDeliveryMethodMail
				inv.Criteria = []bingo.InvitationCriterion{}
				inv.MaxCardAmount = 6
			},
			valFunc: func(ctx context.Context, t *testing.T, inv *bingo.Invitation) {
				//  Make sure doc is inserted correctly
				doc := MustFindOneInvDoc(t, ctx, inv.ID)
				cinv, err := doc.ToAggregate(nil)
				require.NoError(t, err, "expected no error from invitation aggregate conversion")
				MustCompareInvs(t, inv, cinv)
			},
			expectErr:     false,
			expectedErrIs: nil,
		},
		{
			cn:  "insert fail id not hex",
			ctx: context.Background(),
			inv: &bingo.Invitation{
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
				c.beforeFunc(ctx, t, c.inv)
			}

			err := invRepo.Save(ctx, c.inv)
			if c.expectErr {
				require.Error(t, err, "expected error")
				if c.expectedErrIs != nil {
					require.ErrorIs(t, err, c.expectedErrIs, "expected different error")
				}
			} else {
				require.NoError(t, err, "expected no error")
				c.valFunc(ctx, t, c.inv)
			}
		})
	}
}

func MustInsertOneInvDoc(tb testing.TB, ctx context.Context, doc mongo.DocInvitation) {
	tb.Helper()

	_, err := sharedDB.Invitations.InsertOne(ctx, doc)
	require.NoError(tb, err, "expected no error from inserting invitation")
}

func MustFindOneInvDoc(tb testing.TB, ctx context.Context, id string) mongo.DocInvitation {
	tb.Helper()

	oid, err := primitive.ObjectIDFromHex(id)
	require.NoError(tb, err, "expected valid object id")
	var doc mongo.DocInvitation
	err = sharedDB.Invitations.FindOne(ctx, bson.M{"_id": oid}).Decode(&doc)
	require.NoError(tb, err, "expected no error when trying to find given invitation")
	return doc
}

func MustCompareInvs(tb testing.TB, exp, cmp *bingo.Invitation) {
	tb.Helper()

	// Comparing game is not relevant when comparing invitations
	exp.Game = nil
	cmp.Game = nil
	require.EqualValues(tb, exp, cmp, "expected invitations in comparison to be equal")
}
