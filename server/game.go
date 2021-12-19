package bingo

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"runtime"
	"sort"
	"sync"
	"time"
)

var (
	ErrGameNotFound = errors.New("game: game could not be found")
)

type GameRepository interface {
	Save(ctx context.Context, game *Game) error
	Get(ctx context.Context, id string) (*Game, error)
}

type GameService struct {
	gameRepo GameRepository
	cardRepo CardRepository
}

// Creates a new game and saves it.
func (gs *GameService) Create(ctx context.Context, hostId string, name string) (*Game, error) {
	g := CreateGame(hostId, name)

	// Try saving the new game
	if err := gs.gameRepo.Save(ctx, g); err != nil {
		return nil, err
	}

	return g, nil
}

// Creates a new game, saves it, and generates cards for it.
func (gs *GameService) CreateWithCards(ctx context.Context, hostId, name string, cardAmount int) (*Game, []Card, error) {
	g, err := gs.Create(ctx, hostId, name)
	if err != nil {
		return nil, nil, err
	}

	cards, err := gs.GenerateCards(ctx, g.ID, cardAmount)

	return g, cards, nil
}

func (gs *GameService) GenerateCards(ctx context.Context, id string, amount int) ([]Card, error) {
	// Try to get game from id
	g, err := gs.gameRepo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	// Generate random cards for game
	cards, nextCardNum := g.GenerateBulkRandomCards(amount)

	// Save all the new cards that has been generated
	err = gs.cardRepo.SaveAll(ctx, cards)
	if err != nil {
		return nil, err
	}

	// Save new next card number for game
	g.NextCardNumber = nextCardNum
	err = gs.gameRepo.Save(ctx, g)
	if err != nil {
		return nil, err
	}

	return cards, nil
}

// Calls a new number in the game identified by the id and saves it.
func (gs *GameService) CallNumber(ctx context.Context, id string, num int) (*Game, error) {
	// Try to get game from id
	g, err := gs.gameRepo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	// Try to call new number in game
	if err = g.CallNumber(num); err != nil {
		return nil, err
	}

	// Try saving game after calling new number
	if err = gs.gameRepo.Save(ctx, g); err != nil {
		return nil, err
	}

	return g, nil
}

// Matches winning card patterns against the card identified by cardId in game identified by gameId.
func (gs *GameService) MatchWinningCardPattern(ctx context.Context, cardNum int, gameId string) ([]int, error) {
	// Try to get game from id
	g, err := gs.gameRepo.Get(ctx, gameId)
	if err != nil {
		return nil, err
	}

	// Try to get card from id
	c, err := gs.cardRepo.GetByNumber(ctx, cardNum, gameId)
	if err != nil {
		return nil, err
	}

	// Try to match winning patterns on given card
	matches, err := g.MatchWinningCardPatterns(c)
	if err != nil {
		return nil, err
	}

	return matches, nil
}

// Instantiate new game service with dependencies
func NewGameService(gameRepo GameRepository, cardRepo CardRepository) *GameService {
	return &GameService{
		gameRepo: gameRepo,
		cardRepo: cardRepo,
	}
}

// Game entity containing information about games created.
type Game struct {
	ID   string `json:"id"`
	Name string `json:"name"`

	// Identify the user host of game. This is also the owner of the game entity
	HostId string `json:"hostId"`
	Host   *User  `json:"host"`

	NextCardNumber int

	CalledNumbers []Ball `json:"calledNumbers"`

	UpdatedAt time.Time `json:"updatedAt"`
	CreatedAt time.Time `json:"createdAt"`
}

func (g *Game) CreateRandomCard(rs rand.Source, number int) *Card {
	return CreateRandomCard(rs, g.ID, number)
}

// Generate a random card and append it to cards array.
func (g *Game) GenerateRandomCard() *Card {
	rs := rand.NewSource(time.Now().UnixNano())

	c := g.CreateRandomCard(rs, g.NextCardNumber)
	g.NextCardNumber++
	return c
}

