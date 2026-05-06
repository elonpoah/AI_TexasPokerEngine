package test

import (
	"fmt"
	"testing"

	"poker-engine/pkg/card"
	"poker-engine/pkg/engine"
	"poker-engine/pkg/seat"
	"poker-engine/pkg/table"
)

// 基础胜负测试
func TestPayout_SimpleWin(t *testing.T) {

	fmt.Println("====== 基础胜负测试用例 ======")

	seats := []*seat.Seat{
		{
			SeatID:     0,
			PlayerID:   1,
			Chips:      1000,
			Status:     seat.Active,
			Cards:      []card.Card{MakeCard(0, 12), MakeCard(1, 12)}, // AA
			CurrentBet: 100,
			TotalBet:   100,
		},
		{
			SeatID:     1,
			PlayerID:   2,
			Chips:      1000,
			Status:     seat.Active,
			Cards:      []card.Card{MakeCard(0, 11), MakeCard(1, 11)}, // KK
			CurrentBet: 100,
			TotalBet:   100,
		},
		{
			SeatID:     2,
			PlayerID:   3,
			Chips:      1000,
			Status:     seat.Folded,                                   // ❗关键
			Cards:      []card.Card{MakeCard(0, 10), MakeCard(1, 10)}, // QQ（本来更强）
			CurrentBet: 100,
			TotalBet:   100,
		},
	}

	board := []card.Card{
		MakeCard(2, 10), // Q
		MakeCard(3, 5),
		MakeCard(1, 2),
		MakeCard(0, 3),
		MakeCard(2, 7),
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

	// ✅ 安全检查
	if len(handler.Payouts) == 0 {
		t.Fatal("no payout event")
	}
	if len(handler.Payouts[0]) == 0 {
		t.Fatal("no pot generated")
	}

	res := handler.Payouts[0][0]

	if len(res.Winners) != 1 || res.Winners[0] != 0 {
		t.Fatalf("expected winner seat 0, got %v", res.Winners)
	}

	if res.Amount != 300 {
		t.Fatalf("expected pot 300, got %d", res.Amount)
	}

	if seats[0].Chips != 1300 {
		t.Fatalf("seat0 chips wrong: %d", seats[0].Chips)
	}
}
