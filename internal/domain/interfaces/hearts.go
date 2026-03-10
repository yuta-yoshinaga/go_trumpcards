package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// HeartsGame ハーツゲームインタフェース
type HeartsGame interface {
	// interactor が呼ぶメソッド
	Reset()
	NextRound()
	PlayerPass(cardIndices []int) error
	CpuPass()
	ExecutePass()
	PlayerPlay(cardIndex int) error
	CpuPlay()
	ResolveTrick()
	NextTrick()
	ScoreRound()

	// config
	GetConfig() domain.HeartsConfig
	SetConfig(cfg domain.HeartsConfig)

	// state readers
	GetGameEndFlag() bool
	GetPhase() domain.HeartsPhase
	IsHumanTurn() bool
	GetRoundNumber() int
	GetTrickNumber() int
	GetCurrentPlayerIdx() int
	GetCurrentTrick() []*domain.HeartsTrickCard
	GetHeartsBroken() bool
	GetPassDirection() domain.HeartsPassDirection
	GetLeadPlayerIdx() int
	GetWinnerIdx() int
	GetPlayerCnt() int
	GetPlayer(i int) *domain.HeartsPlayer
	GetPassReady() [domain.HeartsPlayerCnt]bool
	GetPassedCards() [domain.HeartsPlayerCnt][]*domain.Card
	GetActionLog() []*domain.ActionLogEntry
}
