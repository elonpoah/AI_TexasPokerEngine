package deck

import (
	"math/rand"
	"time"

	"poker-engine/pkg/card"
)

type Deck struct {
	Cards [52]card.Card // 直接使用数组，避免切片分配
	Index int           // 当前发牌位置
}

func New() *Deck {
	d := &Deck{}

	idx := 0
	for i := range 4 {
		for j := range 13 {
			d.Cards[idx] = card.Card{Suit: i, Value: j}
			idx++
		}
	}

	d.Shuffle()
	return d
}

// ⚠️ 全局随机数生成器（线程安全，性能较好）
var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

func (d *Deck) Shuffle() {
	rng.Shuffle(len(d.Cards), func(i, j int) {
		d.Cards[i], d.Cards[j] = d.Cards[j], d.Cards[i]
	})
	d.Index = 0
}

func (d *Deck) Deal() card.Card {
	if d.Index >= len(d.Cards) {
		panic("no cards left")
	}
	c := d.Cards[d.Index]
	d.Index++
	return c
}
