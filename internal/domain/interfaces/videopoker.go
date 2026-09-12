//go:build !js || !wasm || casino

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// VideoPokerGame ビデオポーカーゲームインタフェース
type VideoPokerGame interface {
	BaseGame
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
	// GetChipsRefilled 直前の Reset が残高を補充したかを返す
	GetChipsRefilled() bool
	// GetResult ゲーム結果を取得する
	GetResult() domain.GameResult
	// GetPayout 配当金額を取得する
	GetPayout() int
	// GetHands RESULT に到達したハンド数を取得する
	GetHands() int
	// GetWins 払戻しが発生したハンド数を取得する
	GetWins() int
	// GetTotalBet セッション中に賭けたコインの累計を取得する
	GetTotalBet() int
	// GetTotalPayout セッション中に払い戻されたコインの累計を取得する
	GetTotalPayout() int
	// GetHandRank ハンドランクを取得する
	GetHandRank() int
	// GetHandName ハンド名を取得する
	GetHandName() string
	// GetHandKey 役の安定キー（ロケール非依存）を取得する
	GetHandKey() string
	// GetCurrentHandKey いま手元の5枚が配当対象の役かを安定キーで返す (無ければ "")
	GetCurrentHandKey() string
	// GetHeldIndices ホールドインデックスを取得する
	GetHeldIndices() [domain.VideoPokerHandSize]bool
	// GetVariantName バリアント名を取得する
	GetVariantName() string
}
