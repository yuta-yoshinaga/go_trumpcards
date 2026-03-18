package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// BaccaratGame バカラゲームインタフェース
type BaccaratGame interface {
	// interactor が呼ぶメソッド
	Reset()
	Bet(amount, betType, ppBet, bpBet int) error
	ClearHistory()

	// presenter が呼ぶメソッド
	GetPlayerHand() []*domain.Card
	GetBankerHand() []*domain.Card
	GetPhase() int
	GetGameEndFlag() bool
	GetBetAmount() int
	GetBetType() int
	GetResult() domain.GameResult
	GetPayout() int
	GetChips() int
	GetPlayerHandValue() int
	GetBankerHandValue() int
	GetActionLog() []*domain.ActionLogEntry
	GetHistory() []int
	GetPlayerPairBet() int
	GetBankerPairBet() int
	GetSideBetResults() []*domain.BacSideBetResult
}
