package bot

import (
	"fmt"
	"poker-engine/pkg/betting"
	"poker-engine/pkg/card"
	"strings"
)

func buildPrompt(ctx Context) string {
	return fmt.Sprintf(`
		你是一个德州扑克 AI 机器人，段位是职业玩家。
		目前测试阶段，翻牌前只能 fold / call 。翻牌后尽量多一些动作选项，增加多样性和不可预测性。
		请根据当前游戏状态和牌力评估，选择一个合法动作。

		========================
		当前游戏状态
		========================

		阶段: %s
		手牌: %s
		公共牌: %s

		底池: %d
		需要跟注: %d
		当前最高下注: %d
		最小加注额: %d
		你的筹码: %d
		剩余玩家: %d
		牌力评估: %.2f（0~1）

		========================
		合法动作（非常重要）
		========================
		
		你只能从以下动作中选择一个,如果是bet/raise，需要在指定范围内,你可以根据牌力评估给适当的金额，不要一直都是范围的最小值：

		%s

		========================
		输出格式（必须严格遵守）
		========================
		示例：
		fold
		call
		check
		allin
		bet 50
		raise 100

		不需要给出理由。
		`,
		ctx.Stage,
		formatCards(ctx.Hand),
		formatCards(ctx.Board),

		ctx.Pot,
		ctx.ToCall,
		ctx.MaxBet,
		ctx.MinRaise,
		ctx.Stack,
		ctx.ActivePlayers,
		ctx.Strength,

		formatActions(ctx.LegalActions),
	)
}

func formatCards(cards []card.Card) string {

	if len(cards) == 0 {
		return "[]"
	}

	var sb strings.Builder

	for i, c := range cards {
		if i > 0 {
			sb.WriteString(" ")
		}
		sb.WriteString(c.String())
	}

	return sb.String()
}

func formatActions(actions []betting.LegalActionOption) string {
	var sb strings.Builder

	for i, a := range actions {
		if i > 0 {
			sb.WriteString("\n")
		}

		sb.WriteString("- ")
		sb.WriteString(a.Type.String())

		if a.Range != nil {
			sb.WriteString(fmt.Sprintf(" (%d ~ %d)", a.Range.Min, a.Range.Max))
		}
	}
	return sb.String()
}

// ========================
// 	决策规则（仅供参考，不是强制）
// 	========================

// 	- 牌力 > 0.8：偏向 raise / allin
// 	- 牌力 0.5~0.8：偏向 call / check / moderate raise
// 	- 牌力 < 0.5：偏向 fold / check / call

// 	- 剩余玩家多：更保守
// 	- 底池大：降低激进程度
