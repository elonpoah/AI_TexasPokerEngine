package main

import (
	"poker-engine/pkg/bot"
	"poker-engine/pkg/player"
	"poker-engine/pkg/room"
	"time"
)

func main() {

	bot.Init(bot.Config{
		DeepSeekKey: "",
	})

	r := room.New(1, 6)
	r.SitDown(&player.Player{ID: 0, Chips: 1000, Bot: "deepseek"})
	r.SitDown(&player.Player{ID: 1, Chips: 1000, Bot: "deepseek"})
	r.SitDown(&player.Player{ID: 2, Chips: 1000, Bot: "deepseek"})
	r.SitDown(&player.Player{ID: 3, Chips: 1000, Bot: "baseai"})
	r.SitDown(&player.Player{ID: 4, Chips: 1000, Bot: "baseai"})
	r.SitDown(&player.Player{ID: 5, Chips: 1000, Bot: "baseai"})
	r.Start(room.BlindConfig{
		SmallBlind: 10,
		BigBlind:   20,
	})

	time.Sleep(time.Second * 1000)

	// r.Leave(2)
}
