package bingo_test

import (
	"context"
	"errors"
	"math/rand"
	"testing"
	"time"

	bingo "github.com/nohns/bingo-box/server"
	"github.com/nohns/bingo-box/server/mock"
	"github.com/nohns/bingo-box/server/requiretest"
	"github.com/stretchr/testify/require"
)

func TestGame_GenerateBulkRandomCards(t *testing.T) {

	hostId, gameName, nextNum := "7c50d3ad-0b0e-45b7-89fd-dc994c40c73d", "test", 1
	g := bingo.CreateGame(hostId, gameName)
	require.Equal(t, nextNum, g.NextCardNumber, "unexpected next card number before generate")

	// Generate many random cards and check them
	amount := 10000
	cards, nextCardNum := g.GenerateBulkRandomCards(amount)
	if cardsLen := len(cards); cardsLen != amount {
		t.Fatalf("generated cards len was %d, expected %d", cardsLen, amount)
	}

	require.Equal(t, gameName, g.Name, "unexpected game name")
	require.Equal(t, hostId, g.HostId, "unexpected host id")

	numBefore := nextNum - 1
	for _, c := range cards {
		require.Equal(t, numBefore+1, c.Number, "card num not expected")
		MustBeValidCardGridNums(t, &c)

		numBefore = c.Number
	}

	require.Equal(t, numBefore+1, nextCardNum, "unexpected next card number after generate")
}

func TestGame_MatchWinningCardPatterns(t *testing.T) {

	t.Run("match rows one for one", func(t *testing.T) {
		testGame := MustMakeTestGame(t)

		rs := rand.NewSource(1)
		card := testGame.CreateRandomCard(rs, testGame.NextCardNumber)

		rowNums := [][]int{
			make([]int, 0, 5),
			make([]int, 0, 5),
			make([]int, 0, 5),
		}
		for _, cgn := range card.GridNumbers {
			rowNums[cgn.Row-1] = append(rowNums[cgn.Row-1], cgn.Number)
		}

		for ri, nums := range rowNums {
			row := ri + 1
			require.Len(t, nums, 5, "row len must be 5")
			for _, n := range nums {
				testGame.CallNumber(n)
			}

			rowMatches, err := testGame.MatchWinningCardPatterns(card)
			require.NoError(t, err, "unexpected error when matching card patterns after calling all number from row %d and above", row)
			require.Contains(t, rowMatches, row, "all numbers from row %d was called but not matched", row)
			require.Len(t, rowMatches, row, "previously matched rows not included in new matches")
		}
	})

	t.Run("match no rows", func(t *testing.T) {
		testGame := MustMakeTestGame(t)

		rs := rand.NewSource(1)
		card := testGame.CreateRandomCard(rs, testGame.NextCardNumber)

		// Blacklist first number of each row
		var rowBlacklisted [3]bool
		blacklistNums := make(map[int]bool)
		for _, cgn := range card.GridNumbers {
			if row := cgn.Row; rowBlacklisted[row-1] {
				continue
			}

			blacklistNums[cgn.Number] = true
			rowBlacklisted[cgn.Row-1] = true
		}

		// Call numbers 1-90, excluding blacklisted num
		for num := 1; num <= 90; num++ {
			if bl, ok := blacklistNums[num]; ok && bl {
				continue
			}

			testGame.CallNumber(num)
		}

		rowMatches, err := testGame.MatchWinningCardPatterns(card)
		require.NoError(t, err, "unexpected error when matching card patterns when no rows should match")
		require.Len(t, rowMatches, 0, "expectedly found %d row matches, when expecting none", len(rowMatches))
	})
}

