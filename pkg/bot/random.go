package bot

import (
	"math/rand"

	"poker-engine/pkg/betting"
)

type RandomBot struct{}

func (b *RandomBot) Name() string {
	return "random"
}

// Act 随机决策（纯娱乐，别当真）
func (b *RandomBot) Act(ctx Context) betting.Action {

	switch rand.Intn(3) {
	case 0:
		return betting.Action{SeatID: ctx.SeatID, Type: betting.Fold}
	case 1:
		return betting.Action{SeatID: ctx.SeatID, Type: betting.Call}
	default:
		return betting.Action{SeatID: ctx.SeatID, Type: betting.AllIn}
	}
}
