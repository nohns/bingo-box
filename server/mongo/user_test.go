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

var implementsUserRepo bingo.UserRepository = &mongo.UserRepository{}

// Test that mongodb card player <-> player aggregate root conversion works
func TestDocUser(t *testing.T) {
	u := &bingo.User{
		ID:             primitive.NewObjectID().Hex(),
		Name:           "test name",
		Email:          "test@test.com",
		HashedPassword: []byte("test pass"),
		UpdatedAt:      time.Now(),
		CreatedAt:      time.Now(),
	}

	t.Run("test data out of date", func(t *testing.T) {
		pFieldsCount := reflect.Indirect(reflect.ValueOf(u)).NumField()
		expectedfc := 6
		require.Equal(t, expectedfc, pFieldsCount, "player test data missing one or more fields")
	})

	t.Run("bidirectional conversion", func(t *testing.T) {

		doc, err := mongo.DocFromUser(u)
		require.NoError(t, err, "expected no error from mongo.DocFromUser")

		cu, err := doc.ToAggregate()
		require.NoError(t, err, "exptected no error from doc.ToAggregate()")
		require.EqualValues(t, u, cu, "expected values of round-trip conversion to equal initial data")
	})
}

type valUserFunc func(context.Context, *testing.T, *bingo.User)

func TestUserRepository_GetByEmail(t *testing.T) {
	userRepo := mongo.NewUserRepository(sharedDB)
	insertDoc := mongo.DocUser{
		ID:             primitive.NewObjectID(),
		Name:           "test name",
		Email:          "test@test.com",
		HashedPassword: []byte("test hash"),
		UpdatedAt:      time.Now(),
		CreatedAt:      time.Now(),
	}
	MustInsertOneUserDoc(t, context.Background(), insertDoc)

	cases := []struct {
		cn            string
		ctx           context.Context
		email         string
		expectErr     bool
		valFunc       valUserFunc
		expectedErrIs error
	}{
		{
			cn:            "success",
			ctx:           context.Background(),
			email:         insertDoc.Email,
			expectErr:     false,
			expectedErrIs: nil,
			valFunc: func(ctx context.Context, t *testing.T, cu *bingo.User) {
				u, err := insertDoc.ToAggregate()
				require.NoError(t, err, "expected no error from user aggregate conversion")
				MustCompareUsers(t, u, cu)
			},
		},
		{
			cn:            "fail no docs",
			ctx:           context.Background(),
			email:         "email@notfound.com",
			expectErr:     true,
			expectedErrIs: mongodb.ErrNoDocuments,
			valFunc:       nil,
		},
	}

	for _, c := range cases {
		t.Run(c.cn, func(t *testing.T) {
			ctx := c.ctx

			u, err := userRepo.GetByEmail(ctx, c.email)
			if c.expectErr {
				require.Error(t, err, "expected error")
				if c.expectedErrIs != nil {
					require.ErrorIs(t, err, c.expectedErrIs, "expected different error")
				}
			} else {
				require.NoError(t, err, "expected no error")
				c.valFunc(ctx, t, u)
			}
		})
	}
}

func TestUserRepository_Get(t *testing.T) {
	userRepo := mongo.NewUserRepository(sharedDB)
	insertDoc := mongo.DocUser{
		ID:             primitive.NewObjectID(),
		Name:           "test name",
		Email:          "test@test.com",
		HashedPassword: []byte("test hash"),
		UpdatedAt:      time.Now(),
		CreatedAt:      time.Now(),
	}
	MustInsertOneUserDoc(t, context.Background(), insertDoc)

	cases := []struct {
		cn            string
		ctx           context.Context
		id            string
		expectErr     bool
		valFunc       valUserFunc
		expectedErrIs error
	}{
		{
			cn:            "success",
			ctx:           context.Background(),
			id:            insertDoc.ID.Hex(),
			expectErr:     false,
			expectedErrIs: nil,
			valFunc: func(ctx context.Context, t *testing.T, cu *bingo.User) {
				u, err := insertDoc.ToAggregate()
				require.NoError(t, err, "expected no error from user aggregate conversion")
				MustCompareUsers(t, u, cu)
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

			u, err := userRepo.Get(ctx, c.id)
			if c.expectErr {
				require.Error(t, err, "expected error")
				if c.expectedErrIs != nil {
					require.ErrorIs(t, err, c.expectedErrIs, "expected different error")
				}
			} else {
				require.NoError(t, err, "expected no error")
				c.valFunc(ctx, t, u)
			}
		})
	}
}

