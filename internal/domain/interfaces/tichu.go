//go:build !js || !wasm || extra2

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// TichuGame ティチューゲームインタフェース
type TichuGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// PlayerDeclare 人間プレイヤーが宣言する
	PlayerDeclare(declType int) error
	// PlayerPlay プレイヤーがカードを出す
	PlayerPlay(indices []int) error
	// CpuPlay CPUプレイヤーが1ターン実行する
	CpuPlay()
	// HasPendingAction ペンディングアクションがあるかを返す
	HasPendingAction() bool
	// SetConfig ゲーム設定をセットする
	SetConfig(config domain.TichuConfig)

	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.TichuPhase
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.TichuPlayer
	// GetCurrentTurn 現在の手番プレイヤーインデックスを取得する
	GetCurrentTurn() int
	// GetTableCombo 場の役を取得する
	GetTableCombo() *domain.TichuCombo
	// GetLastPlayIdx 最後にカードを出したプレイヤーインデックスを取得する
	GetLastPlayIdx() int
	// GetStartLeader 先手プレイヤーを取得する
	GetStartLeader() int
	// GetFinishOrder 上がり順を取得する
	GetFinishOrder() []int
	// GetScores チーム得点を取得する
	GetScores() [2]int
	// GetIsOneTwo ワンツーかどうかを取得する
	GetIsOneTwo() bool
	// GetBombCount ボム使用回数を取得する
	GetBombCount() int
	// GetCpuActions CPU行動記録一覧を取得する
	GetCpuActions() []*domain.TichuCpuAction
	// GetHumanAction 人間の最後の行動記録を取得する
	GetHumanAction() *domain.TichuCpuAction
	// GetConfig ゲーム設定を取得する
	GetConfig() domain.TichuConfig
}
