package test

import (
	"fmt"
	"poker-engine/pkg/betting"
	"poker-engine/pkg/card"
	"poker-engine/pkg/engine"
	"poker-engine/pkg/seat"
)

type MockHandler struct {
	Payouts [][]engine.PayoutResult
	Seats   []*seat.Seat
	Board   []card.Card
}

func (m *MockHandler) OnDeal(userID int, seatID int, cards []card.Card)       {}
func (m *MockHandler) OnBoard(cards []card.Card)                              {}
func (m *MockHandler) OnAction(userID int, seatID int, action betting.Action) {}
func (m *MockHandler) OnRoundEnd()
func (m *MockHandler) OnTurn(pid, seatID, seconds int) {}

func (m *MockHandler) OnPayout(results []engine.PayoutResult) {
	m.Payouts = append(m.Payouts, results)
}

func seatStatusToString(s seat.Status) string {
	switch s {
	case seat.Active:
		return "Active"
	case seat.Folded:
		return "Folded"
	case seat.AllIn:
		return "All-in"
	default:
		return "Unknown"
	}
}
func (m *MockHandler) PrintResult() {
	m.printResult()
}
func (m *MockHandler) printResult() {

	printCard := func(c card.Card) string {
		suits := []string{"♠", "♥", "♦", "♣"}
		ranks := []string{"2", "3", "4", "5", "6", "7", "8", "9", "T", "J", "Q", "K", "A"}
		return ranks[c.Value] + suits[c.Suit]
	}

	// 公共牌
	fmt.Print("公共牌: ")
	for _, c := range m.Board {
		fmt.Print(printCard(c), " ")
	}
	fmt.Println()

	// =====================
	// 玩家信息（下注 + 手牌）
	// =====================
	fmt.Println("玩家信息:")
	for _, s := range m.Seats {
		if s == nil {
			continue
		}

		fmt.Printf("用户%d (Seat %d): ", s.PlayerID, s.SeatID)

		for _, c := range s.Cards {
			fmt.Print(printCard(c), " ")
		}

		fmt.Printf("| 状态: %s | TotalBet: %d | 当前筹码: %d\n",
			seatStatusToString(s.Status),
			s.TotalBet,
			s.Chips,
		)
	}

	// =====================
	// 统计每个玩家赢了多少
	// =====================
	winMap := make(map[int]int) // seatID -> 赢得金额

	for _, pots := range m.Payouts {
		for _, p := range pots {

			share := p.Amount / len(p.Winners)
			remainder := p.Amount % len(p.Winners)

			for i, sid := range p.Winners {
				win := share
				if i == 0 {
					win += remainder
				}
				winMap[sid] += win
			}
		}
	}

	// =====================
	// 奖池分配详情
	// =====================
	fmt.Println("\n奖金分配:")

	for i, pots := range m.Payouts {

		fmt.Printf("第 %d 次结算:\n", i+1)

		for j, p := range pots {

			fmt.Printf("  Pot #%d 金额: %d\n", j+1, p.Amount)

			for _, sid := range p.Winners {

				s := m.Seats[sid]

				fmt.Printf("    用户%d(Seat%d) 赢得: %d\n",
					s.PlayerID,
					sid,
					winMap[sid], // 👈 总赢
				)
			}
		}
	}

	// =====================
	// 盈亏总结（最重要）
	// =====================
	fmt.Println("\n盈亏总结:")

	for _, s := range m.Seats {
		if s == nil {
			continue
		}

		win := winMap[s.SeatID]
		net := win - s.TotalBet

		fmt.Printf("用户%d(Seat%d): 投入=%d 赢得=%d 净收益=%d 当前筹码=%d\n",
			s.PlayerID,
			s.SeatID,
			s.TotalBet,
			win,
			net,
			s.Chips,
		)
	}

	fmt.Println("======================")
}
