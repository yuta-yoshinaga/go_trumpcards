//go:build !js || !wasm || extra3

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// StealingBundlesGame スティーリングバンドルゲームインタフェース
type StealingBundlesGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// PlayerTake 人間が場札を取る
	PlayerTake(cardIndex int) error
	// PlayerSteal 人間が相手の束を奪う
	PlayerSteal(cardIndex, victimIdx int) error
	// PlayerTrail 人間が場に置く
	PlayerTrail(cardIndex int) error
	// CpuPlay CPUが1手打つ
	CpuPlay()
	// GiveUp 投了する
	GiveUp()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.StealingBundlesConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.StealingBundlesConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.StealingBundlesPhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// GetTableMatches 手札で取れる場札の位置を返す
	GetTableMatches(playerIdx, cardIndex int) []int
	// GetStealTargets 手札で奪える相手の席を返す
	GetStealTargets(playerIdx, cardIndex int) []int
	// CanCapture 取れる手があるかを返す
	CanCapture(playerIdx int) bool
	// GetTableCards 場札を取得する
	GetTableCards() []*domain.Card
	// GetDeckRemaining 山札の残り枚数を取得する
	GetDeckRemaining() int
	// GetLastCaptureIdx 最後に取った席を取得する (-1: まだ)
	GetLastCaptureIdx() int
	// GetCurrentPlayerIdx 現在の手番を取得する
	GetCurrentPlayerIdx() int
	// GetTurnNumber 打たれた手の数を取得する
	GetTurnNumber() int
	// GetPacksDealt 配ったパック数を取得する
	GetPacksDealt() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.StealingBundlesPlayer
	// GetWinnerIdx 勝った席を取得する (-1: 未確定)
	GetWinnerIdx() int
	// GetHint ヒントを取得する
	GetHint() *domain.StealingBundlesHint
}
