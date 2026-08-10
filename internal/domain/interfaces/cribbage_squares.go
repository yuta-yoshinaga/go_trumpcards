//go:build !js || !wasm || extra2

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// CribbageSquaresGame はクリベッジ・スクエアズゲームのインタフェース。
type CribbageSquaresGame interface {
	BaseGame
	// GetGameEndFlag reports whether the game has left the playing phase.
	GetGameEndFlag() bool
	// Reset ゲームを初期化する
	Reset()
	// Place 現在のカードをセル(row, col)に配置する
	Place(row, col int) error
	// Undo 直前の操作を取り消す
	Undo() error
	// CanUndo アンドゥ可能かを返す
	CanUndo() bool
	// GiveUp ギブアップする
	GiveUp()
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.CribbageSquaresPhase
	// GetBoard ボードを取得する
	GetBoard() [domain.CribbageSquaresGridSize][domain.CribbageSquaresGridSize]*domain.Card
	// GetCurrentCard 次に配置するカードを取得する
	GetCurrentCard() *domain.Card
	// GetPlacedCount 配置済みカード枚数を取得する
	GetPlacedCount() int
	// IsComplete ゲームが完了したかを返す
	IsComplete() bool
	// RowScore 行の得点を返す
	RowScore(r int) int
	// ColScore 列の得点を返す
	ColScore(c int) int
	// TotalScore 合計得点を返す
	TotalScore() int
	// GetStarter スターター（17 枚目）を返す。めくる前は nil
	GetStarter() *domain.Card
	// RowDetail 行のクリベッジ得点内訳を返す
	RowDetail(r int) domain.CribbageScoreDetail
	// ColDetail 列のクリベッジ得点内訳を返す
	ColDetail(c int) domain.CribbageScoreDetail
	// IsWin 合計得点がクリア基準に達したかを返す
	IsWin() bool
	// GetHint 現在のカードを置く最善のセルを返す (無い場合 nil)
	GetHint() *domain.CribbageSquaresHint
}