func TestGameService_Create(t *testing.T) {

	var ErrGameRepo = errors.New("repo: error occurred")

	cases := []struct {
		caseName    string
		hostId      string
		gameName    string
		saveHandler mock.GameSaveHandler
		expectErr   bool
		expectedErr error
	}{
		{
			"success",
			requiretest.UUIDv4(t),
			"success game",
			func(_ context.Context, g *bingo.Game) error {
				g.ID = requiretest.UUIDv4(t)
				g.CreatedAt = time.Now()
				g.UpdatedAt = time.Now()

				return nil
			},
			false,
			nil,
		},
		{
			"repo error",
			requiretest.UUIDv4(t),
			"failing game",
			func(_ context.Context, g *bingo.Game) error {
				return ErrGameRepo
			},
			true,
			ErrGameRepo,
		},
	}

	for _, tc := range cases {
		t.Run(tc.caseName, func(t *testing.T) {
			gameSvc, mocks := MustCreateGameService(t)
			defer mocks.gameRepo.RequireExpectationsMet()

			mocks.gameRepo.ExpectSave(tc.saveHandler)

			g, err := gameSvc.Create(context.Background(), tc.hostId, tc.gameName)
			if tc.expectErr {
				require.Nil(t, g, "game must be nil when error is expected")
				require.Error(t, err, "error must be set when error is expected")

				if tc.expectedErr == nil {
					return
				}

				require.ErrorIs(t, err, tc.expectedErr, "error must be of expected error kind")
			} else {
				require.NoError(t, err, "no error is expected")
				require.NotNil(t, g, "game must not be nil when no error is expected. Both game and error must not be nil at the same times")

				var expectedType *bingo.Game
				require.IsType(t, expectedType, g, "game must be of type *bingo.Game")

				// Validate
				require.NotEmpty(t, g.ID, "game id must be set")
				require.Equal(t, tc.hostId, g.HostId, "host id must be the same as the input given")
				require.Equal(t, tc.gameName, g.Name, "game name must be the same as the input given")
				require.Equal(t, 1, g.NextCardNumber, "next card number must be 1 because no cards are generated yet")
				require.NotEmpty(t, g.UpdatedAt, "updatedAt must be set")
				require.NotEmpty(t, g.CreatedAt, "updatedAt must be set")
				require.LessOrEqual(t, g.CreatedAt.UnixMilli(), g.UpdatedAt.UnixMilli(), "createdAt must be before updatedAt")
			}
		})
	}
}

func TestGameService_CallNumber(t *testing.T) {

	var ErrGameRepo = errors.New("repo: error occurred")

	testGame := MustMakeTestGame(t)

	testGameGetHandler := MakeSingleGameGetHandler(t, *testGame)
	updatedAtSaveHandler := MakeGameSaveHandler(t)

	cases := []struct {
		caseName     string
		gameId       string
		calledNumber int
		getHandler   mock.GameGetHandler
		saveHandler  mock.GameSaveHandler
		expectErr    bool
		expectedErr  error
	}{
		{
			caseName:     "success",
			gameId:       testGame.ID,
			calledNumber: 1,
			getHandler:   testGameGetHandler,
			saveHandler:  updatedAtSaveHandler,
			expectErr:    false,
			expectedErr:  nil,
		},
		{
			caseName:     "bingo error called number exists",
			gameId:       testGame.ID,
			calledNumber: 11,
			getHandler:   testGameGetHandler,
			saveHandler:  nil,
			expectErr:    true,
			expectedErr:  bingo.ErrCalledNumberExists,
		},
		{
			caseName:     "get error game not found",
			gameId:       "",
			calledNumber: 1,
			getHandler:   testGameGetHandler,
			saveHandler:  nil,
			expectErr:    true,
			expectedErr:  bingo.ErrGameNotFound,
		},
		{
			caseName:     "save error",
			gameId:       testGame.ID,
			calledNumber: 2,
			getHandler:   testGameGetHandler,
			saveHandler:  func(ctx context.Context, game *bingo.Game) error { return ErrGameRepo },
			expectErr:    true,
			expectedErr:  ErrGameRepo,
		},
	}

	for _, tc := range cases {
		t.Run(tc.caseName, func(t *testing.T) {
			gameSvc, mocks := MustCreateGameService(t)
			defer mocks.gameRepo.RequireExpectationsMet()

			mocks.gameRepo.ExpectGet(tc.getHandler)
			if tc.saveHandler != nil {
				mocks.gameRepo.ExpectSave(tc.saveHandler)
			}

			g, err := gameSvc.CallNumber(context.Background(), tc.gameId, tc.calledNumber)
			if tc.expectErr {
				require.Nil(t, g, "game must be nil when error is expected")
				require.Error(t, err, "error must be set when error is expected")

				if tc.expectedErr == nil {
					return
				}

				require.ErrorIs(t, err, tc.expectedErr, "error must be of expected error kind")
			} else {
				require.NoError(t, err, "no error is expected")
				require.NotNil(t, g, "game must not be nil when no error is expected. Both game and error must not be nil at the same times")

				var expectedType *bingo.Game
				require.IsType(t, expectedType, g, "game must be of type *bingo.Game")
			}
		})
	}
}

