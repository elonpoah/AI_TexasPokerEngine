package bot

type Config struct {
	DeepSeekKey string
	OpenAIKey   string
}

func Init(cfg Config) {

	// ⭐ 默认 fallback bot
	SetDefault(&BaseBot{})

	// ⭐ 注册 DeepSeek
	if cfg.DeepSeekKey != "" {
		Register(NewDeepSeekBot(cfg.DeepSeekKey))
	}

	// ⭐ 注册 OpenAI
	if cfg.OpenAIKey != "" {
		Register(NewOpenAIBot(cfg.OpenAIKey))
	}
}

// 使用示例
// b := bot.Get("deepseek")

// ctx := bot.BuildContext(seat, table.Seats, board)

// action := b.Act(ctx)
