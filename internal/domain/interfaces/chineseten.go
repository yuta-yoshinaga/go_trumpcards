//go:build !js || !wasm || extra2

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// ChineseTenGame 撿紅點ゲームインタフェース
type ChineseTenGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// PlayCard プレイヤーが手札の札を出す
	PlayCard(player, handIdx int) error
	// SelectCapture 選択フェーズで取る場札を決める
	SelectCapture(player, layoutIdx int) error
	// ChineseTenCpuDecide CPU が取る手を決める
	ChineseTenCpuDecide(idx int) domain.ChineseTenCpuAction

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.ChineseTenConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.ChineseTenConfig)

	// GetGameEndFlag 終局しているかを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.ChineseTenPhase
	// GetCurrentPlayerIdx 手番のプレイヤー添字を取得する
	GetCurrentPlayerIdx() int
	// GetLayout 場札を取得する
	GetLayout() []*domain.Card
	// GetStockCount 山札の残り枚数を取得する
	GetStockCount() int
	// GetCaptured 指定プレイヤーの取り札を取得する
	GetCaptured(idx int) []*domain.Card
	// GetPlayers 全プレイヤーを取得する
	GetPlayers() []*domain.ChineseTenPlayer
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(idx int) *domain.ChineseTenPlayer
	// GetScore 得点を取得する
	GetScore(idx int) int
	// GetPendingCard 選択待ちの札を取得する
	GetPendingCard() *domain.Card
	// GetSelectableIndices 選択できる場札の添字を取得する
	GetSelectableIndices() []int
	// GetWinnerIdx 勝者の添字を取得する (-1: 引き分け)
	GetWinnerIdx() int
}
