//go:build !js || !wasm || casino

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// AndarBaharGame アンダーバハールゲームインタフェース
type AndarBaharGame interface {
	BaseGame
	// Reset ラウンドを初期化する (罫線は保持)
	Reset()
	// ClearHistory 罫線履歴をクリアする
	ClearHistory()
	// Bet ベットして決着まで自動進行する
	Bet(amount, target, sideAmount, sideBand int) error

	// GetJoker 基準札
	GetJoker() *domain.Card
	// GetAndarCards アンダーに配られた札
	GetAndarCards() []*domain.Card
	// GetBaharCards バハールに配られた札
	GetBaharCards() []*domain.Card
	// GetFirstColumn 先に配る列 (基準札の色で決まる)
	GetFirstColumn() int
	// DealtCount 決着までに配った枚数
	DealtCount() int
	// GetPhase 現在のフェーズ
	GetPhase() int
	// GetGameEndFlag ゲーム終了フラグ
	GetGameEndFlag() bool
	// GetBetAmount メインベット額
	GetBetAmount() int
	// GetBetTarget メインベット先の列
	GetBetTarget() int
	// GetSideAmount サイドベット額
	GetSideAmount() int
	// GetSideBand サイドベットの帯
	GetSideBand() int
	// GetWinner 基準札と同ランクが出た列 (-1: 未決着)
	GetWinner() int
	// GetResult 勝敗結果
	GetResult() domain.GameResult
	// GetPayout 払戻総額
	GetPayout() int
	// GetChips チップ
	GetChips() int
	// GetHistory 罫線履歴
	GetHistory() []int
	// GetHint 助言のキー
	GetHint() string
}
