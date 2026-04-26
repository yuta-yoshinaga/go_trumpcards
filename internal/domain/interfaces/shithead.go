package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// ShitheadGame シットヘッドゲームインタフェース
type ShitheadGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// PlayerPlay プレイヤーがカードを出す (空 = ピックアップ)
	PlayerPlay(indices []int) error
	// CpuPlay CPUプレイヤーが1ターン実行する
	CpuPlay()
	// SetConfig ゲーム設定をセットする
	SetConfig(config domain.ShitheadConfig)

	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.ShitheadPlayer
	// GetCurrentTurn 現在の手番プレイヤーインデックスを取得する
	GetCurrentTurn() int
	// GetDiscardPile 場札を取得する
	GetDiscardPile() []*domain.Card
	// GetTopCard 場札の一番上のカードを取得する
	GetTopCard() *domain.Card
	// GetStockSize 山札の残り枚数を取得する
	GetStockSize() int
	// GetCpuActions CPU行動記録一覧を取得する
	GetCpuActions() []*domain.ShitheadCpuAction
	// GetHumanAction 人間の最後の行動記録を取得する
	GetHumanAction() *domain.ShitheadCpuAction
	// GetSkipNext 8効果が有効かを返す
	GetSkipNext() bool
	// GetSevenActive 7効果が有効かを返す
	GetSevenActive() bool
	// GetConfig ゲーム設定を取得する
	GetConfig() domain.ShitheadConfig
	// CurrentSource 現在のプレイヤーが出すべき場所を返す
	CurrentSource() string
}
