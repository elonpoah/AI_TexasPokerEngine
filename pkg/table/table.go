package table

import (
	"poker-engine/pkg/seat"
)

type Table struct {
	Seats []*seat.Seat

	Button int // 庄家位置
	SB     int // 小盲位置
	BB     int // 大盲位置

	SmallBlind int // 小盲注额
	BigBlind   int // 大盲注额
}

func New(seats []*seat.Seat, sb, bb int) *Table {
	return &Table{
		Seats:      seats,
		SmallBlind: sb,
		BigBlind:   bb,
	}
}

func (t *Table) nextActiveSeat(start int) int {
	n := len(t.Seats)

	for i := 1; i <= n; i++ {
		idx := (start + i) % n
		s := t.Seats[idx]

		if s == nil {
			continue
		}

		// ❗必须过滤状态
		if s.Status == seat.Folded || s.Status == seat.AllIn {
			continue
		}

		return idx
	}

	return -1
}
func (t *Table) MoveButton() {
	n := len(t.Seats)
	if n == 0 {
		return
	}

	// 找第一个活着的 button
	t.Button = t.nextActiveSeat(t.Button)
	if t.Button == -1 {
		return
	}

	t.SB = t.nextActiveSeat(t.Button)
	if t.SB == -1 {
		return
	}

	t.BB = t.nextActiveSeat(t.SB)
}

func (t *Table) PostBlinds() {
	sb := t.Seats[t.SB]
	bb := t.Seats[t.BB]

	if sb != nil {
		amount := min(sb.Chips, t.SmallBlind)
		sb.Chips -= amount
		sb.CurrentBet += amount
		sb.TotalBet += amount

		if sb.Chips == 0 {
			sb.Status = seat.AllIn
		}
	}

	if bb != nil {
		amount := min(bb.Chips, t.BigBlind)
		bb.Chips -= amount
		bb.CurrentBet += amount
		bb.TotalBet += amount

		if bb.Chips == 0 {
			bb.Status = seat.AllIn
		}
	}
}

func (t *Table) PreFlopStart() int {
	return t.nextActiveSeat(t.BB)
}

func (t *Table) PostFlopStart() int {
	return t.nextActiveSeat(t.Button)
}
