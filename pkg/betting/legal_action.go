package betting

import "poker-engine/pkg/seat"

type LegalActionOption struct {
	Type ActionType

	// nil = 不需要金额（Check / Fold / Call）
	Range *BetRange
}

type BetRange struct {
	Min int
	Max int
}

type LegalActionsContext struct {
	MaxBet   int // 当前轮最高下注
	ToCall   int // 跟注金额
	Stack    int //	玩家剩余筹码
	MinRaise int // 最小加注额
}

func LegalActions(ctx LegalActionsContext, s *seat.Seat) []LegalActionOption {

	actions := make([]LegalActionOption, 0, 4)

	canBet := ctx.MaxBet == 0
	canRaise := ctx.MaxBet > 0
	isAllInPossible := s.Chips > 0

	// =========================
	// 1️⃣ 无人下注
	// =========================
	if canBet {

		// Check
		actions = append(actions, LegalActionOption{
			Type: Check,
		})

		if s.Chips > 0 {

			// Bet 范围
			minBet := ctx.MinRaise // 一般等于大盲
			if minBet <= 0 {
				minBet = 1
			}

			maxBet := s.Chips

			// ✔ Bet
			if maxBet >= minBet {
				actions = append(actions, LegalActionOption{
					Type: Bet,
					Range: &BetRange{
						Min: minBet,
						Max: maxBet,
					},
				})
			}

			// ✔ All-in（单独保留，UI更清晰）
			actions = append(actions, LegalActionOption{
				Type: AllIn,
				Range: &BetRange{
					Min: s.Chips,
					Max: s.Chips,
				},
			})
		}

		return actions
	}

	// =========================
	// 2️⃣ 已有人下注
	// =========================

	// Fold
	actions = append(actions, LegalActionOption{
		Type: Fold,
	})

	// Call / Check
	if ctx.ToCall > 0 {
		actions = append(actions, LegalActionOption{
			Type: Call,
		})
	} else {
		actions = append(actions, LegalActionOption{
			Type: Check,
		})
	}

	// =========================
	// ✔ Raise（关键修复点）
	// =========================

	if canRaise && s.Chips > ctx.ToCall {

		// 最小 raise 到多少（不是加多少！）
		minRaiseTo := ctx.MaxBet + ctx.MinRaise

		// 玩家最多能 raise 到多少
		maxRaiseTo := s.CurrentBet + s.Chips

		// 👉 必须还能形成“合法加注”
		if maxRaiseTo > minRaiseTo {

			actions = append(actions, LegalActionOption{
				Type: Raise,
				Range: &BetRange{
					Min: minRaiseTo,
					Max: maxRaiseTo,
				},
			})
		}
	}

	// =========================
	// ✔ All-in（永远合法）
	// =========================

	if isAllInPossible {
		allInTo := s.CurrentBet + s.Chips

		actions = append(actions, LegalActionOption{
			Type: AllIn,
			Range: &BetRange{
				Min: allInTo,
				Max: allInTo,
			},
		})
	}

	return actions
}
