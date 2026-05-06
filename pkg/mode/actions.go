package mode

import "poker-engine/pkg/betting"

type ActionOption struct {
	Type betting.ActionType

	// nil = 不需要金额（Check / Fold / Call）
	Range *BetRange
}

type BetRange struct {
	Min int
	Max int
}
