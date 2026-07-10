package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// PrsiGame プルシー(Prší / チェコ版クレイジーエイト)ゲームインタフェース
type PrsiGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// PlayerPlay プレイヤーがカードを出す
	PlayerPlay(cardIndex int) error
	// PlayerDraw プレイヤーが山札からカードを引く
	PlayerDraw() error
	// CpuPlay CPUプレイヤーが1ターン実行する
	CpuPlay()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.PrsiConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.PrsiConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.PrsiPhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// GetCurrentPlayerIdx 現在のプレイヤーインデックスを取得する
	GetCurrentPlayerIdx() int
	// GetDiscardTop 捨て札の一番上のカードを取得する
	GetDiscardTop() *domain.Card
	// GetDrawPileCount 山札の残り枚数を取得する
	GetDrawPileCount() int
	// GetPenaltyDrawCount 累積7ペナルティ枚数を取得する
	GetPenaltyDrawCount() int
	// GetPendingSkips 累積スキップ数を取得する
	GetPendingSkips() int
	// GetWinnerIdx 勝者インデックスを取得する
	GetWinnerIdx() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.PrsiPlayer
}
