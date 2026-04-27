package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// SlapjackGame スラップジャックゲームインタフェース
type SlapjackGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// ResetWithConfig 設定を更新してゲームを初期化する
	ResetWithConfig(cfg domain.SlapjackConfig)
	// Step 現手番プレイヤーがストック先頭1枚を場に出す
	Step() error
	// Slap 指定プレイヤーがスラップを試みる
	Slap(playerIdx int) error
	// Tick 保留中の CPU アクションを進行させる
	Tick() domain.SlapjackPendingKind

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.SlapjackConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.SlapjackConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.SlapjackPhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.SlapjackPlayer
	// GetWinnerIdx 勝者インデックスを取得する
	GetWinnerIdx() int
	// GetCenterPileSize 場の総枚数を取得する
	GetCenterPileSize() int
	// GetTopCard 場のトップカードを取得する
	GetTopCard() *domain.Card
	// GetCurrentTurnIdx 現在の手番プレイヤーを取得する
	GetCurrentTurnIdx() int
	// IsTopJack 場のトップが J かを返す
	IsTopJack() bool
	// GetPending 保留中の CPU アクションを取得する
	GetPending() domain.SlapjackPending
	// GetLastEvent 直近イベントを取得する
	GetLastEvent() domain.SlapjackLastEvent
}
