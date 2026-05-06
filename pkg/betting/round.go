package betting

import (
	"poker-engine/pkg/seat"
)

type Round struct {
	Seats []*seat.Seat

	CurrentBet int // 当前轮的最高下注
	MinRaise   int // 最小加注额

	Position      int // 当前行动玩家位置
	LastAggressor int // 最后一个加注玩家位置，-1表示没有人加注

	Over bool // 轮是否结束
}

func New(seats []*seat.Seat, start int, bb int) *Round {
	return &Round{
		Seats:         seats,
		Position:      start,
		CurrentBet:    bb,
		MinRaise:      bb,
		LastAggressor: -1,
	}
}

func (r *Round) CurrentPlayer() *seat.Seat {
	return r.Seats[r.Position]
}

func (r *Round) next() {
	n := len(r.Seats)

	for i := 0; i < n; i++ {
		r.Position = (r.Position + 1) % n

		s := r.Seats[r.Position]

		if s != nil && s.Status == seat.Active {
			return
		}
	}
}

func (r *Round) Next() {
	r.next()
}

func (r *Round) Apply(s *seat.Seat, a Action) {

	if r.Over || s != r.CurrentPlayer() {
		return
	}

	switch a.Type {

	case Fold:
		s.Fold()

	case Check:
		if r.CurrentBet > s.CurrentBet {
			return //panic("invalid check")
		}
		s.Acted = true

	case Call:
		s.Call(r.CurrentBet)
		s.Acted = true
	case Bet:
		s.Bet(a.Amount)
		if s.CurrentBet > r.CurrentBet {
			r.CurrentBet = s.CurrentBet
		}
		s.Acted = true
	case Raise:
		// Raise 到目标总下注
		s.Raise(a.Amount)

		r.MinRaise = a.Amount - r.CurrentBet
		r.CurrentBet = a.Amount
		r.LastAggressor = s.SeatID
		// 🚨 raise 之后所有人重新未行动
		for _, p := range r.Seats {
			if p != nil && p.Status == seat.Active {
				p.Acted = false
			}
		}

		s.Acted = true

	case AllIn:
		s.Bet(s.Chips)

		if s.CurrentBet > r.CurrentBet {
			r.CurrentBet = s.CurrentBet
			r.LastAggressor = s.SeatID

			for _, p := range r.Seats {
				if p != nil && p.Status == seat.Active {
					p.Acted = false
				}
			}
		}
		s.Acted = true
	}

	r.next()
	r.checkOver()
}

func (r *Round) ResetRound() {
	for _, s := range r.Seats {
		if s == nil {
			continue
		}

		if s.Status == seat.Active {
			s.Acted = false
			s.CurrentBet = 0
		}
	}

	r.CurrentBet = 0
	r.MinRaise = 0
	r.LastAggressor = -1
}

func (r *Round) checkOver() {

	activeCount := 0
	inHandCount := 0

	for _, s := range r.Seats {
		if s == nil {
			continue
		}

		// ❌ Folded 不参与
		if s.Status == seat.Folded {
			continue
		}

		// ✔ 仍在牌局
		inHandCount++

		// ✔ 还能行动
		if s.Status == seat.Active {
			activeCount++
		}
	}

	// 🚨 1. 只剩 1 个“在牌局的人” → 直接结束
	if inHandCount <= 1 {
		r.Over = true
		return
	}

	// 🚨 2. 如果没有人还能行动（全 AllIn 或 Fold）
	if activeCount == 0 {
		r.Over = true
		return
	}

	// 3️⃣ 所有人已行动 + 下注对齐
	for _, s := range r.Seats {
		if s == nil || s.Status != seat.Active {
			continue
		}

		if !s.Acted {
			return
		}

		if s.CurrentBet != r.CurrentBet {
			return
		}
	}

	// ✔ 所有人对齐 → 结束本轮
	r.Over = true
}
