package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgconn"
	bingo "github.com/nohns/bingo-box/server"
)

type CardRepository struct {
	psql *DB
}

func (cr *CardRepository) Save(ctx context.Context, card bingo.Card) error {
	return cr.insert(ctx, []bingo.Card{card})
}

func (cr *CardRepository) SaveAll(ctx context.Context, cards []bingo.Card) error {
	return cr.insert(ctx, cards)
}

func (cr *CardRepository) GetByNumber(ctx context.Context, cardNum int, gameId string) (*bingo.Card, error) {
	// Try to form uuid from game id string
	gameUuid, err := uuid.FromString(gameId)
	if err != nil {
		return nil, err
	}

	// Find cards with game id and card number
	card, err := cr.findOne(ctx, "game_id = $1 AND number = $2", gameUuid, cardNum)
	if err != nil {
		return nil, err
	}

	return card, nil
}

func (cr *CardRepository) findOne(ctx context.Context, where string, param ...interface{}) (*bingo.Card, error) {
	card := new(bingo.Card)

	sql := "SELECT id, game_id, number FROM game_cards %s WHERE %s LIMIT 1"
	err := cr.psql.conn.QueryRow(ctx, sql, param...).Scan(&card.ID, &card.GameID, &card.Number)
	if err != nil {
		return nil, cr.translatePSQLError(err)
	}

	numbers, err := cr.findNumbers(ctx, card.ID)
	if err != nil {
		return nil, err
	}
	card.GridNumbers = numbers

	return card, nil
}

func (cr *CardRepository) findNumbers(ctx context.Context, cardId string) ([]bingo.CardGridNumber, error) {
	cardNums := make([]bingo.CardGridNumber, 0, bingo.CardGridNumbersCap)

	sql := "SELECT row, col, number FROM game_card_spots %s WHERE game_card_id = $1"
	rows, err := cr.psql.conn.Query(ctx, sql, cardId)
	if err != nil {
		return nil, cr.translatePSQLError(err)
	}
	defer rows.Close()

	for rows.Next() {
		var cardNum bingo.CardGridNumber
		rows.Scan(&cardNum.Row, &cardNum.Col, &cardNum.Number)
		cardNums = append(cardNums, cardNum)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return cardNums, nil
}

func (cr *CardRepository) insert(ctx context.Context, cards []bingo.Card) error {
	// Start transaction and make sure it is valid
	tx, err := cr.psql.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Build Insert values, params, and also card numbers map for inserting numbers
	insertColCount := 2
	values := make([]string, 0, len(cards))
	params := make([]interface{}, 0, len(cards)*insertColCount)
	cardNums := make(map[string][]bingo.CardGridNumber)
	for i, c := range cards {
		base := i * insertColCount
		insertCols := make([]string, 0, insertColCount)
		for paramNum := base; paramNum < base+insertColCount; paramNum++ {
			insertCols = append(insertCols, fmt.Sprintf("$%d", paramNum))
		}
		values = append(values, fmt.Sprintf("(%s)", strings.Join(insertCols, ", ")))
		params = append(params, c.GameID, c.Number)

		cardNums[c.ID] = c.GridNumbers
	}

	sql := fmt.Sprintf("INSERT INTO game_cards (game_id, number) VALUES \n%s", strings.Join(values, ",\n"))
	err = cr.psql.conn.QueryRow(ctx, sql, params...).Scan()
	if err != nil {
		return cr.translatePSQLError(err)
	}

	// Insert numbers
	err = cr.insertNumbers(ctx, cardNums)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (cr *CardRepository) insertNumbers(ctx context.Context, cardNums map[string][]bingo.CardGridNumber) error {
	// Start transaction and make sure it is valid
	tx, err := cr.psql.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	insertColCount := 4
	values := make([]string, 0, len(cardNums))
	params := make([]interface{}, 0, len(cardNums)*insertColCount)
	i := 0
	for cardId, cns := range cardNums {
		for _, cn := range cns {
			base := i * insertColCount
			insertCols := make([]string, 0, insertColCount)
			for paramNum := base; paramNum < base+insertColCount; paramNum++ {
				insertCols = append(insertCols, fmt.Sprintf("$%d", paramNum))
			}
			values = append(values, fmt.Sprintf("(%s)", strings.Join(insertCols, ", ")))
			params = append(params, cardId, cn.Row, cn.Col, cn.Number)
			i++
		}
	}

	sql := fmt.Sprintf("INSERT INTO game_card_spots (game_card_id, row, col, number) VALUES \n%s", strings.Join(values, ",\n"))
	err = cr.psql.conn.QueryRow(ctx, sql, params...).Scan()
	if err != nil {
		return cr.translatePSQLError(err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (cr *CardRepository) translatePSQLError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return err
	}

	// Handle postgres error codes
	switch pgErr.Code {
	case "23505": // Unique violation
		return bingo.ErrUserAlreadyExists
	default:
		return err
	}
}
