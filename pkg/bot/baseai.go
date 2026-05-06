package bot

import "poker-engine/pkg/betting"

type BaseBot struct{}

func (r *BaseBot) Name() string {
	return "baseai"
}

func (r *BaseBot) Act(ctx Context) betting.Action {
	return Decide(ctx)
}
