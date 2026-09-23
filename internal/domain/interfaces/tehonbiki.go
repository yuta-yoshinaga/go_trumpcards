//go:build !js || !wasm || extra2

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// TehonbikiGame is the use-case boundary for 手本引き.
type TehonbikiGame interface {
	BaseGame
	Reset()
	PlaceBet([]int, domain.TehonbikiBetType, int) error
	NextRound() error
	GetConfig() domain.TehonbikiConfig
	SetConfig(domain.TehonbikiConfig)
	GetPhase() domain.TehonbikiPhase
	GetGameEndFlag() bool
	GetParentCard() int
	GetNumbers() []int
	GetBetType() domain.TehonbikiBetType
	GetBet() int
	GetResult() domain.TehonbikiResult
	GetPayout() int
	GetChips() int
	GetRoundNumber() int
	GetRemainingCards() int
	GetHint() *domain.TehonbikiHint
}
