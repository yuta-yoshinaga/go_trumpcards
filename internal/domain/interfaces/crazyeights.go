package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// CrazyEightsGame クレイジーエイトゲームインタフェース
type CrazyEightsGame interface {
	// Control flow
	Reset()
	NextRound()
	PlayerPlay(cardIndex int) error
	PlayerChooseSuit(suit int) error
	PlayerDraw() error
	CpuPlay()
	CpuChooseSuit()
	ScoreRound()

	// Config
	GetConfig() domain.CrazyEightsConfig
	SetConfig(cfg domain.CrazyEightsConfig)

	// State readers
	GetGameEndFlag() bool
	GetPhase() domain.CrazyEightsPhase
	IsHumanTurn() bool
	GetRoundNumber() int
	GetCurrentPlayerIdx() int
	GetDiscardTop() *domain.Card
	GetDrawPileCount() int
	GetChosenSuit() int
	GetWinnerIdx() int
	GetPlayerCnt() int
	GetPlayer(i int) *domain.CrazyEightsPlayer
	GetActionLog() []*domain.ActionLogEntry
}
