package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// MemoryGame 神経衰弱ゲームインタフェース
type MemoryGame interface {
	// interactor が呼ぶメソッド
	Reset()
	PlayerFlip(pos int) error
	CpuFlip()
	ResolveFlip()

	// config
	GetConfig() domain.MemoryConfig
	SetConfig(cfg domain.MemoryConfig)

	// state readers
	GetGameEndFlag() bool
	GetPhase() domain.MemoryPhase
	IsHumanTurn() bool
	GetCurrentPlayerIdx() int
	GetFirstFlipPos() int
	GetSecondFlipPos() int
	GetLastMatchResult() bool
	GetTurnNumber() int
	GetWinnerIdx() int
	GetPlayerCnt() int
	GetPlayer(i int) *domain.MemoryPlayer
	GetBoard() [domain.MemoryBoardSize]*domain.MemoryBoardCard
	GetBoardCard(pos int) *domain.MemoryBoardCard
	GetActionLog() []*domain.ActionLogEntry
}
