package mock

import (
	"context"
	"testing"

	bingo "github.com/nohns/bingo-box/server"
	"github.com/stretchr/testify/require"
)

type CardSaveHandler func(ctx context.Context, card *bingo.Card) error
type CardSaveAllHandler func(ctx context.Context, cards []bingo.Card) error
type CardGetByNumberHandler func(ctx context.Context, cardNum int, gameId string) (*bingo.Card, error)

type CardRespository struct {
	tb            testing.TB
	savesExpected int
	savesExecuted int
	saveHandlers  []CardSaveHandler

	saveAllsExpected int
	saveAllsExecuted int
	saveAllHandlers  []CardSaveAllHandler

	getByNumbersExpected int
	getByNumbersExecuted int
	getByNumberHandlers  []CardGetByNumberHandler
}

func (gr *CardRespository) ExpectSave(h CardSaveHandler) {
	gr.saveHandlers = append(gr.saveHandlers, h)
	gr.savesExpected++
}

func (gr *CardRespository) ExpectSaveAll(h CardSaveAllHandler) {
	gr.saveAllHandlers = append(gr.saveAllHandlers, h)
	gr.saveAllsExpected++
}

func (gr *CardRespository) ExpectGetByNumber(h CardGetByNumberHandler) {
	gr.getByNumberHandlers = append(gr.getByNumberHandlers, h)
	gr.getByNumbersExpected++
}

func (gr *CardRespository) Save(ctx context.Context, card *bingo.Card) error {
	require.Less(gr.tb, gr.savesExecuted, gr.savesExpected, "mock(card_repository): Save() called more times than expected")

	h := gr.saveHandlers[gr.savesExecuted]
	gr.savesExecuted++

	return h(ctx, card)
}

func (gr *CardRespository) SaveAll(ctx context.Context, cards []bingo.Card) error {
	require.Less(gr.tb, gr.saveAllsExecuted, gr.saveAllsExpected, "mock(card_repository): SaveAll() called more times than expected")

	h := gr.saveAllHandlers[gr.saveAllsExecuted]
	gr.saveAllsExecuted++
	return h(ctx, cards)
}

func (gr *CardRespository) GetByNumber(ctx context.Context, cardNum int, gameId string) (*bingo.Card, error) {
	require.Less(gr.tb, gr.getByNumbersExecuted, gr.getByNumbersExpected, "mock(card_repository): GetByNumber() called more times than expected")

	h := gr.getByNumberHandlers[gr.getByNumbersExecuted]
	gr.getByNumbersExecuted++

	return h(ctx, cardNum, gameId)
}

func (cr *CardRespository) RequireExpectationsMet() {
	require.Equal(cr.tb, cr.savesExecuted, cr.savesExpected, "mock(game_repository): Save() was not called enough times")
	require.Equal(cr.tb, cr.saveAllsExecuted, cr.saveAllsExpected, "mock(game_repository): SaveAll() was not called enough times")
	require.Equal(cr.tb, cr.getByNumbersExecuted, cr.getByNumbersExpected, "mock(game_repository): GetByNumber() was not called enough times")
}

func NewCardRepository(tb testing.TB) *CardRespository {
	return &CardRespository{
		tb:                  tb,
		saveHandlers:        make([]CardSaveHandler, 0, 1),
		saveAllHandlers:     make([]CardSaveAllHandler, 0, 1),
		getByNumberHandlers: make([]CardGetByNumberHandler, 0, 1),
	}
}
