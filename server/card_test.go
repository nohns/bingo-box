package bingo_test

import (
	"fmt"
	"math"
	"math/rand"
	"testing"

	bingo "github.com/nohns/bingo-box/server"
	"github.com/nohns/bingo-box/server/requiretest"
	"github.com/stretchr/testify/require"
)

func TestCard_Numbers(t *testing.T) {
	rs := rand.NewSource(1)
	gameId, nextNum := requiretest.UUIDv4(t), 1
	c := bingo.CreateRandomCard(rs, gameId, nextNum)

	orderedNums := c.Numbers()

	require.Len(t, orderedNums, 15, "amount of numbers does not match card grid requirements")

	lastNum := 0
	for _, on := range orderedNums {
		require.Greater(t, on, lastNum, "unexpected lower or equal num in unique ascending ordered numbers")
		lastNum = on
	}
}

func TestCard_Matrix(t *testing.T) {
	rs := rand.NewSource(1)
	gameId, nextNum := requiretest.UUIDv4(t), 1
	c := bingo.CreateRandomCard(rs, gameId, nextNum)

	m := c.Matrix()

	t.Run("matrix array length", func(t *testing.T) {
		require.Len(t, m, 3, "matrix, unexpectedly, only has %d rows", len(m))

		for ri, cols := range m {
			require.Len(t, cols, 9, "matrix row index %d, unexpectedly, only has %d rows", ri, len(cols))
		}
	})
	t.Run("grid numbers match matrix", func(t *testing.T) {
		for _, cgn := range c.GridNumbers {

			require.GreaterOrEqual(t, cgn.Row, 1, "row is not inside range 1-3. row = ", cgn.Row)
			require.LessOrEqual(t, cgn.Row, 3, "row is not inside range 1-3. row = ", cgn.Row)

			require.GreaterOrEqual(t, cgn.Col, 1, "col is not inside range 1-9. col = ", cgn.Col)
			require.LessOrEqual(t, cgn.Col, 9, "col is not inside range 1-9. col = ", cgn.Col)

			num := m[cgn.Row-1][cgn.Col-1]
			require.Equal(t, cgn.Number, num, "num found in matrix did not match grid number struct")
		}
	})
}

func TestCreateRandomCard(t *testing.T) {

	rs := rand.NewSource(1)
	gameId, nextNum := requiretest.UUIDv4(t), 1
	c := bingo.CreateRandomCard(rs, gameId, nextNum)

	require.Empty(t, c.ID, "unexpected non-zero id on brand new card")
	require.Equal(t, gameId, c.GameID, "unexpected game id")
	require.Equal(t, nextNum, c.Number, "unexpected card num")
	require.NotNil(t, c.GridNumbers, "card nums slice must not be nil")

	MustBeValidCardGridNums(t, c)
}

func MustBeValidCardGridNums(t *testing.T, card *bingo.Card) {
	t.Helper()

	gridNums := card.GridNumbers
	uniqueNums := make(map[int]bool)
	uniqueSpots := make(map[string]int)
	rowHits := [3]int{}
	colHits := [9]int{}

	require.Len(t, gridNums, 15, "too many/few numbers for card")

	for _, cgn := range gridNums {
		require.NotContains(t, uniqueNums, cgn.Number, "duplicate card num %d", card.Number)

		spot := fmt.Sprintf("%d-%d", cgn.Row, cgn.Col)
		require.NotContains(t, uniqueSpots, spot, "duplicate card spot row %d col %d for %d", cgn.Row, cgn.Col, cgn.Number)

		require.GreaterOrEqual(t, cgn.Number, 1, "card num outside valid range")
		require.LessOrEqual(t, cgn.Number, 90, "card num outside valid range")

		// Clamp col guess from 1 to 9 and require the assigned col to match
		guessedCol := int(math.Min(math.Max(math.Floor(float64(cgn.Number/10)), 0), 8)) + 1
		require.Equal(t, cgn.Col, guessedCol, "col for num %d was expected to be %d but was %d", cgn.Number, guessedCol, cgn.Col)

		uniqueNums[cgn.Number] = true
		uniqueSpots[spot] = cgn.Number
		rowHits[cgn.Row-1]++
		colHits[cgn.Col-1]++
	}
	t.Run("should have five columns in each row", func(t *testing.T) {
		for i, hits := range rowHits {
			require.Equal(t, hits, 5, "cols in row index %d was only %d", i, hits)
		}
	})
	t.Run("should have one to three rows in each column", func(t *testing.T) {
		highBound := 3
		lowBound := 1

		for i, hits := range colHits {
			require.GreaterOrEqual(t, hits, lowBound, "rows in col index %d was %d, expected between %d and %d", i, hits, 1, lowBound)
			require.LessOrEqual(t, hits, highBound, "rows in col index %d was %d, expected between %d and %d", i, hits, 1, highBound)
		}
	})
}
