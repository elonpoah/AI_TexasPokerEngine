## Poker AI Lab PRO

#### Texas hold'em Engine, AI Engine, SDK, If you need this engine, leave your issues.
#### 德州扑克,AI机器人，SDK， 如果你需要这个引擎，留下你的issues


- ✔ 可训练（Self-play + RL ready）
- ✔ 可对战（Human / AI / Bot混合）
-  ✔ 可回放（Event-sourcing）
- ✔ 可扩展策略（CFR / PPO / Heuristic）
- ✔ 可横向扩展（多房间/多进程）
- ✔ 可观测（metrics + log + replay）


### 🧠 一、PRO级架构（核心思想）

```md
                 ┌──────────────────────┐
                 │   Match Orchestrator │
                 └─────────┬────────────┘
                           │
        ┌──────────────────┼──────────────────┐
        ▼                  ▼                  ▼
 ┌─────────────┐  ┌─────────────┐  ┌─────────────┐
 │ Room Engine │  │ Replay Log  │  │ AI Trainer  │
 └──────┬──────┘  └──────┬──────┘  └──────┬──────┘
        │               │               │
        ▼               ▼               ▼
   Game State      Event Stream     Policy Update
        │
        ▼
   Agents (CFR / PPO / Bot / Human)
```

### 🧩 二、PRO工程结构

```
poker-ai-lab/
├── cmd
│   └── main.go
├── configs
│   └── config.yaml
├── go.mod
├── internal
│   ├── model
│   ├── repository
│   └── service
├── main.go
├── pkg
│   ├── betting
│   │   ├── action.go
│   │   ├── legal_action.go
│   │   ├── round.go
│   │   └── sidepot.go
│   ├── bot
│   │   ├── baseai.go
│   │   ├── bootstrap.go
│   │   ├── bot.go
│   │   ├── context.go
│   │   ├── context_builder.go
│   │   ├── decision.go
│   │   ├── deepseek.go
│   │   ├── openai.go
│   │   ├── prompt.go
│   │   ├── random.go
│   │   ├── registry.go
│   │   ├── stage.go
│   │   ├── strategy.go
│   │   └── strength.go
│   ├── card
│   │   └── card.go
│   ├── deck
│   │   └── deck.go
│   ├── engine
│   │   ├── engine.go
│   │   └── payout.go
│   ├── evaluator
│   │   └── evaluator.go
│   ├── mode
│   │   └── actions.go
│   ├── player
│   │   └── player.go
│   ├── room
│   │   └── room.go
│   ├── seat
│   │   └── seat.go
│   ├── table
│       └── table.go
│   
│
├── README.md
│
└── test
    ├── common.go
    ├── mock_handler.go
    ├── payout_sidepot_test.go
    ├── payout_simple_test.go
    └── payout_split_test.go
```

### 🧠 三、PRO核心升级点

#### 1️⃣ State（从“数据”升级为“特征向量”）
```go
type State struct {
	SeatID int

	// 🃏 手牌
	HoleCards []card.Card

	// 🪑 公共信息
	Board     []card.Card
	Pot       int
	ToCall    int
	Position  int

	// 👥 其他玩家（关键 PRO）
	Players []PlayerView

	// 📊 特征工程（PRO）
	StackRatio float64
	PotOdds    float64
	Street     int // preflop/flop/turn/river
}
```
#### 2️⃣ Action（标准化）
```go
type Action struct {
	Type   ActionType
	Amount int
}
```
#### 3️⃣ Event Stream
```go
type Event struct {
	Type      string
	SeatID    int
	PlayerID  int
	Action    Action
	State     State
	Board     []card.Card
	Timestamp int64
}
```
#### 4️⃣ Agent
```go
type Agent interface {
	ID() int

	// 决策
	Act(State) Action

	// 学习（RL用）
	Learn(traj Trajectory)

	// 可选：模型导出
	Model() []byte
}
```
### 🧠 四、PRO训练系统（Self-play）
#### Trainer（核心）
```go
type Trainer struct {
	Env *env.PokerEnv

	Agents []agent.Agent

	Buffer *rl.Buffer
}
```
#### Self-play循环
```go
func (t *Trainer) Run() {

	for episode := 0; episode < 100000; episode++ {

		t.Env.Reset()

		var trajectory []rl.Step

		for !t.Env.Done() {

			for _, ag := range t.Agents {

				state := t.Env.State(ag.ID())

				action := ag.Act(state)

				next, reward, done := t.Env.Step(ag.ID(), action)

				trajectory = append(trajectory, rl.Step{
					State:  state,
					Action: action,
					Reward: reward,
				})

				if done {
					break
				}
			}
		}

		t.Buffer.Push(trajectory)

		// 🔥 PPO / CFR 更新
		t.UpdatePolicy()
	}
}
```
### 🧠 五、PPO训练
```go
type PPOTrainer struct {
	Policy Network
	Value  Network
}
func (p *PPOTrainer) Update(batch []Trajectory) {

	// policy loss
	// value loss
	// advantage estimate

	// PPO clip objective
}
```

