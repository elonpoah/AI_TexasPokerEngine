package bot

import (
	"poker-engine/pkg/betting"
	"poker-engine/pkg/card"
	"poker-engine/pkg/seat"
)

// BuildContext = 给 AI 的“视野”
func BuildContext(
	s *seat.Seat,
	seats []*seat.Seat,
	board []card.Card,
	r *betting.Round,
	actions []betting.LegalActionOption,
) Context {

	var (
		pot           int
		activePlayers int
	)

	// ---------- 扫描桌面 ----------
	for _, p := range seats {
		if p == nil {
			continue
		}
		// 底池 = 所有人 TotalBet
		pot += p.TotalBet

		// 还在局中的玩家（Active + AllIn）
		if p.Status == seat.Active || p.Status == seat.AllIn {
			activePlayers++
		}
	}

	// ---------- 计算 ToCall ----------
	toCall := max(r.CurrentBet-s.CurrentBet, 0)

	// ---------- 判断阶段 ----------
	stage := getStage(board)

	// ---------- 牌力 ----------
	strength := CalcStrength(s.Cards, board)

	return Context{
		SeatID:   s.SeatID,
		PlayerID: s.PlayerID,

		Hand:  s.Cards,
		Board: board,

		Stack: s.Chips,

		CurrentBet: s.CurrentBet,
		MaxBet:     r.CurrentBet,
		ToCall:     toCall,

		Pot: pot,

		ActivePlayers: activePlayers,
		Stage:         stage,

		Strength: strength,

		MinRaise: r.MinRaise,

		LegalActions: actions,
	}
}
