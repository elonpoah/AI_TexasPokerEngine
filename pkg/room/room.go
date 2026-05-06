package room

import (
	"fmt"
	"sync"
	"time"

	"poker-engine/pkg/betting"
	"poker-engine/pkg/card"
	"poker-engine/pkg/engine"
	"poker-engine/pkg/player"
	"poker-engine/pkg/seat"
	"poker-engine/pkg/table"
)

type GameEvent struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}
type TurnEvent struct {
	PlayerID int `json:"playerId"`
	SeatID   int `json:"seatId"`
	Timeout  int `json:"seconds"` // 秒
}
type DealEvent struct {
	SeatID   int         `json:"seatId"`
	PlayerID int         `json:"playerId"`
	Cards    []card.Card `json:"cards"`
}

type BoardEvent struct {
	Cards []card.Card `json:"cards"`
}

type ActionEvent struct {
	SeatID   int                `json:"seatId"`
	PlayerID int                `json:"playerId"`
	Type     betting.ActionType `json:"type"`
	Amount   int                `json:"amount"`
}

type RoundEndEvent struct{}

type PayoutPlayer struct {
	SeatID   int `json:"seatId"`
	PlayerID int `json:"playerId"`
	Win      int `json:"win"`
}

type PayoutEvent struct {
	PotIndex int            `json:"potIndex"`
	Amount   int            `json:"amount"`
	Players  []PayoutPlayer `json:"players"`
}

type BlindConfig struct {
	SmallBlind int
	BigBlind   int
}

type RoomStatus int

const (
	StatusWaiting RoomStatus = iota
	StatusPlaying
	StatusPaused
	StatusStopped
)

type Room struct {
	ID       int
	Seats    []*seat.Seat
	MaxSeats int

	Players map[int]*player.Player

	Table  *table.Table
	Engine *engine.Engine

	Status RoomStatus
	mu     sync.Mutex
}

func (a RoomStatus) String() string {
	switch a {
	case StatusWaiting:
		return "waiting"
	case StatusPlaying:
		return "playing"
	case StatusPaused:
		return "paused"
	case StatusStopped:
		return "stopped"
	default:
		return "unknown"
	}
}

func New(id int, max int) *Room {
	return &Room{
		ID:       id,
		MaxSeats: max,
		Seats:    make([]*seat.Seat, max),
		Players:  make(map[int]*player.Player),
		Status:   StatusWaiting,
	}
}

func DefaultConfig() BlindConfig {
	return BlindConfig{
		SmallBlind: 1,
		BigBlind:   2,
	}
}

func (r *Room) SitDown(p *player.Player) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i := 0; i < r.MaxSeats; i++ {
		if r.Seats[i] == nil {
			r.Seats[i] = &seat.Seat{
				SeatID:   i,
				PlayerID: p.ID,
				Chips:    p.Chips,
				Status:   seat.Active,
				Bot:      p.Bot,
			}
			r.Players[p.ID] = p
			return nil
		}
	}
	return fmt.Errorf("no seat")
}

func (r *Room) Leave(pid int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, s := range r.Seats {
		if s != nil && s.PlayerID == pid {
			r.Seats[i] = nil
			delete(r.Players, pid)
			fmt.Println("玩家离开:", pid)
			return
		}
	}
}
func (r *Room) Start(cfg ...BlindConfig) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.Status == StatusPlaying {
		return
	}

	c := DefaultConfig()
	if len(cfg) > 0 {
		c = cfg[0]
	}

	r.Table = table.New(r.Seats, c.SmallBlind, c.BigBlind)

	// ✅ 把 Room 作为事件处理器
	r.Engine = engine.New(r.Table, r)

	r.Status = StatusPlaying

	go r.loop()
}

func (r *Room) Pause() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.Status == StatusPlaying {
		r.Status = StatusPaused
	}
}

func (r *Room) Resume() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.Status == StatusPaused || r.Status == StatusWaiting {
		r.Status = StatusPlaying
	}
}