func TestUserRepository_Save(t *testing.T) {
	userRepo := mongo.NewUserRepository(sharedDB)
	commonSub := &bingo.User{
		Name:           "test name",
		Email:          "test@test.com",
		HashedPassword: []byte("pass"),
		UpdatedAt:      time.Now(),
		CreatedAt:      time.Now(),
	}

	cases := []struct {
		cn            string
		ctx           context.Context
		u             *bingo.User
		beforeFunc    valUserFunc
		valFunc       valUserFunc
		expectErr     bool
		expectedErrIs error
	}{
		{
			cn:         "insert success",
			ctx:        context.Background(),
			u:          commonSub,
			beforeFunc: nil,
			valFunc: func(ctx context.Context, t *testing.T, u *bingo.User) {
				require.NotEmpty(t, u.ID, commonSub.ID, "expected assigned id for non-inserted user to be set")

				//  Make sure doc is inserted correctly
				doc := MustFindOneUserDoc(t, ctx, u.ID)
				cu, err := doc.ToAggregate()
				require.NoError(t, err, "expected no error from user aggregate conversion")
				MustCompareUsers(t, u, cu)

			},
			expectErr:     false,
			expectedErrIs: nil,
		},
		{
			cn:  "update success",
			ctx: context.Background(),
			u:   commonSub,
			beforeFunc: func(ctx context.Context, t *testing.T, u *bingo.User) {
				// Mutate user
				u.Email = "test2@test.com"
				u.HashedPassword = []byte("new hash")
				u.Name = "new name"
			},
			valFunc: func(ctx context.Context, t *testing.T, u *bingo.User) {
				//  Make sure doc is inserted correctly
				doc := MustFindOneUserDoc(t, ctx, u.ID)
				cu, err := doc.ToAggregate()
				require.NoError(t, err, "expected no error from user aggregate conversion")
				MustCompareUsers(t, u, cu)
			},
			expectErr:     false,
			expectedErrIs: nil,
		},
		{
			cn:  "insert fail id not hex",
			ctx: context.Background(),
			u: &bingo.User{
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
				c.beforeFunc(ctx, t, c.u)
			}

			err := userRepo.Save(ctx, c.u)
			if c.expectErr {
				require.Error(t, err, "expected error")
				if c.expectedErrIs != nil {
					require.ErrorIs(t, err, c.expectedErrIs, "expected different error")
				}
			} else {
				require.NoError(t, err, "expected no error")
				c.valFunc(ctx, t, c.u)
			}
		})
	}
}

func MustInsertOneUserDoc(tb testing.TB, ctx context.Context, doc mongo.DocUser) {
	tb.Helper()

	_, err := sharedDB.Users.InsertOne(ctx, doc)
	require.NoError(tb, err, "expected no error from inserting user")
}

func MustFindOneUserDoc(tb testing.TB, ctx context.Context, id string) mongo.DocUser {
	tb.Helper()

	oid, err := primitive.ObjectIDFromHex(id)
	require.NoError(tb, err, "expected valid object id")
	var doc mongo.DocUser
	err = sharedDB.Users.FindOne(ctx, bson.M{"_id": oid}).Decode(&doc)
	require.NoError(tb, err, "expected no error when trying to find given user")
	return doc
}

func MustCompareUsers(tb testing.TB, exp, cmp *bingo.User) {
	tb.Helper()

	// Workaround for differences in precision between go and mongodb time precision
	if cmp.CreatedAt.UnixMilli() == exp.CreatedAt.UnixMilli() {
		cmp.CreatedAt = exp.CreatedAt
	}
	if cmp.UpdatedAt.UnixMilli() == exp.UpdatedAt.UnixMilli() {
		cmp.UpdatedAt = exp.UpdatedAt
	}
	require.EqualValues(tb, exp, cmp, "expected users in comparison to be equal")
}
