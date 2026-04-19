package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// TrashGame トラッシュゲームインタフェース
type TrashGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// Draw 山札から1枚引いて連鎖を解決する
	Draw() error
	// PlaceWild ワイルドを指定位置に配置する
	PlaceWild(pos int) error
	// CpuStep CPUのターンを1ステップ進める
	CpuStep() error
	// IsCpuTurn 現在のターンがCPUか
	IsCpuTurn() bool
	// IsCpuPlayer プレイヤーがCPUか
	IsCpuPlayer(idx int) bool
	// GetPhase 現在のフェーズ
	GetPhase() domain.TrashPhase
	// GetCurrent 現在ターンのプレイヤーインデックス
	GetCurrent() int
	// GetMoveCount ドロー回数
	GetMoveCount() int
	// GetStockSize 山札残り枚数
	GetStockSize() int
	// GetDiscardSize 捨て札の枚数
	GetDiscardSize() int
	// GetDiscardTop 捨て札の一番上
	GetDiscardTop() *domain.Card
	// GetPending 連鎖中のpendingカード
	GetPending() *domain.Card
	// GetPlayerSlots プレイヤーのスロット一覧
	GetPlayerSlots(idx int) []domain.TrashSlot
	// GetWinner 勝者インデックス (-1 なら未決着)
	GetWinner() int
}
