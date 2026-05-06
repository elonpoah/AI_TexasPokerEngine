package test

import (
	"fmt"
	"testing"

	"poker-engine/pkg/card"
	"poker-engine/pkg/engine"
	"poker-engine/pkg/seat"
	"poker-engine/pkg/table"
)

// 测试分池情况，确保不同池的赢家正确分配
func TestPayout_SplitPot(t *testing.T) {

	fmt.Println("====== 奖金分池测试用例 ======")

	seats := []*seat.Seat{
		{
			SeatID:     0,
			PlayerID:   1,
			Chips:      1000,
			Status:     seat.Active,
			Cards:      []card.Card{MakeCard(0, 10), MakeCard(1, 9)},
			CurrentBet: 100,
			TotalBet:   100,
		},
		{
			SeatID:     1,
			PlayerID:   2,
			Chips:      1000,
			Status:     seat.Active,
			Cards:      []card.Card{MakeCard(2, 10), MakeCard(3, 9)},
			CurrentBet: 100,
			TotalBet:   100,
		},
	}

	board := []card.Card{
		MakeCard(0, 2), MakeCard(1, 3),
		MakeCard(2, 4), MakeCard(3, 5),
		MakeCard(0, 6),
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
	if len(handler.Payouts[0]) == 0 {
		t.Fatal("no pot")
	}

	res := handler.Payouts[0][0]

	if len(res.Winners) != 2 {
		t.Fatalf("expected split pot, got %v", res.Winners)
	}

	if seats[0].Chips != 1100 || seats[1].Chips != 1100 {
		t.Fatalf("split failed: %d %d", seats[0].Chips, seats[1].Chips)
	}
}
