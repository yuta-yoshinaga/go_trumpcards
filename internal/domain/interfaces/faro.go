//go:build !js || !wasm || extra2

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// FaroGame はファロゲームのインタフェース。
type FaroGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のディールを開始する
	NextRound()
	// PlayerPlaceBet ランクにベットを置く（copper=敗北予想）
	PlayerPlaceBet(rank, amount int, copper bool) error
	// PlayerClearBet 指定ランクのベットを取り消す
	PlayerClearBet(rank int) error
	// PlayerClearAll すべてのベットを取り消す
	PlayerClearAll() error
	// PlayerDealTurn 2枚をめくってベットを解決する
	PlayerDealTurn() error
	// PlayerCall 残り3枚の順序を予想する
	PlayerCall(order []int) error

	// GetPhase 現在のフェーズ
	GetPhase() int
	// GetGameEndFlag ゲーム終了フラグ
	GetGameEndFlag() bool
	// GetChips チップ
	GetChips() int
	// GetTurnsPlayed 消化済みターン数
	GetTurnsPlayed() int
	// GetTurnsTotal 1ディールの総ターン数
	GetTurnsTotal() int
	// GetRemainingCount デッキ残枚数
	GetRemainingCount() int
	// FaroRemainingByRank ランク別の残り枚数 (index 1..13 が A..K)
	FaroRemainingByRank() [domain.FaroMaxRank + 1]int
	// GetSoda 焼かれたソーダ札
	GetSoda() *domain.Card
	// GetLastTurn 直近ターンの結果
	GetLastTurn() *domain.FaroTurnResult
	// GetCallCards コール対象の残り3枚
	GetCallCards() []*domain.Card
	// GetCallOrder 宣言したコール順
	GetCallOrder() []int
	// GetCallWon コール成功フラグ
	GetCallWon() bool
	// GetTotalPayout 直近ラウンドの純損益
	GetTotalPayout() int
	// GetBets レイアウト上の全ベット
	GetBets() map[int]*domain.FaroBet
	// GetBetRanks ベット中のランク（昇順）
	GetBetRanks() []int
}
