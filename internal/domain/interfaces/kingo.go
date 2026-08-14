//go:build !js || !wasm || extra

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// KingoGame キンゴゲームインタフェース
type KingoGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// PlaceBet 子として張る
	PlaceBet(amount int) error
	// Deal 親として配る
	Deal() error
	// NextRound 次のラウンドを始める
	NextRound() error

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.KingoConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.KingoConfig)

	// GetPhase 現在のフェーズ
	GetPhase() domain.KingoPhase
	// GetGameEndFlag ゲーム終了フラグ
	GetGameEndFlag() bool

	// GetPlayers 席の一覧
	GetPlayers() []*domain.KingoPlayer
	// GetBankerSeat 親の席
	GetBankerSeat() int
	// GetRoundNumber 何ラウンド目か
	GetRoundNumber() int
	// GetResults ラウンドの結果
	GetResults() []domain.KingoResult
	// GetRemainingCards 山の残り枚数
	GetRemainingCards() int
	// HumanSeat 人間の席
	HumanSeat() int
	// IsHumanBanker 人間が親か
	IsHumanBanker() bool
	// IsHumanTurn 人間の入力待ちか
	IsHumanTurn() bool
	// WinnerSeat チップがいちばん多い席
	WinnerSeat() int
	// GetHint 助言
	GetHint() *domain.KingoHint
}
