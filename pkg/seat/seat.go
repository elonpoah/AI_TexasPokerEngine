package seat

import (
	"fmt"
	"poker-engine/pkg/card"
)

type Status int

const (
	Waiting Status = iota
	Active
	Folded
	AllIn
)

type Seat struct {
	SeatID     int         // 座位号
	PlayerID   int         // 玩家唯一ID
	Chips      int         // 当前筹码
	Cards      []card.Card // 玩家手牌
	Status     Status      // 玩家状态: Waiting/Active/Folded/AllIn
	CurrentBet int         // 当前轮下注
	TotalBet   int         // 总下注（整个局）
	Acted      bool        // 本轮是否已经行动过（用于判断轮是否结束）
	Bot        string      // AI策略（如果是AI玩家）deepseek / openai / rule / empty=真人
}

func (a Status) String() string {
	switch a {
	case Waiting:
		return "waiting"
	case Active:
		return "active"
	case Folded:
		return "folded"
	case AllIn:
		return "allin"
	default:
		return "unknown"
	}
}

// -------------------- 玩家方法 --------------------

// Reset 本轮重置（新一轮发牌前调用）
func (s *Seat) ResetForHand() {
	s.Cards = nil
	s.Status = Active
	s.CurrentBet = 0
	s.TotalBet = 0
}

// Fold 弃牌
func (s *Seat) Fold() {
	s.Status = Folded
	s.CurrentBet = 0
}

// CanBet 判断玩家是否可以下注
func (s *Seat) CanBet(amount int) bool {
	return s.Status == Active && amount > 0 && s.Chips > 0
}

// Bet 玩家下注（包含全下逻辑）
// 返回实际下注金额
func (s *Seat) Bet(amount int) int {
	if !s.CanBet(amount) {
		return 0
	}

	// 全下逻辑
	if amount >= s.Chips {
		amount = s.Chips
		s.Chips = 0
		s.Status = AllIn
	} else {
		s.Chips -= amount
	}

	s.CurrentBet += amount
	s.TotalBet += amount
	return amount
}

// Call 跟注，跟桌面当前最大下注
func (s *Seat) Call(maxBet int) int {
	toCall := maxBet - s.CurrentBet
	if toCall <= 0 {
		return 0 // 已经跟注
	}
	return s.Bet(toCall)
}

// Raise 加注（增加到新总下注）
func (s *Seat) Raise(newTotal int) int {
	toRaise := newTotal - s.CurrentBet
	if toRaise <= 0 {
		return 0 // 无效加注
	}
	return s.Bet(toRaise)
}

// Win 玩家赢得筹码
func (s *Seat) Win(amount int) {
	if amount > 0 {
		s.Chips += amount
	}
}

// Lose 玩家输掉筹码（通常只是用于记录统计）
func (s *Seat) Lose(amount int) {
	if amount > 0 && amount <= s.Chips {
		s.Chips -= amount
	}
}

// IsAllIn 判断玩家是否全下
func (s *Seat) IsAllIn() bool {
	return s.Status == AllIn
}

// IsFolded 判断玩家是否弃牌
func (s *Seat) IsFolded() bool {
	return s.Status == Folded
}

// IsActive 判断玩家是否仍在轮中
func (s *Seat) IsActive() bool {
	return s.Status == Active
}

func (s *Seat) String() string {
	return fmt.Sprintf(
		"[Seat:%d Player:%d Chips:%d CurrentBet:%d TotalBet:%d Status:%s Acted:%v Cards:%v]",
		s.SeatID, s.PlayerID, s.Chips, s.CurrentBet, s.TotalBet, s.Status.String(), s.Acted, s.Cards,
	)
}
