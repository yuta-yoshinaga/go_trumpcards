//go:build !js || !wasm || extra3

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// RollingStoneGame ローリングストーンゲームインタフェース
type RollingStoneGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// PlayerPlay プレイヤーがカードを出す
	PlayerPlay(cardIndex int) error
	// PlayerPickUp プレイヤーが場札を引き取る
	PlayerPickUp() error
	// CpuPlay CPUプレイヤーが1手打つ
	CpuPlay()
	// GiveUp 投了する
	GiveUp()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.RollingStoneConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.RollingStoneConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.RollingStonePhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// MustPickUp 指定席がフォローできず引き取るしかないかを返す
	MustPickUp(playerIdx int) bool
	// GetLeadSuit このトリックのリードスートを返す（誰も出していなければ 0）
	GetLeadSuit() int
	// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す
	GetValidPlayIndices(playerIdx int) []int
	// GetCurrentTrick 現在のトリックを取得する
	GetCurrentTrick() []*domain.TrickCard
	// GetCurrentPlayerIdx 現在のプレイヤーインデックスを取得する
	GetCurrentPlayerIdx() int
	// GetLeadPlayerIdx リード席を取得する
	GetLeadPlayerIdx() int
	// GetTrickNumber 解決済みのトリック数を取得する
	GetTrickNumber() int
	// GetLastPickupIdx 直前に引き取った席を取得する (-1: 無し)
	GetLastPickupIdx() int
	// GetFinishedCnt 上がった人数を取得する
	GetFinishedCnt() int
	// GetDiscarded 場から抜けた札の枚数を取得する
	GetDiscarded() int
	// GetDeckSize この卓で使うデッキ枚数を取得する
	GetDeckSize() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.RollingStonePlayer
	// GetWinnerIdx 勝った席を取得する (-1: 未確定)
	GetWinnerIdx() int
	// GetHint ヒントを取得する
	GetHint() *domain.RollingStoneHint
}
