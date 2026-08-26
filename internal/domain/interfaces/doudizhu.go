//go:build !js || !wasm || extra4

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// DoudizhuGame 斗地主ゲームインタフェース
type DoudizhuGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// PlayerBid 人間プレイヤーがビッドする
	PlayerBid(value int) error
	// PlayerPlay プレイヤーがカードを出す
	PlayerPlay(indices []int) error
	// CpuPlay CPUプレイヤーが1ターン実行する
	CpuPlay()
	// HasPendingAction ペンディングアクションがあるかを返す
	HasPendingAction() bool
	// SetConfig ゲーム設定をセットする
	SetConfig(config domain.DoudizhuConfig)

	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.DoudizhuPhase
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.DoudizhuPlayer
	// GetCurrentTurn 現在の手番プレイヤーインデックスを取得する
	GetCurrentTurn() int
	// GetTableCombo 場の役を取得する
	GetTableCombo() *domain.DoudizhuCombo
	// GetLastPlayIdx 最後にカードを出したプレイヤーインデックスを取得する
	GetLastPlayIdx() int
	// GetKittyCards 底牌を取得する
	GetKittyCards() []*domain.Card
	// GetLandlordIdx 地主プレイヤーインデックスを取得する
	GetLandlordIdx() int
	// GetBaseBid ビッド値を取得する
	GetBaseBid() int
	// GetBombCount ボム/ロケット使用回数を取得する
	GetBombCount() int
	// GetScores スコアを取得する
	GetScores() [domain.DoudizhuPlayerCnt]int
	// GetBidValues 全プレイヤーのビッド値を取得する
	GetBidValues() [domain.DoudizhuPlayerCnt]int
	// GetHighestBid 現在の最高ビッド値を取得する
	GetHighestBid() int
	// GetCpuActions CPU行動記録一覧を取得する
	GetCpuActions() []*domain.DoudizhuCpuAction
	// GetHumanAction 人間の最後の行動記録を取得する
	GetHumanAction() *domain.DoudizhuCpuAction
	// GetConfig ゲーム設定を取得する
	GetConfig() domain.DoudizhuConfig
}
