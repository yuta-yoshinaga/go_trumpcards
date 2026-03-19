package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// OmahaGame オマハホールデムゲームインタフェース
type OmahaGame interface {
	Reset() error
	PlayerAction(action, amount int) error
	GetPhase() int
	GetPlayers() []*domain.OmahaPlayer
	GetPlayer(i int) *domain.OmahaPlayer
	GetPlayerCnt() int
	GetCommunityCards() []*domain.Card
	GetPot() int
	GetSidePots() []domain.OmahaSidePot
	GetDealerIdx() int
	GetCurrentTurn() int
	GetGameEndFlag() bool
	GetLastBet() int
	GetMinRaise() int
	GetRaiseCount() int
	GetRoundResults() []domain.OmahaResult
	GetCpuActions() []domain.OmahaCpuAction
	GetConfig() domain.OmahaConfig
	SetConfig(cfg domain.OmahaConfig)
	IsHumanTurn() bool
	GetActedFlags() []bool
	GetHandCount() int
	Resize(players []*domain.OmahaPlayer)
	Rebuy() error
	SkipRebuy() error
	Addon() error
	SkipAddon() error
	IsRebuyAvailable() bool
	IsAddonAvailable() bool
	GetRebuyCounts() []int
	GetAddonUsed() []bool
	GetRebuyPhaseType() int
	Muck() error
	ShowHand() error
	IsMuckAvailable() bool
	GetActionLog() []*domain.ActionLogEntry
	GetEquity() *domain.HoldemEquityResult
	GetPotOdds() float64
}
