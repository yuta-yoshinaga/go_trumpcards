package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// DoubtGame ダウトゲームインタフェース
type DoubtGame interface {
	// interactor が呼ぶメソッド
	Reset()
	PlayerPlay(cardIndices []int, claimedValue int) error
	CpuPlay()
	ResolveDoubt(doubterIndices []int)
	SkipDoubt()

	// state readers
	GetGameEndFlag() bool
	GetPhase() domain.DoubtPhase
	IsHumanTurn() bool
	GetCurrentTurn() int
	GetPlayerCnt() int
	GetPlayer(i int) *domain.DoubtPlayer
	GetTableCardCount() int
	GetTableCards() []*domain.Card
	GetLastAction() *domain.DoubtAction
	GetCpuDoubters() []int
	GetWinnerIdx() int
	GetCpuActions() []*domain.DoubtCpuAction
	GetHumanAction() *domain.DoubtCpuAction
	GetLastDoubtResult() *domain.DoubtDoubtResult
}