### 📼 六、Replay系统
#### Event Store（不可变日志）
```go
type EventStore interface {
	Save(gameID string, events []Event)
	Load(gameID string) []Event
}
```
#### 文件实现
```go
type FileStore struct{}

func (f *FileStore) Save(id string, e []Event) {
	data, _ := json.Marshal(e)
	os.WriteFile("replay_"+id+".json", data, 0644)
}
```
#### 回放播放器
```go
func Play(events []Event) {

	for _, e := range events {

		fmt.Printf("[%s] Seat%d Action=%v Pot=%d\n",
			e.Type,
			e.SeatID,
			e.Action,
			e.State.Pot,
		)

		time.Sleep(300 * time.Millisecond)
	}
}
```
### 🧠 七、Match系统
```
✔ AI vs AI
✔ Human vs AI
✔ 多房间并行
```
```go
type Match struct {
	Rooms []*room.Room
}
func (m *Match) Run() {

	for _, r := range m.Rooms {
		go r.Start()
	}
}
```
### 📊 八、观测系统
```go
type Metrics struct {
	WinRate map[int]float64
	Epoch   int
}
// AI1 winrate: 52%
// AI2 winrate: 48%
```
### 🚀 九、总结

训练能力
- ✔ Self-play 10万局
- ✔ PPO / CFR ready
- ✔ replay buffer

对战能力
- ✔ AI vs AI
- ✔ AI vs Human
- ✔ 多策略混合

✔ 回放能力
- ✔ 完整牌局事件流
- ✔ 可视化 replay
- ✔ 可debug每一步决策

✔ 工业能力
- ✔ 多房间并发
- ✔ 可扩展存储
- ✔ 可接 WebSocket
- ✔ 可接前端 UI


### 输出结果

```

{Type:action Data:{SeatID:0 PlayerID:0 Type:call Amount:0}}
{Type:action Data:{SeatID:1 PlayerID:1 Type:fold Amount:0}}
{Type:action Data:{SeatID:2 PlayerID:2 Type:call Amount:0}}
{Type:action Data:{SeatID:3 PlayerID:3 Type:check Amount:0}}
{Type:round_end Data:{}}
{Type:action Data:{SeatID:2 PlayerID:2 Type:check Amount:0}}
{Type:action Data:{SeatID:3 PlayerID:3 Type:check Amount:0}}
{Type:action Data:{SeatID:0 PlayerID:0 Type:check Amount:0}}
{Type:round_end Data:{}}
{Type:action Data:{SeatID:2 PlayerID:2 Type:check Amount:0}}
{Type:action Data:{SeatID:3 PlayerID:3 Type:bet Amount:20}}
{Type:action Data:{SeatID:0 PlayerID:0 Type:fold Amount:0}}
{Type:action Data:{SeatID:2 PlayerID:2 Type:fold Amount:0}}
{Type:round_end Data:{}}
{Type:payout Data:[{PotIndex:0 Amount:80 Players:[{SeatID:3 PlayerID:3 Win:80}]}]}
========== ROOM STATE ==========
RoomID: 1 Status: waiting
----- TABLE -----
Button: 1 SB: 2 BB: 3
Blinds: SB=10 BB=20
----- BOARD -----
Board: [2♠ 4♥ 3♠ J♥]
----- SEATS -----
[Seat:0 Player:0 Chips:980 CurrentBet:0 TotalBet:20 Status:folded Acted:false Cards:[Q♣ J♠]]
[Seat:1 Player:1 Chips:1000 CurrentBet:0 TotalBet:0 Status:folded Acted:false Cards:[4♠ 6♣]]
[Seat:2 Player:2 Chips:980 CurrentBet:0 TotalBet:20 Status:folded Acted:true Cards:[7♦ J♣]]
[Seat:3 Player:3 Chips:1040 CurrentBet:0 TotalBet:40 Status:active Acted:false Cards:[J♦ 9♦]]
[Seat:4 Player:4 Chips:1000 CurrentBet:0 TotalBet:0 Status:folded Acted:false Cards:[6♠ 10♦]]
[Seat:5 Player:5 Chips:1000 CurrentBet:0 TotalBet:0 Status:folded Acted:false Cards:[3♣ 4♣]]
================================
```