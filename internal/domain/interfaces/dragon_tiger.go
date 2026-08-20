//go:build !js || !wasm || extra4

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// DragonTigerGame ドラゴンタイガーゲームインタフェース
type DragonTigerGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// ClearHistory 罫線履歴をクリアする
	ClearHistory()
	// Bet ベットしカードを配り、結果まで自動進行する
	Bet(amount, betType int) error

	// GetDragonCard ドラゴン枠のカード
	GetDragonCard() *domain.Card
	// GetTigerCard タイガー枠のカード
	GetTigerCard() *domain.Card
	// GetPhase 現在のフェーズ
	GetPhase() int
	// GetGameEndFlag ゲーム終了フラグ
	GetGameEndFlag() bool
	// GetBetAmount ベット額
	GetBetAmount() int
	// GetBetType ベットタイプ
	GetBetType() int
	// GetResult 勝敗結果
	GetResult() domain.GameResult
	// GetPayout 配当金額
	GetPayout() int
	// GetChips チップ
	GetChips() int
	// GetHistory 罫線履歴
	GetHistory() []int
}
