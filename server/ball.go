package bingo

import (
	"encoding/json"
	"strconv"
)

// Bingo ball value object
type Ball struct {
	Number int
}

func (b Ball) String() string {
	return strconv.Itoa(b.Number)
}

func (b Ball) MarshalJSON() ([]byte, error) {
	return json.Marshal(b.Number)
}

func (b Ball) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &b.Number)
}

func newBall(num int) Ball {
	return Ball{Number: num}
}
