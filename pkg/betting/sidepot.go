package betting

import (
	"poker-engine/pkg/seat"
	"sort"
)

type Pot struct {
	Amount   int
	Eligible []int // SeatID 改用 slice 避免 map 内存分配
}

func hasAllIn(seats []*seat.Seat) bool {
	for _, s := range seats {
		if s != nil && s.Status == seat.AllIn {
			return true
		}
	}
	return false
}

func Build(seats []*seat.Seat) []Pot {

	if !hasAllIn(seats) {

		total := 0
		for _, s := range seats {
			if s != nil {
				total += s.TotalBet
			}
		}

		// 所有人都参与
		var eligible []int
		for _, s := range seats {
			if s != nil {
				eligible = append(eligible, s.SeatID)
			}
		}

		return []Pot{
			{
				Amount:   total,
				Eligible: eligible,
			},
		}
	}

	// 👇 只有 All-in 才走你原来的 side pot 逻辑
	return buildSidePot(seats)
}

func buildSidePot(seats []*seat.Seat) []Pot {

	type node struct {
		id  int // SeatID
		bet int
	}

	var arr []node

	for _, s := range seats {
		if s != nil && s.TotalBet > 0 {
			arr = append(arr, node{s.SeatID, s.TotalBet})
		}
	}

	sort.Slice(arr, func(i, j int) bool {
		return arr[i].bet < arr[j].bet
	})

	var pots []Pot
	prev := 0

	for i := 0; i < len(arr); i++ {

		if arr[i].bet == prev {
			continue
		}

		level := arr[i].bet - prev

		var eligible []int

		for j := i; j < len(arr); j++ {
			eligible = append(eligible, arr[j].id)
		}

		pots = append(pots, Pot{
			Amount:   level * len(eligible),
			Eligible: eligible,
		})

		prev = arr[i].bet
	}

	return pots
}
