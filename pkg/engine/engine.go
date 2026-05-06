package engine

import (
	"time"

	"poker-engine/pkg/betting"
	"poker-engine/pkg/bot"
	"poker-engine/pkg/card"
	"poker-engine/pkg/deck"
	"poker-engine/pkg/seat"
	"poker-engine/pkg/table"
)

type EventHandler interface {
	OnDeal(userID int, seatID int, cards []card.Card)
	OnBoard(cards []card.Card)
	OnAction(userID int, seatID int, action betting.Action)
	OnTurn(pid, seatID, seconds int)
	OnRoundEnd()
	OnPayout(results []PayoutResult)
}

type Engine struct {
	Table *table.Table
	Deck  *deck.Deck
	Board []card.Card

	ActionChan chan betting.Action // 用于接收玩家动作的通道，容量足够大以避免阻塞
	Handler    EventHandler        // 事件处理器接口
}

func New(t *table.Table, h EventHandler) *Engine {
	return &Engine{
		Table:      t,
		Deck:       deck.New(),
		ActionChan: make(chan betting.Action, 32),
		Handler:    h,
	}
}

func (e *Engine) reset() {
	e.Board = nil
	e.Deck.Shuffle()

	for _, s := range e.Table.Seats {
		if s == nil {
			continue
		}
		s.ResetForHand()
	}
}

func (e *Engine) deal() {
	for i := 0; i < 2; i++ {
		for _, s := range e.Table.Seats {
			if s != nil && s.Status == seat.Active {

				card := e.Deck.Deal()
				s.Cards = append(s.Cards, card)

				if e.Handler != nil {
					e.Handler.OnDeal(s.PlayerID, s.SeatID, s.Cards)
				}
			}
		}
	}
}

func (e *Engine) emitBoard() {
	if e.Handler != nil {
		e.Handler.OnBoard(e.Board)
	}
}

func (e *Engine) flop() {
	e.Board = append(e.Board, e.Deck.Deal(), e.Deck.Deal(), e.Deck.Deal())
	e.emitBoard()
}

func (e *Engine) turn() {
	e.Board = append(e.Board, e.Deck.Deal())
	e.emitBoard()
}

func (e *Engine) river() {
	e.Board = append(e.Board, e.Deck.Deal())
	e.emitBoard()
}

func (e *Engine) runRound(start int, isPreflop bool) {

	r := betting.New(e.Table.Seats, start, e.Table.BigBlind)

	if !isPreflop {
		r.CurrentBet = 0
	}

	for !r.Over {

		s := r.CurrentPlayer()
		// 正常不会为 nil，如果出现说明逻辑有问题
		if s == nil {
			panic("invalid state: current player is nil")
		}
		// fmt.Println("Current player:", s)

		// ⭐1. 广播：轮到谁
		e.Handler.OnTurn(s.PlayerID, s.SeatID, 10)
		// ⭐2. 计算合法动作
		lgActions := betting.LegalActions(betting.LegalActionsContext{
			MaxBet:   r.CurrentBet,
			ToCall:   r.CurrentBet - s.CurrentBet,
			Stack:    s.Chips,
			MinRaise: r.MinRaise,
		}, s)

		var action betting.Action
		switch {
		case s.Bot != "":
			ctx := bot.BuildContext(s, e.Table.Seats, e.Board, r, lgActions)
			// fmt.Println("ctx:", ctx)
			b := bot.Get(s.Bot)

			action = b.Act(ctx)

		case true:
			// 非机器人
			select {
			case action = <-e.ActionChan:

				if action.SeatID != s.SeatID {
					continue
				}

			case <-time.After(10 * time.Second):
				action = betting.Action{
					SeatID: s.SeatID,
					Type:   betting.Fold,
				}
			}
		}
		r.Apply(s, action)

		if e.Handler != nil {
			e.Handler.OnAction(s.PlayerID, s.SeatID, action)
		}
	}

	if e.Handler != nil {
		// ✅ 重置轮状态，准备下一轮
		r.ResetRound()
		// 广播一轮结束
		e.Handler.OnRoundEnd()
	}
}

// 还在手牌中的（没Fold）
func (e *Engine) inHandCount() int {
	count := 0
	for _, s := range e.Table.Seats {
		if s != nil && s.Status != seat.Folded {
			count++
		}
	}
	return count
}

// 能行动的
func (e *Engine) canActCount() int {
	count := 0
	for _, s := range e.Table.Seats {
		if s != nil && s.Status == seat.Active {
			count++
		}
	}
	return count
}

// 是否结束整手牌
func (e *Engine) isHandOver() bool {
	return e.inHandCount() <= 1
}

// 是否跳过 betting（全 all-in）
func (e *Engine) shouldSkipBetting() bool {
	return e.canActCount() == 0 && e.inHandCount() > 1
}

func (e *Engine) PlayHand() {

	e.Table.MoveButton()
	e.reset()

	e.Table.PostBlinds()

	// ================= PRE FLOP =================

	e.deal()
	e.runRound(e.Table.PreFlopStart(), true)

	if e.isHandOver() {
		e.Payout()
		return
	}

	if e.shouldSkipBetting() {
		goto runout
	}

	// ================= FLOP =================
	e.flop()
	e.runRound(e.Table.PostFlopStart(), false)

	if e.isHandOver() {
		e.Payout()
		return
	}

	if e.shouldSkipBetting() {
		goto runout
	}

	// ================= TURN =================
	e.turn()
	e.runRound(e.Table.PostFlopStart(), false)

	if e.isHandOver() {
		e.Payout()
		return
	}

	if e.shouldSkipBetting() {
		goto runout
	}

	// ================= RIVER =================
	e.river()
	e.runRound(e.Table.PostFlopStart(), false)

runout:
	// ✔ All-in 情况：直接发完剩余公共牌
	for len(e.Board) < 5 {
		e.Board = append(e.Board, e.Deck.Deal())
		e.emitBoard()
	}

	e.Payout()
}