func TestGameService_GenerateCards(t *testing.T) {

	var ErrGameRepo = errors.New("repo: error occurred")

	testGame := MustMakeTestGame(t)

	gameGetHandler := MakeSingleGameGetHandler(t, *testGame)
	gameSaveHandler := MakeGameSaveHandler(t)
	cardSaveAllHandler := func(ctx context.Context, cards []bingo.Card) error {
		for _, c := range cards {
			require.NotEmpty(t, c.GameID, "cardsaveallhandler: a game id must be present before cards can be saved")
		}
		return nil
	}

	cases := []struct {
		caseName           string
		gameId             string
		cardAmount         int
		gameGetHandler     mock.GameGetHandler
		gameSaveHandler    mock.GameSaveHandler
		cardSaveAllHandler mock.CardSaveAllHandler
		expectErr          bool
		expectedErr        error
	}{
		{
			caseName:           "success",
			gameId:             testGame.ID,
			cardAmount:         3,
			gameGetHandler:     gameGetHandler,
			gameSaveHandler:    gameSaveHandler,
			cardSaveAllHandler: cardSaveAllHandler,
			expectErr:          false,
			expectedErr:        nil,
		},
		{
			caseName:           "get error game not found",
			gameId:             "",
			cardAmount:         3,
			gameGetHandler:     gameGetHandler,
			cardSaveAllHandler: nil,
			gameSaveHandler:    nil,
			expectErr:          true,
			expectedErr:        bingo.ErrGameNotFound,
		},
		{
			caseName:           "bingo error card number exists",
			gameId:             testGame.ID,
			cardAmount:         3,
			gameGetHandler:     gameGetHandler,
			cardSaveAllHandler: func(ctx context.Context, cards []bingo.Card) error { return bingo.ErrCardNumberExists },
			expectErr:          true,
			expectedErr:        bingo.ErrCardNumberExists,
		},
		{
			caseName:           "save game error",
			gameId:             testGame.ID,
			cardAmount:         3,
			gameGetHandler:     gameGetHandler,
			cardSaveAllHandler: cardSaveAllHandler,
			gameSaveHandler:    func(ctx context.Context, game *bingo.Game) error { return ErrGameRepo },
			expectErr:          true,
			expectedErr:        ErrGameRepo,
		},
	}

	for _, tc := range cases {
		t.Run(tc.caseName, func(t *testing.T) {
			gameSvc, mocks := MustCreateGameService(t)
			defer mocks.gameRepo.RequireExpectationsMet()

			mocks.gameRepo.ExpectGet(tc.gameGetHandler)
			if tc.cardSaveAllHandler != nil {
				mocks.cardRepo.ExpectSaveAll(tc.cardSaveAllHandler)
			}
			if tc.gameSaveHandler != nil {
				mocks.gameRepo.ExpectSave(tc.gameSaveHandler)
			}

			cards, err := gameSvc.GenerateCards(context.Background(), tc.gameId, tc.cardAmount)
			if tc.expectErr {
				require.Nil(t, cards, "game must be nil when error is expected")
				require.Error(t, err, "error must be set when error is expected")

				if tc.expectedErr == nil {
					return
				}

				require.ErrorIs(t, err, tc.expectedErr, "error must be of expected error kind")
			} else {
				require.NoError(t, err, "no error is expected")
				require.NotNil(t, cards, "game must not be nil when no error is expected. Both game and error must not be nil at the same times")

				var expectedType []bingo.Card
				require.IsType(t, expectedType, cards, "game must be of type *bingo.Game")
				require.Len(t, cards, tc.cardAmount, "len of generated cards must be same as card amount")
				for _, c := range cards {
					require.Equal(t, tc.gameId, c.GameID, "a card's gameid must be the same as the game it was generated for")
				}
			}
		})
	}
}

type gameServiceMocks struct {
	gameRepo *mock.GameRepository
	cardRepo *mock.CardRespository
}

func MustCreateGameService(tb testing.TB) (*bingo.GameService, *gameServiceMocks) {
	tb.Helper()

	gameRepo := mock.NewGameRepository(tb)
	cardRepo := mock.NewCardRepository(tb)

	gameSvc := bingo.NewGameService(gameRepo, cardRepo)
	mocks := &gameServiceMocks{
		gameRepo: gameRepo,
		cardRepo: cardRepo,
	}

	return gameSvc, mocks
}

func MakeGameSaveHandler(tb testing.TB) mock.GameSaveHandler {
	tb.Helper()

	return func(ctx context.Context, game *bingo.Game) error {
		if game.ID == "" {
			game.ID = requiretest.UUIDv4(tb)
			game.CreatedAt = time.Now()
		}

		game.UpdatedAt = time.Now()
		return nil
	}
}

func MakeSingleGameGetHandler(tb testing.TB, game bingo.Game) mock.GameGetHandler {
	tb.Helper()

	return func(_ context.Context, id string) (*bingo.Game, error) {
		if id != game.ID {
			return nil, bingo.ErrGameNotFound
		}

		return &game, nil
	}
}

func MustMakeTestGame(tb testing.TB) *bingo.Game {
	tb.Helper()

	return &bingo.Game{
		ID:             requiretest.UUIDv4(tb),
		Name:           "new game",
		HostId:         requiretest.UUIDv4(tb),
		Host:           nil,
		NextCardNumber: 1,
		CalledNumbers:  []bingo.Ball{{11}, {45}, {67}},
		UpdatedAt:      time.Now(),
		CreatedAt:      time.Now().Add(-5 * time.Hour),
	}
}
