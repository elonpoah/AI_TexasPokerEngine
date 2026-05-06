package bot

import "poker-engine/pkg/card"

func getStage(board []card.Card) string {
	switch len(board) {
	case 0:
		return "preflop"
	case 3:
		return "flop"
	case 4:
		return "turn"
	case 5:
		return "river"
	default:
		return "unknown"
	}
}
