//go:build !js || !wasm || extra4

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// BhabhiGame バービー (Bhabhi) ゲームインタフェース
type BhabhiGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// PlayerPlay プレイヤーがカードを出す
	PlayerPlay(cardIndex int) error
	// CpuPlay CPUプレイヤーが1枚出す
	CpuPlay()
	// GiveUp 投了する
	GiveUp()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.BhabhiConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.BhabhiConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.BhabhiPhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// GetTrickNumber これまでに閉じたトリック数を取得する
	GetTrickNumber() int
	// GetLeadSuit リードスートを取得する (0: トリック未開始)
	GetLeadSuit() int
	// GetPile 場に出ている札を取得する
	GetPile() []*domain.TrickCard
	// GetLastPickupIdx 直前に場札を引き取った人を取得する (-1: まだ無い)
	GetLastPickupIdx() int
	// GetLastPickupSize 直前に引き取った枚数を取得する
	GetLastPickupSize() int
	// GetCurrentPlayerIdx 現在のプレイヤーインデックスを取得する
	GetCurrentPlayerIdx() int
	// GetLeadPlayerIdx リードプレイヤーインデックスを取得する
	GetLeadPlayerIdx() int
	// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す
	GetValidPlayIndices(playerIdx int) []int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.BhabhiPlayer
	// GetAliveCount まだ手札が残っている人数を取得する
	GetAliveCount() int
	// GetBhabhiIdx 敗者を取得する (-1: 未確定)
	GetBhabhiIdx() int
	// IsStalemate 膠着で打ち切られたかを返す
	IsStalemate() bool
	// GetHint ヒントを取得する
	GetHint() *domain.BhabhiHint
}