func (r *Room) loop() {
	for {
		r.mu.Lock()
		status := r.Status
		r.mu.Unlock()

		switch status {

		case StatusStopped:
			return

		case StatusPaused, StatusWaiting:
			time.Sleep(time.Second)
			continue

		case StatusPlaying:
			r.Engine.PlayHand()
		}
	}
}

func (r *Room) DoAction(a betting.Action) {

	r.mu.Lock()
	engine := r.Engine
	r.mu.Unlock()

	if engine == nil {
		return
	}

	// ✅ 基本校验
	if a.SeatID < 0 || a.SeatID >= len(r.Seats) {
		return
	}

	select {
	case r.Engine.ActionChan <- a:
	default:
		fmt.Println("action channel full")
	}
}

func (r *Room) emit(event GameEvent) {
	// TODO: 这里替换成 WebSocket 广播
	// 比如: r.hub.Broadcast(event)

	// 临时调试可以打开：
	fmt.Printf("%+v\n", event)
}

func (r *Room) OnDeal(userID int, seatID int, cards []card.Card) {
	// r.emit(GameEvent{
	// 	Type: "deal",
	// 	Data: DealEvent{
	// 		SeatID:   seatID,
	// 		PlayerID: userID,
	// 		Cards:    cards,
	// 	},
	// })
}

func (r *Room) OnBoard(cards []card.Card) {

	// r.emit(GameEvent{
	// 	Type: "board",
	// 	Data: BoardEvent{
	// 		Cards: cards,
	// 	},
	// })
}

func (r *Room) OnAction(userID int, seatID int, action betting.Action) {

	r.emit(GameEvent{
		Type: "action",
		Data: ActionEvent{
			SeatID:   seatID,
			PlayerID: userID,
			Type:     action.Type,
			Amount:   action.Amount,
		},
	})
}

func (r *Room) OnTurn(pid, seatID, seconds int) {
	// r.emit(GameEvent{
	// 	Type: "turn",
	// 	Data: TurnEvent{PlayerID: pid, SeatID: seatID, Timeout: seconds},
	// })
}

func (r *Room) OnRoundEnd() {
	r.emit(GameEvent{
		Type: "round_end",
		Data: RoundEndEvent{},
	})
}

func (r *Room) OnPayout(results []engine.PayoutResult) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var events []PayoutEvent

	for _, p := range results {

		var players []PayoutPlayer

		for i, sid := range p.Winners {

			s := r.Seats[sid]
			if s == nil {
				continue
			}

			win := p.Share
			if i == 0 {
				win += p.Remainder
			}

			players = append(players, PayoutPlayer{
				SeatID:   sid,
				PlayerID: s.PlayerID,
				Win:      win,
			})
		}

		events = append(events, PayoutEvent{
			PotIndex: p.PotIndex,
			Amount:   p.Amount,
			Players:  players,
		})
	}

	r.emit(GameEvent{
		Type: "payout",
		Data: events,
	})
	// 一局结束
	r.Status = StatusWaiting
	// 调试输出当前状态
	r.dumpState()
}

func (r *Room) dumpState() {

	fmt.Println("========== ROOM STATE ==========")

	fmt.Printf("RoomID: %d Status: %s\n", r.ID, r.Status.String())

	fmt.Println("----- TABLE -----")
	fmt.Printf("Button: %d SB: %d BB: %d\n",
		r.Table.Button, r.Table.SB, r.Table.BB)
	fmt.Printf("Blinds: SB=%d BB=%d\n",
		r.Table.SmallBlind, r.Table.BigBlind)
	fmt.Println("----- BOARD -----")
	if r.Engine != nil {
		fmt.Printf("Board: %v\n", r.Engine.Board)
	}

	fmt.Println("----- SEATS -----")

	for _, s := range r.Seats {
		if s == nil {
			continue
		}

		fmt.Println(s.String())
	}

	fmt.Println("================================")
}
