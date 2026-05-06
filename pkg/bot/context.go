package bot

import (
	"poker-engine/pkg/betting"
	"poker-engine/pkg/card"
)

type Context struct {
	SeatID        int                         // 玩家座位ID
	PlayerID      int                         // 玩家ID
	Hand          []card.Card                 // 手牌
	Board         []card.Card                 // 公牌
	Stack         int                         // 剩余筹码
	CurrentBet    int                         // 下注金额（已经投入的筹码）
	MaxBet        int                         // 当前轮的最高下注
	ToCall        int                         // 跟注金额（还需要投入的筹码）
	Pot           int                         // 底池
	ActivePlayers int                         // 可选增强（推荐）
	Stage         string                      // preflop / flop / turn / river
	Strength      float64                     // 0 ~ 1 牌力
	MinRaise      int                         // 最小加注额
	LegalActions  []betting.LegalActionOption // 可执行的动作列表
}
