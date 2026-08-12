//go:build !js || !wasm || solo

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// SnapGame スナップゲームインタフェース
type SnapGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// PlayerStep 人間がストック先頭1枚を場に出す
	PlayerStep() error
	// PlayerSnap 人間がスナップを宣言する
	PlayerSnap() error
	// Tick 保留中の CPU アクションを進行させる
	Tick() domain.SnapPendingKind
	// GiveUp 投了する
	GiveUp()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.SnapConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.SnapConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.SnapPhase
	// IsSnapAvailable いま宣言が正しいかを返す
	IsSnapAvailable() bool
	// GetCenterPile 場札を取得する
	GetCenterPile() []*domain.Card
	// GetCenterPileSize 場札の枚数を取得する
	GetCenterPileSize() int
	// GetTopCard 場札のいちばん上を取得する (無ければ nil)
	GetTopCard() *domain.Card
	// GetCurrentTurnIdx 次にめくる席を取得する
	GetCurrentTurnIdx() int
	// GetPending 保留中の CPU アクションを取得する
	GetPending() domain.SnapPending
	// GetLastEvent 直近イベントを取得する
	GetLastEvent() domain.SnapLastEvent
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.SnapPlayer
	// GetWinnerIdx 勝った席を取得する (-1: 未確定/決着なし)
	GetWinnerIdx() int
	// GetHint ヒントを取得する
	GetHint() *domain.SnapHint
}
