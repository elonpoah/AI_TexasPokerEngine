package betting

type ActionType int

const (
	Fold  ActionType = iota // 弃牌
	Call                    // 跟注
	Check                   // 过牌
	Bet                     // 下注
	Raise                   // 加注
	AllIn                   // 全下
)

type Action struct {
	SeatID int
	Type   ActionType
	Amount int
}

func (a ActionType) String() string {
	switch a {
	case Fold:
		return "fold"
	case Call:
		return "call"
	case Check:
		return "check"
	case Bet:
		return "bet"
	case Raise:
		return "raise"
	case AllIn:
		return "allin"
	default:
		return "unknown"
	}
}
