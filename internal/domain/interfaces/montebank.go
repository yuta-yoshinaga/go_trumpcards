//go:build !js || !wasm || casino

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// MonteBankGame モンテバンクゲームインタフェース
type MonteBankGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// PlaceBet 場札 idx に賭けてゲートをめくる
	PlaceBet(idx, bet int) error
	// NextRound 次のラウンドを始める
	NextRound() error

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.MonteBankConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.MonteBankConfig)

	// GetPhase 現在のフェーズ
	GetPhase() domain.MonteBankPhase
	// GetGameEndFlag ゲーム終了フラグ
	GetGameEndFlag() bool

	// GetLayout 場札
	GetLayout() []*domain.Card
	// GetGate ゲートの札 (めくる前は nil)
	GetGate() *domain.Card
	// GetPick 賭けた場札の位置 (賭ける前は -1)
	GetPick() int
	// GetBet 賭け金
	GetBet() int
	// GetResult ラウンドの決着
	GetResult() domain.MonteBankResult
	// GetPayout このラウンドで戻ってきた総額
	GetPayout() int

	// SuitCountInLayout 場札に指定スートが何枚出ているか
	SuitCountInLayout(design int) int
	// RemainingOfSuit 場札を除いた残りに指定スートが何枚あるか
	RemainingOfSuit(design int) int

	// GetChips 保有チップ数
	GetChips() int
	// GetRoundNumber ラウンド数
	GetRoundNumber() int
	// GetRemainingCards 山の残り枚数
	GetRemainingCards() int
	// GetHint 助言
	GetHint() *domain.MonteBankHint
}
