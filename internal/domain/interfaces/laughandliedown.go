//go:build !js || !wasm || extra2

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// LaughAndLieDownGame ラフ・アンド・ライダウンゲームインタフェース
type LaughAndLieDownGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// PlayCard 手札の札を出し、場から takeCount 枚を取る
	PlayCard(player, handIdx, takeCount int) error
	// CanTakeThree 3 枚取りができるかを返す
	CanTakeThree(player, handIdx int) bool
	// LaughAndLieDownCpuDecide CPU が取る手を決める
	LaughAndLieDownCpuDecide(idx int) domain.LaughAndLieDownCpuAction

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.LaughAndLieDownConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.LaughAndLieDownConfig)

	// GetGameEndFlag 終局しているかを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.LaughAndLieDownPhase
	// GetCurrentPlayerIdx 手番のプレイヤー添字を取得する
	GetCurrentPlayerIdx() int
	// GetLayout 表向きの場札を取得する
	GetLayout() []*domain.Card
	// GetValidPlayIndices 出せる手札の添字を取得する
	GetValidPlayIndices(player int) []int
	// GetWonCount 取得枚数を取得する
	GetWonCount(idx int) int
	// IsLaidDown 降りているかを取得する
	IsLaidDown(idx int) bool
	// GetScore 収支を取得する
	GetScore(idx int) int
	// GetDealerIdx 親の添字を取得する
	GetDealerIdx() int
	// GetLastInIdx 最後まで手札が残っていた人の添字を取得する (-1: 該当なし)
	GetLastInIdx() int
	// GetPlayers 全プレイヤーを取得する
	GetPlayers() []*domain.LaughAndLieDownPlayer
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(idx int) *domain.LaughAndLieDownPlayer
}