// Generate the amount of cards specified and assign them to the game.
// The cards are generated asynchronously, but the API is synchronous for simplicity.
func (g *Game) GenerateBulkRandomCards(amount int) (cards []Card, nextCardNum int) {

	// Create cards concurrently to hopefully get a performance increase from parallism
	cardChan := make(chan Card, amount)
	go g.queueConcurrentCardGeneration(cardChan, amount)

	cards = make([]Card, 0, amount)
	nextCardNum = g.NextCardNumber + amount

	// Append created bingo cards to slice
	for card := range cardChan {
		cards = append(cards, card)
	}

	// Sort by card number
	sort.Slice(cards, func(i, j int) bool {
		return cards[i].Number < cards[j].Number
	})

	return cards, nextCardNum
}

// Distribute card generation onto goroutines amounting to the available cores on the hardware.
func (g *Game) queueConcurrentCardGeneration(c chan Card, amount int) {
	numCpu := runtime.NumCPU()
	tasksPrRoutine := int(math.Floor(float64(amount / numCpu)))
	restTasks := amount % numCpu

	var wg sync.WaitGroup
	wg.Add(numCpu)

	nextNum := g.NextCardNumber

	rs := rand.NewSource(time.Now().UnixNano() - int64(time.Second*5))
	for i := 0; i < restTasks; i++ {
		c <- *g.CreateRandomCard(rs, nextNum)

		nextNum++
	}

	// Launch goroutines on each CPU
	for i := 0; i < numCpu; i++ {
		rs = rand.NewSource(time.Now().UnixNano())

		go func(rs rand.Source, nextNum int) {

			// Generate the amount of cards assigned to this goroutine
			for j := 0; j < tasksPrRoutine; j++ {
				c <- *g.CreateRandomCard(rs, nextNum+j)
			}

			wg.Done()
		}(rs, nextNum+tasksPrRoutine*i)
	}

	// When all goroutines are done generating, close the channel from the writing side
	wg.Wait()
	close(c)
}

var (
	ErrCalledNumberExists = errors.New("game: number has already been called")
)

// Register a number into the already called number of the game
func (g *Game) CallNumber(num int) error {

	// Make sure called number has not already been registered
	for _, cn := range g.CalledNumbers {
		if num == cn.Number {
			return ErrCalledNumberExists
		}
	}

	// Add newly called number to existing called numbers
	g.CalledNumbers = append(g.CalledNumbers, newBall(num))
	return nil
}

var (
	ErrCardDoesNotBelongToGame = errors.New("game: card does not belong to this game")
)

// Matches winning card patterns and return the rows that matched.
func (g *Game) MatchWinningCardPatterns(card *Card) ([]int, error) {
	// Make sure card belongs to game
	if card.GameID != g.ID {
		return nil, ErrCardDoesNotBelongToGame
	}

	// Register the called numbers in a map so it is easy and cheap to check if a given number has been called
	calledNumMap := make(map[int]bool)
	for _, cn := range g.CalledNumbers {
		calledNumMap[cn.Number] = true
	}

	matchedRows := make([]int, 0, 3)

	// Check card rows for pattern matches: If all 5 numbers on a given row has been called then the row is match
	m := card.Matrix()
	for ri, colVals := range m {
		matches := 0
		for _, n := range colVals {
			if _, ok := calledNumMap[n]; ok {
				matches++
			}
		}

		// If there has been five matches then we have a match for the given line
		if matches == 5 {
			row := ri + 1
			matchedRows = append(matchedRows, row)
		}
	}

	return matchedRows, nil
}

// Game constructor
func CreateGame(hostId string, name string) *Game {
	now := time.Now()
	return &Game{
		HostId:         hostId,
		NextCardNumber: 1,
		Name:           name,
		UpdatedAt:      now,
		CreatedAt:      now,
	}
}
