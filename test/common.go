package test

import "poker-engine/pkg/card"

func MakeCard(suit, value int) card.Card {
	return card.Card{
		Suit:  suit,
		Value: value,
	}
}
