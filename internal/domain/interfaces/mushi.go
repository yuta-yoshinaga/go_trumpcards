//go:build !js || !wasm || extra2

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// MushiGame 虫ゲームインタフェース
type MushiGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// PlayCard プレイヤーが手札の札を出す
	PlayCard(player, handIdx int) error
	// SelectCapture 選択フェーズで取る場札を決める
	SelectCapture(player, fieldIdx int) error
	// NextRound 次のラウンドを開始する
	NextRound() error
	// MushiCpuDecide CPU が取る手を決める
	MushiCpuDecide(idx int) domain.MushiCpuAction

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.MushiConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.MushiConfig)

	// GetGameEndFlag 終局しているかを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.MushiPhase
	// GetCurrentPlayerIdx 手番のプレイヤー添字を取得する
	GetCurrentPlayerIdx() int
	// GetDealerIdx 親の添字を取得する
	GetDealerIdx() int
	// GetRoundNumber 現在のラウンド番号を取得する
	GetRoundNumber() int
	// GetField 場札を取得する
	GetField() []*domain.Card
	// GetStockCount 山札の残り枚数を取得する
	GetStockCount() int
	// GetCaptured 指定プレイヤーの取り札を取得する
	GetCaptured(idx int) []*domain.Card
	// GetPlayers 全プレイヤーを取得する
	GetPlayers() []*domain.MushiPlayer
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(idx int) *domain.MushiPlayer
	// GetScore 累計得点を取得する
	GetScore(idx int) int
	// GetRoundResult 直前のラウンドでの増減を取得する
	GetRoundResult(idx int) int
	// GetPendingCard 選択待ちの札を取得する
	GetPendingCard() *domain.Card
	// GetSelectableIndices 選択できる場札の添字を取得する
	GetSelectableIndices() []int
	// GetWinnerIdx 勝者の添字を取得する (-1: 未確定または引き分け)
	GetWinnerIdx() int
}
