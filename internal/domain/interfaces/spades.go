package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// SpadesGame スペードゲームインタフェース
type SpadesGame interface {
	// interactor が呼ぶメソッド
	Reset()
	NextRound()
	PlayerBid(bid int) error
	CpuBid()
	PlayerPlay(cardIndex int) error
	CpuPlay()
	ResolveTrick()
	NextTrick()
	ScoreRound()

	// config
	GetConfig() domain.SpadesConfig
	SetConfig(cfg domain.SpadesConfig)

	// state readers
	GetGameEndFlag() bool
	GetPhase() domain.SpadesPhase
	IsHumanTurn() bool
	IsHumanBidTurn() bool
	GetRoundNumber() int
	GetTrickNumber() int
	GetCurrentPlayerIdx() int
	GetCurrentTrick() []*domain.SpadesTrickCard
	GetSpadesBroken() bool
	GetLeadPlayerIdx() int
	GetBidPlayerIdx() int
	GetWinnerIdx() int
	GetPlayerCnt() int
	GetPlayer(i int) *domain.SpadesPlayer
	GetActionLog() []*domain.ActionLogEntry
}
