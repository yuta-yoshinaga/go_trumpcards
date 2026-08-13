//go:build !js || !wasm || solo

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// TuSacGame 四色牌ゲームインタフェース
type TuSacGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// Draw 山または捨て札から 1 枚引く
	Draw(fromDiscard bool) error
	// Meld 手札の指定の札を組み合わせとして場に出す
	Meld(indexes []int) error
	// Discard 手札から 1 枚捨てる
	Discard(index int) error
	// NextRound 次のラウンドを始める
	NextRound() error

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.TuSacConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.TuSacConfig)

	// GetPhase 現在のフェーズ
	GetPhase() domain.TuSacPhase
	// GetGameEndFlag ゲーム終了フラグ
	GetGameEndFlag() bool

	// GetPlayers 席の一覧
	GetPlayers() []*domain.TuSacPlayer
	// GetTurnSeat いまの手番
	GetTurnSeat() int
	// GetRoundNumber 何ラウンド目か
	GetRoundNumber() int
	// GetStockCount 山の残り枚数
	GetStockCount() int
	// GetDiscardTop 捨て札の一番上 (無ければ nil)
	GetDiscardTop() *domain.Card
	// GetDiscardCount 捨て札の枚数
	GetDiscardCount() int
	// GetWentOutSeat 上がった席 (-1 なら山切れ)
	GetWentOutSeat() int
	// GetResults ラウンドの結果
	GetResults() []domain.TuSacResult
	// HumanSeat 人間の席
	HumanSeat() int
	// IsHumanTurn 人間の手番か
	IsHumanTurn() bool
	// WinnerSeat 得点がいちばん高い席
	WinnerSeat() int
	// GetHint 助言
	GetHint() *domain.TuSacHint
}
