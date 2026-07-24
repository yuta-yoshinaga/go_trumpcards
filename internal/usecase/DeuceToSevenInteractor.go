//go:build !js || !wasm || casino

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// DeuceToSevenInteractorIF is the boundary interface for 2-7 Triple Draw use
// cases. The adapter layer depends on this, not the concrete type, so
// controllers can be tested with mock presenters.
type DeuceToSevenInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset initialises a fresh hand using the current config.
	Reset() string
	// ResetWithConfig swaps the config and initialises a fresh hand.
	// profileData, if non-empty, is imported as the meta-AI profile.
	ResetWithConfig(cfg domain.DeuceToSevenConfig, profileData []byte) string
	// GetConfig returns a copy of the current config.
	GetConfig() domain.DeuceToSevenConfig
	// Action executes a human betting action.
	Action(action, amount, humanPlayMs int) string
	// Exchange executes a draw with the given hand indices.
	Exchange(indices []int) string
	// Stand passes on the current draw.
	Stand() string
	// Hint returns the draw-phase hint output.
	Hint() string
	// ActionLog returns the log output for the current hand.
	ActionLog() string
}

// DeuceToSevenInteractor wires the DeuceToSeven domain game to a presenter.
type DeuceToSevenInteractor struct {
	GameBase[interfaces.DeuceToSevenGame]
	pp presenter.DeuceToSevenPresenter
}

// NewDeuceToSevenInteractor constructs the interactor. Panics if either
// dependency is nil (same contract as the other games; helps surface DI bugs
// at startup).
func NewDeuceToSevenInteractor(g interfaces.DeuceToSevenGame, pp presenter.DeuceToSevenPresenter) *DeuceToSevenInteractor {
	mustNotNil("DeuceToSevenInteractor", map[string]any{"g": g, "pp": pp})
	return &DeuceToSevenInteractor{
		GameBase: GameBase[interfaces.DeuceToSevenGame]{Game: g},
		pp:       pp,
	}
}

// Reset starts a fresh hand with the current config.
func (di *DeuceToSevenInteractor) Reset() string {
	return execAndPresent(di.Game, di.pp, di.Game.Reset)
}

// GetConfig returns the current game config.
func (di *DeuceToSevenInteractor) GetConfig() domain.DeuceToSevenConfig {
	return di.Game.GetConfig()
}

// ResetWithConfig validates the supplied config, applies it, and re-deals. An
// invalid config is reported as a presenter error without touching game state.
func (di *DeuceToSevenInteractor) ResetWithConfig(cfg domain.DeuceToSevenConfig, profileData []byte) string {
	if err := cfg.Validate(); err != nil {
		return di.pp.Output(di.Game, err)
	}
	di.Game.SetConfig(cfg)
	err := di.Game.Reset()
	if len(profileData) > 0 {
		_ = di.Game.ImportProfile(profileData)
	}
	return di.pp.Output(di.Game, err)
}

// Action performs a human betting action.
func (di *DeuceToSevenInteractor) Action(action, amount, humanPlayMs int) string {
	return execAndPresent(di.Game, di.pp, func() error {
		return di.Game.PlayerAction(action, amount, humanPlayMs)
	})
}

// Exchange performs a human draw with the given indices.
func (di *DeuceToSevenInteractor) Exchange(indices []int) string {
	return execAndPresent(di.Game, di.pp, func() error {
		return di.Game.PlayerExchange(indices)
	})
}

// Stand stands pat for the current draw.
func (di *DeuceToSevenInteractor) Stand() string {
	return execAndPresent(di.Game, di.pp, di.Game.PlayerStand)
}

// ActionLog returns the log output for the current hand.
func (di *DeuceToSevenInteractor) ActionLog() string {
	return di.pp.ActionLogOutput(di.Game)
}

// Hint returns the draw-phase hint output.
func (di *DeuceToSevenInteractor) Hint() string {
	return di.pp.HintOutput(di.Game)
}

// RestoreDeuceToSevenInteractor deserialises JSON into a DeuceToSevenInteractor.
// Used by the KV-backed worker session provider.
func RestoreDeuceToSevenInteractor(data []byte, pp presenter.DeuceToSevenPresenter) (*DeuceToSevenInteractor, error) {
	return restoreAndBuild[domain.DeuceToSeven](data, func(g *domain.DeuceToSeven) *DeuceToSevenInteractor {
		return &DeuceToSevenInteractor{
			GameBase: GameBase[interfaces.DeuceToSevenGame]{Game: g},
			pp:       pp,
		}
	})
}
