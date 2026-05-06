package test

import (
	"fmt"
	"testing"

	"poker-engine/pkg/card"
	"poker-engine/pkg/engine"
	"poker-engine/pkg/seat"
	"poker-engine/pkg/table"
)

// All-in + Side Pot 测试，确保全下玩家只能赢主池，侧池由其他玩家竞争
func TestPayout_SidePot(t *testing.T) {
	fmt.Println("====== All-in + Side Pot 测试用例 ======")

	seats := []*seat.Seat{
		{
			SeatID:     0,
			PlayerID:   1,
			Chips:      0,
			Status:     seat.AllIn,
			Cards:      []card.Card{MakeCard(0, 12), MakeCard(1, 12)}, // AA
			CurrentBet: 100,
			TotalBet:   100,
		},
		{
			SeatID:     1,
			PlayerID:   2,
			Chips:      0,
			Status:     seat.AllIn,
			Cards:      []card.Card{MakeCard(0, 11), MakeCard(1, 11)}, // KK
			CurrentBet: 200,
			TotalBet:   200,
		},
		{
			SeatID:     2,
			PlayerID:   3,
			Chips:      0,
			Status:     seat.AllIn,
			Cards:      []card.Card{MakeCard(0, 10), MakeCard(1, 10)}, // QQ
			CurrentBet: 200,
			TotalBet:   200,
		},
	}

	board := []card.Card{
		MakeCard(2, 2), MakeCard(2, 5),
		MakeCard(3, 7), MakeCard(1, 9),
		MakeCard(0, 3),
	}

	tbl := table.New(seats, 1, 2)
	handler := &MockHandler{
		Seats: seats,
		Board: board,
	}

	e := engine.New(tbl, handler)
	e.Board = board

	e.Payout()

	handler.PrintResult()

	if len(handler.Payouts) == 0 {
		t.Fatal("no payout")
	}

	results := handler.Payouts[0]

	if len(results) != 2 {
		t.Fatalf("expected 2 pots, got %d", len(results))
	}

	mainPot := results[0]
	sidePot := results[1]

	// 主池：100×3=300 → AA赢
	if mainPot.Amount != 300 {
		t.Fatalf("main pot wrong: %d", mainPot.Amount)
	}
	if mainPot.Winners[0] != 0 {
		t.Fatalf("main pot winner wrong: %v", mainPot.Winners)
	}

	// 边池：(200-100)*2=200 → KK赢
	if sidePot.Amount != 200 {
		t.Fatalf("side pot wrong: %d", sidePot.Amount)
	}
	if sidePot.Winners[0] != 1 {
		t.Fatalf("side pot winner wrong: %v", sidePot.Winners)
	}
}
