package evaluator

import (
	"poker-engine/pkg/card"
)

type HandValue struct {
	Level int
	Value int64
}

const (
	HighCard      = 1 + iota // 高牌
	Pair                     // 一对
	TwoPair                  // 两对
	ThreeKind                // 三条
	Straight                 // 顺子
	Flush                    // 同花
	FullHouse                // 葫芦
	FourKind                 // 四条
	StraightFlush            // 同花顺
)

// 6选5组合预计算（性能关键）
var comb6to5 = [6][5]int{
	{0, 1, 2, 3, 4},
	{0, 1, 2, 3, 5},
	{0, 1, 2, 4, 5},
	{0, 1, 3, 4, 5},
	{0, 2, 3, 4, 5},
	{1, 2, 3, 4, 5},
}

// ⚠️ 7选5的组合预计算（性能关键）
var comb7to5 = [21][5]int{
	{0, 1, 2, 3, 4}, {0, 1, 2, 3, 5}, {0, 1, 2, 3, 6},
	{0, 1, 2, 4, 5}, {0, 1, 2, 4, 6}, {0, 1, 2, 5, 6},
	{0, 1, 3, 4, 5}, {0, 1, 3, 4, 6}, {0, 1, 3, 5, 6},
	{0, 1, 4, 5, 6},
	{0, 2, 3, 4, 5}, {0, 2, 3, 4, 6}, {0, 2, 3, 5, 6},
	{0, 2, 4, 5, 6},
	{0, 3, 4, 5, 6},
	{1, 2, 3, 4, 5}, {1, 2, 3, 4, 6}, {1, 2, 3, 5, 6},
	{1, 2, 4, 5, 6},
	{1, 3, 4, 5, 6},
	{2, 3, 4, 5, 6},
}

// Evaluate 评估手牌
func Evaluate(cards []card.Card) HandValue {

	switch len(cards) {

	case 5:
		return evaluate5(cards)

	case 6:
		return evaluate6(cards)

	case 7:
		return evaluate7(cards)

	default:
		// ⭐ 防御（避免崩）
		return HandValue{}
	}
}

// Evaluate 评估手牌
func evaluate6(cards []card.Card) HandValue {

	var best HandValue
	var tmp [5]card.Card

	for i := 0; i < 6; i++ {
		idx := comb6to5[i]

		tmp[0] = cards[idx[0]]
		tmp[1] = cards[idx[1]]
		tmp[2] = cards[idx[2]]
		tmp[3] = cards[idx[3]]
		tmp[4] = cards[idx[4]]

		hv := evaluate5(tmp[:])

		if hv.Level > best.Level ||
			(hv.Level == best.Level && hv.Value > best.Value) {
			best = hv
		}
	}

	return best
}

func evaluate7(cards []card.Card) HandValue {
	var best HandValue

	// ⚠️ 栈上数组（不会逃逸，不会GC）
	var tmp [5]card.Card

	for i := 0; i < 21; i++ {
		idx := comb7to5[i]

		tmp[0] = cards[idx[0]]
		tmp[1] = cards[idx[1]]
		tmp[2] = cards[idx[2]]
		tmp[3] = cards[idx[3]]
		tmp[4] = cards[idx[4]]

		hv := evaluate5(tmp[:])

		// 比较（避免函数调用开销）
		if hv.Level > best.Level ||
			(hv.Level == best.Level && hv.Value > best.Value) {
			best = hv
		}
	}

	return best
}
func evaluate5(cards []card.Card) HandValue {
	count := [13]int{}
	suitCount := [4]int{}

	for _, c := range cards {
		count[c.Value]++
		suitCount[c.Suit]++
	}

	// 是否同花
	flush := false
	for _, c := range suitCount {
		if c == 5 {
			flush = true
			break
		}
	}

	// mask（顺子用）
	mask := 0
	for i := 0; i < 13; i++ {
		if count[i] > 0 {
			mask |= 1 << i
		}
	}

	// 顺子检测
	straightHigh := -1
	for i := 12; i >= 4; i-- {
		if (mask>>(i-4))&0x1F == 0x1F {
			straightHigh = i
			break
		}
	}
	// A2345
	if straightHigh == -1 && (mask&0b1000000001111) == 0b1000000001111 {
		straightHigh = 3
	}

	// 统计
	var four, three int = -1, -1
	pairs := [2]int{-1, -1}
	pairCount := 0
	for i := 12; i >= 0; i-- {
		switch count[i] {
		case 4:
			four = i
		case 3:
			three = i
		case 2:
			if pairCount < 2 {
				pairs[pairCount] = i
			}
			pairCount++
		}
	}

	// 按牌型优先级返回
	switch {
	case flush && straightHigh != -1:
		return HandValue{StraightFlush, int64(straightHigh)}
	case four != -1:
		kicker := getKicker(count, []int{four})
		return HandValue{FourKind, pack(four, []int{kicker})}
	case three != -1 && pairCount > 0:
		return HandValue{FullHouse, pack(three, []int{pairs[0]})}
	case flush:
		return HandValue{Flush, buildHighCardValue(count)}
	case straightHigh != -1:
		return HandValue{Straight, int64(straightHigh)}
	case three != -1:
		kickers := getTopKickers(count, []int{three}, 2)
		return HandValue{ThreeKind, pack(three, kickers)}
	case pairCount >= 2:
		kicker := getKicker(count, pairs[:2])
		return HandValue{TwoPair, pack(pairs[0], []int{pairs[1], kicker})}
	case pairCount == 1:
		kickers := getTopKickers(count, []int{pairs[0]}, 3)
		return HandValue{Pair, pack(pairs[0], kickers)}
	default:
		return HandValue{HighCard, buildHighCardValue(count)}
	}
}

// 将主牌和 kicker 打包成 int64 用于比较大小
func pack(main int, kickers []int) int64 {
	val := int64(main)
	for _, k := range kickers {
		val = (val << 4) | int64(k)
	}
	return val
}

// 返回 count 中排除 exclude 后的最大点数
func getKicker(count [13]int, exclude []int) int {
	ex := [13]bool{}
	for _, e := range exclude {
		if e >= 0 {
			ex[e] = true
		}
	}
	for i := 12; i >= 0; i-- {
		if count[i] > 0 && !ex[i] {
			return i
		}
	}
	return 0
}

// 返回 count 中排除 exclude 后的前 n 张最大牌
func getTopKickers(count [13]int, exclude []int, n int) []int {
	ex := [13]bool{}
	for _, e := range exclude {
		if e >= 0 {
			ex[e] = true
		}
	}
	res := make([]int, 0, n)
	for i := 12; i >= 0 && len(res) < n; i-- {
		if count[i] > 0 && !ex[i] {
			res = append(res, i)
		}
	}
	return res
}

// 构建高牌 int64 值
func buildHighCardValue(count [13]int) int64 {
	val := int64(0)
	for i := 12; i >= 0; i-- {
		for j := 0; j < count[i]; j++ {
			val = (val << 4) | int64(i)
		}
	}
	return val
}
