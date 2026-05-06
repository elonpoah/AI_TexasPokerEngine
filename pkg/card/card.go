package card

import "fmt"

type Card struct {
	Suit  int // 花色	0-3
	Value int // 点数	0-12
}

var values = [...]string{
	"2", "3", "4", "5", "6", "7",
	"8", "9", "10", "J", "Q", "K", "A",
}

var suits = [...]string{"♠", "♥", "♦", "♣"}

func (c Card) String() string {

	// ⭐ 防御式编程（非常重要）
	if c.Value < 0 || c.Value >= len(values) {
		return fmt.Sprintf("InvalidValue(%d)", c.Value)
	}

	if c.Suit < 0 || c.Suit >= len(suits) {
		return fmt.Sprintf("InvalidSuit(%d)", c.Suit)
	}

	return values[c.Value] + suits[c.Suit]
}
