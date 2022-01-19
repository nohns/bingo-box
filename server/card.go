package bingo

import (
	"context"
	"errors"
	"math/rand"
	"sort"
)

var (
	ErrCardNumberExists = errors.New("game: card number already exists in game")
)

type CardRepository interface {
	Save(ctx context.Context, card *Card) error
	SaveAll(ctx context.Context, cards []Card) error
	GetByNumber(ctx context.Context, cardNum int, gameId string) (*Card, error)
}

// Card entity / root aggregate for card related data
type Card struct {
	ID     string `json:"id"`
	Number int    `json:"number"`

	// Game in which the bingo card is used
	GameID string `json:"gameId"`

	// Player who owns the card, if it was generated via invitation
	PlayerID string  `json:"playerId,omitempty"`
	Player   *Player `json:"player,omitempty"`

	GridNumbers []CardGridNumber `json:"gridNumbers"`
}

func (c *Card) Validate() error {
	return nil
}

type CardMatrix [3][9]int

func (c *Card) Matrix() CardMatrix {
	var m CardMatrix

	for _, s := range c.GridNumbers {
		m[s.Row-1][s.Col-1] = s.Number
	}

	return m
}

func (c *Card) Numbers() []int {
	nums := make([]int, 0, len(c.GridNumbers))
	for _, cgn := range c.GridNumbers {
		nums = append(nums, cgn.Number)
	}

	sort.Ints(nums)
	return nums
}

func CreateRandomCard(rs rand.Source, gameId string, number int) *Card {
	return &Card{
		GameID:      gameId,
		Number:      number,
		GridNumbers: generateCardGridNumbers(rs),
	}
}

type CardGridNumber struct {
	Row    int `json:"row"`
	Col    int `json:"col"`
	Number int `json:"number"`
}

// Defining contants for card generation
const (
	CardGridNumbersCap = 15
)

// Generates a slice of random numbers distributed on card grid conforming to rules and constraints of 90-ball bingo cards:
//
// 1. Numbers 1 through 90 (including 1 and 90).
//
// 2. All cards contain exactly 15 numbers and 12 free slots.
//
// 3. Each column should have at least one number and each row must have 5 numbers.
//
// 4. Column x must only include numbers x * 10 through x * 10 + 9 (eg. column 4 must only include 40-49),
// with the exception of column 1 where it is x * 10 + 1 through x * 10 + 9 (excluding 0) and
// column 9 where it is x * 10 through x * 10 + 10 (including 90).
//
// 5. Columns must be arranged in ascending order.
//
func generateCardGridNumbers(rs rand.Source) []CardGridNumber {

	rnd := rand.New(rs)

	// Create card template with 15 numbers, 5 on each of the 3 rows.
	// This marks what spots on the bingo card should be filled out with numbers
	rndMatrix := [][]int{
		rnd.Perm(9)[:5],
		rnd.Perm(9)[:5],
		make([]int, 0, 5),
	}

	// Figure out the last row by seeing what columns has not been filled. All columns must have at least one number in one of the three rows.
	filledColIndicies := map[int]bool{
		0: false,
		1: false,
		2: false,
		3: false,
		4: false,
		5: false,
		6: false,
		7: false,
		8: false,
	}
	// Check row 1 and 2 to see what columns are filled
	for _, ci := range rndMatrix[0] {
		filledColIndicies[ci] = true
	}
	for _, ci := range rndMatrix[1] {
		filledColIndicies[ci] = true
	}
	missingIndicies := make([]int, 0, 5)
	for ci, filled := range filledColIndicies {
		if !filled {
			missingIndicies = append(missingIndicies, ci)
			rndMatrix[2] = append(rndMatrix[2], ci)
		}
	}
	rndRow := rnd.Perm(9)
	for _, rci := range rndRow {
		if len(rndMatrix[2]) == 5 {
			break
		}
		if !filledColIndicies[rci] {
			continue
		}

		// Random column index is not already missing
		rndMatrix[2] = append(rndMatrix[2], rci)
	}

	// Define arrays holding information about card
	var colCounts [9]int
	var colMatrix [9][3]int

	// Gather information about card for generating the card grid numbers and where to fill in the card grid numbers afterwards
	for ri, colIndicies := range rndMatrix {
		for _, ci := range colIndicies {
			// Increment the amount of different values to be generated for current column index
			colCounts[ci]++

			// Mark number filling spot in the column by column matrix, by assigning it to -1
			colMatrix[ci][ri] = -1
		}
	}

	// Define array for holding all generated card grid numbers in a random order at first
	numbers := make([]int, 0, CardGridNumbersCap)

	// Generate 15 random numbers for card
	for ci, cc := range colCounts {
		min, max := minMaxForColIndex(ci)
		for _, rndNum := range rnd.Perm(max - min)[:cc] {
			n := ci*10 + rndNum + 1
			numbers = append(numbers, n)
		}
	}

	// Sort card grid numbers in ascending order, so it is easy to just fill them in when going column by column
	sort.Slice(numbers, func(i, j int) bool {
		return numbers[i] < numbers[j]
	})

	cardGridNums := make([]CardGridNumber, 0, CardGridNumbersCap)

	// Iterating variable for numbers. Used to keep track of next number, when filling in card
	i := 0

	// Fill in all the marked spots with numbers column by column
	for ci, rows := range colMatrix {
		for ri, spot := range rows {
			// If spot in matrix is not marked as fillable (-1) then go to the next spot
			if spot != -1 {
				continue
			}

			cardGridNums = append(cardGridNums, CardGridNumber{
				Row:    ri + 1,
				Col:    ci + 1,
				Number: numbers[i],
			})

			i++
		}
	}

	return cardGridNums
}

// Get min and max range for generating card grid number
func minMaxForColIndex(colIndex int) (int, int) {
	// Special case for first (col index equals 0): range starts from 1 instead of 0, as 0 is not included in the overall valid range
	// Special case for last (col index equals 8): range ends with 10 instead of 9, as 90 is included in the overall valid range
	min := 0
	max := 9
	if colIndex == 0 {
		min = 1
	}
	if colIndex == 8 {
		max = 10
	}

	return min, max
}
