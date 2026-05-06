package bot

import (
	"poker-engine/pkg/betting"
)

func Decide(ctx Context) betting.Action {

	// ⭐ 强牌
	if ctx.Strength > 0.5 {

		// 无人下注 → Bet
		if ctx.ToCall == 0 {
			target := ctx.Pot / 2
			return normalizeAction(ctx, safeBet(ctx, target))
		}

		if ctx.ToCall >= ctx.Stack {
			return normalizeAction(ctx, betting.Action{
				SeatID: ctx.SeatID,
				Type:   betting.AllIn,
			})
		}

		return normalizeAction(ctx, betting.Action{
			SeatID: ctx.SeatID,
			Type:   betting.Call,
		})
	}

	// ⭐ 中等牌
	if ctx.Strength > 0.2 {

		if ctx.ToCall == 0 {
			return normalizeAction(ctx, betting.Action{
				SeatID: ctx.SeatID,
				Type:   betting.Check,
			})
		}
		odds := potOdds(ctx)

		// 如果胜率高于赔率 → call
		if ctx.Strength > odds && ctx.ToCall < ctx.Stack {
			return normalizeAction(ctx, betting.Action{
				SeatID: ctx.SeatID,
				Type:   betting.Call,
			})
		}

		return normalizeAction(ctx, betting.Action{
			SeatID: ctx.SeatID,
			Type:   betting.Fold,
		})
	}

	// ⭐ 弱牌
	if ShouldBluff(ctx) {
		if ctx.ToCall == 0 {
			target := ctx.Pot / 3
			return normalizeAction(ctx, safeBet(ctx, target))
		}

		target := ctx.MaxBet + ctx.Pot/3
		return normalizeAction(ctx, safeRaise(ctx, target))
	}

	return normalizeAction(ctx, betting.Action{
		SeatID: ctx.SeatID,
		Type:   betting.Fold,
	})
}

func potOdds(ctx Context) float64 {
	return float64(ctx.ToCall) / float64(ctx.Pot+ctx.ToCall)
}

func safeBet(ctx Context, target int) betting.Action {

	// 超过筹码 → allin
	if target >= ctx.Stack {
		return betting.Action{
			SeatID: ctx.SeatID,
			Type:   betting.AllIn,
		}
	}

	return betting.Action{
		SeatID: ctx.SeatID,
		Type:   betting.Bet,
		Amount: target,
	}
}
func safeRaise(ctx Context, target int) betting.Action {

	// 如果超过筹码 → allin
	if target >= ctx.Stack {
		return betting.Action{
			SeatID: ctx.SeatID,
			Type:   betting.AllIn,
		}
	}

	return betting.Action{
		SeatID: ctx.SeatID,
		Type:   betting.Raise,
		Amount: target,
	}
}
func normalizeAction(ctx Context, act betting.Action) betting.Action {

	// 1️⃣ 找到对应合法动作
	var option *betting.LegalActionOption

	for i := range ctx.LegalActions {
		if ctx.LegalActions[i].Type == act.Type {
			option = &ctx.LegalActions[i]
			break
		}
	}

	// ❌ 不合法 → fallback
	if option == nil {
		return fallbackAction(ctx)
	}

	// 2️⃣ 如果不需要金额
	if option.Range == nil {
		act.Amount = 0
		return act
	}

	// 3️⃣ 需要金额（Bet / Raise / AllIn）

	min := option.Range.Min
	max := option.Range.Max

	// clamp
	if act.Amount < min {
		act.Amount = min
	}
	if act.Amount > max {
		act.Amount = max
	}

	return act
}
func fallbackAction(ctx Context) betting.Action {

	// 1️⃣ 优先：不投入筹码的动作（最安全）
	for _, a := range ctx.LegalActions {
		if a.Type == betting.Check {
			return betting.Action{
				SeatID: ctx.SeatID,
				Type:   betting.Check,
			}
		}
	}

	// 2️⃣ 次优：最小成本动作（Call）
	for _, a := range ctx.LegalActions {
		if a.Type == betting.Call {
			return betting.Action{
				SeatID: ctx.SeatID,
				Type:   betting.Call,
			}
		}
	}

	// 3️⃣ 再次：Fold（保守）
	for _, a := range ctx.LegalActions {
		if a.Type == betting.Fold {
			return betting.Action{
				SeatID: ctx.SeatID,
				Type:   betting.Fold,
			}
		}
	}

	// 4️⃣ 再次：最小下注 / 加注
	for _, a := range ctx.LegalActions {
		if a.Range != nil {
			return betting.Action{
				SeatID: ctx.SeatID,
				Type:   a.Type,
				Amount: a.Range.Min,
			}
		}
	}

	// 5️⃣ 最终兜底（理论不会到这里）
	a := ctx.LegalActions[0]

	action := betting.Action{
		SeatID: ctx.SeatID,
		Type:   a.Type,
	}

	if a.Range != nil {
		action.Amount = a.Range.Min
	}

	return action
}
