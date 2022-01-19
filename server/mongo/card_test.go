package mongo_test

import (
	"reflect"
	"testing"

	bingo "github.com/nohns/bingo-box/server"
	"github.com/nohns/bingo-box/server/mongo"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var implementsCardRepo bingo.CardRepository = &mongo.CardRepository{}

// Test that mongodb card doc <-> card aggregate root conversion works
func TestDocCard(t *testing.T) {
	c := &bingo.Card{
		ID:       primitive.NewObjectID().Hex(),
		Number:   1,
		GameID:   primitive.NewObjectID().Hex(),
		PlayerID: primitive.NewObjectID().Hex(),
		Player:   nil,
		GridNumbers: []bingo.CardGridNumber{
			{
				Row:    1,
				Col:    2,
				Number: 11,
			},
			{
				Row:    2,
				Col:    5,
				Number: 43,
			},
		},
	}

	t.Run("test data out of date", func(t *testing.T) {
		cFieldsCount := reflect.Indirect(reflect.ValueOf(c)).NumField()
		expectedfc := 6
		require.Equal(t, expectedfc, cFieldsCount, "game test data missing one or more fields")
	})

	t.Run("bidirectional conversion", func(t *testing.T) {
		doc, err := mongo.DocFromCard(c)
		require.NoError(t, err, "no error expected from mongo.DocFromCard")

		cc, err := doc.ToAggregate(c.Player)
		require.NoError(t, err, "no error exptected from doc.ToAggregate()")
		require.EqualValues(t, c, cc, "Expected values of round-trip conversion to equal initial data")
	})
}
