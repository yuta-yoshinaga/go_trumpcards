package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// PigsTailGame ぶたのしっぽゲームインタフェース
type PigsTailGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// SetConfig ゲーム設定をセットする
	SetConfig(config domain.PigsTailConfig)
	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// PlayerAction 人間プレイヤーのアクション (山札から1枚引く)
	PlayerAction(actionIdx int) error
	// CpuAction CPUプレイヤーが1ターン実行する
	CpuAction() error
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.PigsTailPlayer
	// GetCurrentTurn 現在の手番プレイヤーインデックスを取得する
	GetCurrentTurn() int
	// GetLoserIdx 負けプレイヤーインデックスを取得する
	GetLoserIdx() int
	// GetLastDrawCard 最後に引いたカードを取得する
	GetLastDrawCard() *domain.Card
	// GetLastPenalty 最後のアクションでペナルティが発生したかを取得する
	GetLastPenalty() bool
	// GetCpuActions CPUターンの行動履歴を取得する
	GetCpuActions() []*domain.PigsTailCpuAction
	// GetHumanAction 人間の最後の行動記録を取得する
	GetHumanAction() *domain.PigsTailCpuAction
	// GetCircleCount 山札の残り枚数を取得する
	GetCircleCount() int
	// GetCenter 中央の場札を取得する
	GetCenter() []*domain.Card
	// GetCenterTopCard 場札の一番上のカードを取得する
	GetCenterTopCard() *domain.Card
	// GetConfig ゲーム設定を取得する
	GetConfig() domain.PigsTailConfig
}
