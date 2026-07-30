//go:build !js || !wasm || extra3

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// SkitgubbeGame シートグッベゲームインタフェース
type SkitgubbeGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// PlayCard プレイヤーが手札の札を出す
	PlayCard(player, handIdx int) error
	// PickUp 第2フェーズで場の札を引き取る
	PickUp(player int) error
	// SkitgubbeCpuDecide CPU が取る手を決める
	SkitgubbeCpuDecide(idx int) domain.SkitgubbeCpuAction

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.SkitgubbeConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.SkitgubbeConfig)

	// GetGameEndFlag 終局しているかを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.SkitgubbePhase
	// GetCurrentPlayerIdx 手番のプレイヤー添字を取得する
	GetCurrentPlayerIdx() int
	// GetStockCount 山札の残り枚数を取得する
	GetStockCount() int
	// GetTrumpSuit 切札のスートを取得する (-1: 未確定)
	GetTrumpSuit() int
	// GetDuel 第1フェーズで場に出ている札を取得する
	GetDuel() []*domain.Card
	// GetDuelLeader 第1フェーズのリード側の添字を取得する
	GetDuelLeader() int
	// GetPile 第2フェーズで場に出ている札を取得する
	GetPile() []*domain.Card
	// GetCollectedCount 第1フェーズで集めた枚数を取得する
	GetCollectedCount(idx int) int
	// GetValidPlayIndices 出せる手札の添字を取得する
	GetValidPlayIndices(player int) []int
	// IsFinished 指定プレイヤーが上がっているかを取得する
	IsFinished(idx int) bool
	// GetPlayers 全プレイヤーを取得する
	GetPlayers() []*domain.SkitgubbePlayer
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(idx int) *domain.SkitgubbePlayer
	// GetLoserIdx 敗者 (Skitgubbe) の添字を取得する (-1: 未決着)
	GetLoserIdx() int
}
