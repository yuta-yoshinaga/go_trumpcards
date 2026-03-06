package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// calcMaxBetAmount ベッティングリミットに応じた最大ベット額を計算
func calcMaxBetAmount(limit domain.BettingLimitType, pot, lastBet int) int {
	switch limit {
	case domain.BettingLimitPotLimit:
		return pot + lastBet
	default:
		return 0
	}
}
