package engine

import (
	"poker-engine/pkg/betting"
	"poker-engine/pkg/card"
	"poker-engine/pkg/evaluator"
	"poker-engine/pkg/seat"
)

type PayoutResult struct {
	PotIndex  int   // 第几个池
	Amount    int   // 池金额
	Winners   []int // SeatID
	Share     int   // 每人分多少
	Remainder int   // 余数
}

func (e *Engine) Payout() {

	pots := betting.Build(e.Table.Seats)

	results := make(map[int]evaluator.HandValue)

	for _, s := range e.Table.Seats {
		if s == nil || s.Status == seat.Folded {
			continue
		}

		all := append([]card.Card{}, s.Cards...)
		all = append(all, e.Board...)

		results[s.SeatID] = evaluator.Evaluate(all)

	}

	var payoutResults []PayoutResult

	for i, pot := range pots {

		var winners []int
		var best evaluator.HandValue

		for _, sid := range pot.Eligible {

			val, ok := results[sid]
			if !ok {
				continue
			}

			if len(winners) == 0 || compare(val, best) > 0 {
				winners = []int{sid}
				best = val
			} else if compare(val, best) == 0 {
				winners = append(winners, sid)
			}
		}

		if len(winners) == 0 {
			continue
		}

		share := pot.Amount / len(winners)
		remainder := pot.Amount % len(winners)

		// ✅ 3. 分配筹码
		for i, sid := range winners {

			s := e.Table.Seats[sid]
			if s == nil {
				continue
			}

			win := share
			if i == 0 {
				win += remainder
			}

			s.Chips += win
		}

		// ✅ 4. 收集结果
		payoutResults = append(payoutResults, PayoutResult{
			PotIndex:  i,
			Amount:    pot.Amount,
			Winners:   winners,
			Share:     share,
			Remainder: remainder,
		})
	}

	// ✅ 5. 统一抛出（关键）
	if e.Handler != nil {
		e.Handler.OnPayout(payoutResults)
	}
}

func compare(a, b evaluator.HandValue) int {
	if a.Level != b.Level {
		return a.Level - b.Level
	}
	if a.Value > b.Value {
		return 1
	}
	if a.Value < b.Value {
		return -1
	}
	return 0
}
