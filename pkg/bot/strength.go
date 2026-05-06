package bot

import (
	"poker-engine/pkg/card"
	"poker-engine/pkg/evaluator"
)

func CalcStrength(hand []card.Card, board []card.Card) float64 {

	all := append([]card.Card{}, hand...)
	all = append(all, board...)

	// ⭐ Preflop（只有2张手牌）
	if len(all) < 5 {
		return preflopStrength(hand)
	}

	val := evaluator.Evaluate(all)

	strength := float64(val.Level) / 10.0

	if strength > 1 {
		strength = 1
	}

	return strength
}
func preflopStrength(hand []card.Card) float64 {

	if len(hand) != 2 {
		return 0.1
	}

	c1 := hand[0]
	c2 := hand[1]

	// ⭐ pair（对子）
	if c1.Value == c2.Value {
		return 0.7
	}

	// ⭐ suited（同花）
	if c1.Suit == c2.Suit {
		return 0.4
	}

	// ⭐ high cards（高牌）
	high := max(c1.Value, c2.Value)

	// A = 12, K = 11, Q = 10, J = 9 ...
	if high >= 10 { // Q+
		return 0.5
	}

	if high >= 8 { // 10+
		return 0.35
	}

	// ⭐ weak hand
	return 0.2
}
