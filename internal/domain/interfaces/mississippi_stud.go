//go:build !js || !wasm || casino

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// MississippiStudGame ミシシッピ・スタッドゲームインタフェース
type MississippiStudGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// Bet アンティをベットしカードを配る
	Bet(amount int) error
	// Play 1x/2x/3x のストリートベットを置きフェーズを進める
	Play(multiplier int) error
	// Fold ベットを没収しゲームを終了する
	Fold() error

	// GetPlayerHand ホールカードを取得する
	GetPlayerHand() []*domain.Card
	// GetCommunityCards コミュニティカードを取得する
	GetCommunityCards() []*domain.Card
	// GetCommunityRevealed コミュニティカードの公開状態を取得する
	GetCommunityRevealed() [domain.MississippiStudCommunityCnt]bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() int
	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetAnteAmount アンティ額を取得する
	GetAnteAmount() int
	// GetStreetMultipliers ストリートベット倍率を取得する
	GetStreetMultipliers() [domain.MississippiStudStreetCnt]int
	// GetFolded フォールドフラグを取得する
	GetFolded() bool
	// GetTotalBet 投入チップ総量を取得する
	GetTotalBet() int
	// GetResult ゲーム結果を取得する
	GetResult() domain.GameResult
	// GetHandRank ハンドランクを取得する
	GetHandRank() int
	// GetCurrentMadeHand 既知のカードからできている役を取得する
	GetCurrentMadeHand() *domain.MississippiStudMadeHand
	// RecommendBet 現在のストリートでの推奨アクションを取得する
	RecommendBet() string
	// GetPayoutMultiplier 適用された配当倍率を取得する
	GetPayoutMultiplier() int
	// GetAntePayout アンティ部分の配当を取得する
	GetAntePayout() int
	// GetStreetPayouts ストリート部分の配当を取得する
	GetStreetPayouts() [domain.MississippiStudStreetCnt]int
	// GetTotalPayout 合計配当を取得する
	GetTotalPayout() int
	// GetChips チップを取得する
	GetChips() int
}
