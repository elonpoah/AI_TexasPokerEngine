package bot

import (
	"math/rand"
)

func ShouldBluff(ctx Context) bool {

	// 基础诈唬概率
	base := 0.1

	// 人少 → 更容易诈唬
	if ctx.ActivePlayers <= 2 {
		base += 0.15
	}

	// 无需跟注 → 更容易诈唬
	if ctx.ToCall == 0 {
		base += 0.1
	}

	// preflop 少诈唬
	if ctx.Stage == "preflop" {
		base -= 0.05
	}

	return rand.Float64() < base
}
