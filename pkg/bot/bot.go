package bot

import "poker-engine/pkg/betting"

type Bot interface {
	Name() string
	Act(ctx Context) betting.Action
}
