//go:build !js || !wasm || extra4

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// BassetGame is the interface exposed by the Basset domain to adapters.
type BassetGame interface {
	BaseGame
	Reset()
	NextRound()
	PlayerPlaceBet(rank, amount int) error
	PlayerDealTurn() error
	PlayerTakeWinnings() error
	PlayerPressParoli() error
	GetPhase() int
	GetGameEndFlag() bool
	GetChips() int
	GetTurnsPlayed() int
	GetTurnsTotal() int
	GetRemainingCount() int
	GetRemainingByRank() [domain.BassetMaxRank + 1]int
	GetBet() (*domain.BassetBet, int)
	GetLastTurn() *domain.BassetTurnResult
	GetTotalPayout() int
}
