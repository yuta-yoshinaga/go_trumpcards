package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// VideoPokerGame ビデオポーカーゲームインタフェース
type VideoPokerGame interface {
	// Reset ゲームを初期化する
	Reset()
	// Bet ベットしてディールする
	Bet(amount int) error
	// Hold ホールド選択＆ドロー
	Hold(indices []int) error

	// GetHand ハンドを取得する
	GetHand() []*domain.Card
	// GetPhase 現在のフェーズを取得する
	GetPhase() int
	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetBetAmount ベット額を取得する
	GetBetAmount() int
	// GetChips チップを取得する
	GetChips() int
	// GetResult ゲーム結果を取得する
	GetResult() domain.GameResult
	// GetPayout 配当金額を取得する
	GetPayout() int
	// GetHandRank ハンドランクを取得する
	GetHandRank() int
	// GetHandName ハンド名を取得する
	GetHandName() string
	// GetHeldIndices ホールドインデックスを取得する
	GetHeldIndices() [domain.VideoPokerHandSize]bool
	// GetActionLog 棋譜を取得する
	GetActionLog() []*domain.ActionLogEntry
}
