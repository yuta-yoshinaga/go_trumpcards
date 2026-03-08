package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// HoldemGame テキサスホールデムゲームインタフェース
type HoldemGame interface {
	Reset() error
	PlayerAction(action, amount int) error
	GetPhase() int
	GetPlayers() []*domain.HoldemPlayer
	GetPlayer(i int) *domain.HoldemPlayer
	GetPlayerCnt() int
	GetCommunityCards() []*domain.Card
	GetPot() int
	GetSidePots() []domain.HoldemSidePot
	GetDealerIdx() int
	GetCurrentTurn() int
	GetGameEndFlag() bool
	GetLastBet() int
	GetMinRaise() int
	GetRaiseCount() int
	GetRoundResults() []domain.HoldemResult
	GetCpuActions() []domain.HoldemCpuAction
	GetConfig() domain.HoldemConfig
	SetConfig(cfg domain.HoldemConfig)
	IsHumanTurn() bool
	GetActedFlags() []bool
	GetHandCount() int
	Resize(players []*domain.HoldemPlayer)
}
