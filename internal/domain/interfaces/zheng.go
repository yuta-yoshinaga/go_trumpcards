//go:build !js || !wasm || solo

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// ZhengGame 争上游ゲームインタフェース
type ZhengGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// PlayerPlay プレイヤーがカードを出す
	PlayerPlay(indices []int) error
	// CpuPlay CPUプレイヤーが1ターン実行する
	CpuPlay()
	// HasPendingAction ペンディングアクションがあるかを返す
	HasPendingAction() bool
	// SetConfig ゲーム設定をセットする
	SetConfig(config domain.ZhengConfig)

	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.ZhengPlayer
	// GetTableCards 場のカード一覧を取得する
	GetTableCards() []*domain.Card
	// GetTablePlayType 場のプレイタイプを取得する
	GetTablePlayType() domain.ZhengPlayType
	// GetLastPlayPlayerIdx 最後にカードを出したプレイヤーインデックスを取得する
	GetLastPlayPlayerIdx() int
	// GetHumanAction 人間の最後の行動記録を取得する
	GetHumanAction() *domain.ZhengAction
	// GetCpuActions CPU行動記録一覧を取得する
	GetCpuActions() []*domain.ZhengAction
	// GetCurrentTurn 現在の手番プレイヤーインデックスを取得する
	GetCurrentTurn() int
	// GetConfig ゲーム設定を取得する
	GetConfig() domain.ZhengConfig
	// GetPassCount パス回数を取得する
	GetPassCount() int
}
