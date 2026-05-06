package player

type Player struct {
	ID    int
	Chips int
	// ⭐ AI策略
	Bot string // openai / deepseek / rule / empty=真人
}
