package mock

import (
	"context"
	"testing"

	bingo "github.com/nohns/bingo-box/server"
	"github.com/stretchr/testify/require"
)

type GameRepository struct {
	tb           testing.TB
	saveVisited  int
	saveExpected int
	saveHandlers []GameSaveHandler

	getVisited  int
	getExpected int
	getHandlers []GameGetHandler
}

type GameSaveHandler func(ctx context.Context, game *bingo.Game) error
type GameGetHandler func(ctx context.Context, id string) (*bingo.Game, error)

func (gr *GameRepository) ExpectSave(h GameSaveHandler) {
	gr.saveHandlers = append(gr.saveHandlers, h)
	gr.saveExpected++
}

func (gr *GameRepository) ExpectGet(h GameGetHandler) {
	gr.getHandlers = append(gr.getHandlers, h)
	gr.getExpected++
}

func (gr *GameRepository) Save(ctx context.Context, game *bingo.Game) error {
	require.Less(gr.tb, gr.saveVisited, gr.saveExpected, "mock(game_repository): Save() called more times than expected")
	h := gr.saveHandlers[gr.saveVisited]
	gr.saveVisited++

	return h(ctx, game)
}

func (gr *GameRepository) Get(ctx context.Context, id string) (*bingo.Game, error) {
	require.Less(gr.tb, gr.getVisited, gr.getExpected, "mock(game_repository): Get() called more times than expected")
	h := gr.getHandlers[gr.getVisited]
	gr.getVisited++

	return h(ctx, id)
}

func (gr *GameRepository) RequireExpectationsMet() {
	require.Equal(gr.tb, gr.saveExpected, gr.saveVisited, "mock(game_repository): Save() call expectations was not met.")
	require.Equal(gr.tb, gr.getExpected, gr.getVisited, "mock(game_repository): Get() call expectations was not met.")
}

func NewGameRepository(tb testing.TB) *GameRepository {
	return &GameRepository{
		tb:           tb,
		saveHandlers: make([]GameSaveHandler, 0, 1),
		getHandlers:  make([]GameGetHandler, 0, 1),
	}
}
