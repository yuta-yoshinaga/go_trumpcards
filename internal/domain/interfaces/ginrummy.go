package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// GinRummyGame ジンラミーゲームインタフェース
type GinRummyGame interface {
	// Control flow
	Reset()
	NextRound()
	PlayerDrawFromStock() error
	PlayerDrawFromDiscard() error
	PlayerDiscard(cardIndex int) error
	PlayerKnock(cardIndex int) error
	PlayerLayoff(cardIndices []int) error
	CpuPlay()
	ScoreRound()

	// Config
	GetConfig() domain.GinRummyConfig
	SetConfig(cfg domain.GinRummyConfig)

	// State readers
	GetGameEndFlag() bool
	GetPhase() domain.GinRummyPhase
	IsHumanTurn() bool
	GetRoundNumber() int
	GetCurrentPlayerIdx() int
	GetDiscardTop() *domain.Card
	GetDrawPileCount() int
	GetWinnerIdx() int
	GetPlayerCnt() int
	GetPlayer(i int) *domain.GinRummyPlayer
	GetActionLog() []*domain.ActionLogEntry
	GetKnockerIdx() int
	GetKnockerMelds() [][]*domain.Card
	GetKnockerDeadwood() []*domain.Card
	GetIsGin() bool
}
