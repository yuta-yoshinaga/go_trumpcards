//go:build !js || !wasm || solo

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// MemoryGame 神経衰弱ゲームインタフェース
type MemoryGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// PlayerFlip プレイヤーがカードをめくる
	PlayerFlip(pos int) error
	// CpuFlip CPUプレイヤーがカードをめくる
	CpuFlip()
	// ResolveFlip めくり結果を判定する
	ResolveFlip()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.MemoryConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.MemoryConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.MemoryPhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// GetCurrentPlayerIdx 現在のプレイヤーインデックスを取得する
	GetCurrentPlayerIdx() int
	// GetFirstFlipPos 1枚目のめくり位置を取得する
	GetFirstFlipPos() int
	// GetSecondFlipPos 2枚目のめくり位置を取得する
	GetSecondFlipPos() int
	// GetLastMatchResult 最後のめくりがマッチしたかを返す
	GetLastMatchResult() bool
	// GetTurnNumber 現在のターン番号を取得する
	GetTurnNumber() int
	// GetWinnerIdx 勝者インデックスを取得する
	GetWinnerIdx() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.MemoryPlayer
	// GetBoard ボード全体を取得する
	GetBoard() []*domain.MemoryBoardCard
	// GetBoardCard 指定位置のボードカードを取得する
	GetBoardCard(pos int) *domain.MemoryBoardCard
}
