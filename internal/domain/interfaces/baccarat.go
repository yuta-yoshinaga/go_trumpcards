package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// BaccaratGame バカラゲームインタフェース
type BaccaratGame interface {
	// Reset ゲームを初期化する
	Reset()
	// Bet ベットしてゲームを実行する
	Bet(amount, betType, ppBet, bpBet int) error
	// ClearHistory 罫線履歴をクリアする
	ClearHistory()

	// GetPlayerHand プレイヤーハンドを取得する
	GetPlayerHand() []*domain.Card
	// GetBankerHand バンカーハンドを取得する
	GetBankerHand() []*domain.Card
	// GetPhase 現在のフェーズを取得する
	GetPhase() int
	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetBetAmount ベット額を取得する
	GetBetAmount() int
	// GetBetType ベットタイプを取得する
	GetBetType() int
	// GetResult ゲーム結果を取得する
	GetResult() domain.GameResult
	// GetPayout 配当金額を取得する
	GetPayout() int
	// GetChips チップを取得する
	GetChips() int
	// GetPlayerHandValue プレイヤーハンド合計値を取得する
	GetPlayerHandValue() int
	// GetBankerHandValue バンカーハンド合計値を取得する
	GetBankerHandValue() int
	// GetActionLog 棋譜を取得する
	GetActionLog() []*domain.ActionLogEntry
	// GetHistory 罫線履歴を取得する
	GetHistory() []int
	// GetPlayerPairBet プレイヤーペアベット額を取得する
	GetPlayerPairBet() int
	// GetBankerPairBet バンカーペアベット額を取得する
	GetBankerPairBet() int
	// GetSideBetResults サイドベット結果を取得する
	GetSideBetResults() []*domain.BacSideBetResult
}
