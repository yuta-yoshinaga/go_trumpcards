//go:build !js || !wasm || solo

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// PokerSquaresGame はポーカー・スクエアズゲームのインタフェース。
type PokerSquaresGame interface {
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
	GetPhase() domain.PokerSquaresPhase
	// GetBoard ボードを取得する
	GetBoard() [domain.PokerSquaresGridSize][domain.PokerSquaresGridSize]*domain.Card
	// GetCurrentCard 次に配置するカードを取得する
	GetCurrentCard() *domain.Card
	// GetPlacedCount 配置済みカード枚数を取得する
	GetPlacedCount() int
	// IsComplete ゲームが完了したかを返す
	IsComplete() bool
	// EvaluateRow 行のハンドランクを返す (未完成時 -1)
	EvaluateRow(r int) int
	// EvaluateCol 列のハンドランクを返す (未完成時 -1)
	EvaluateCol(c int) int
	// RowScore 行の得点を返す
	RowScore(r int) int
	// ColScore 列の得点を返す
	ColScore(c int) int
	// TotalScore 合計得点を返す
	TotalScore() int
	// GetHint 現在のカードを置く最善のセルを返す (無い場合 nil)
	GetHint() *domain.PokerSquaresHint
}
