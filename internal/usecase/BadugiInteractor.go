//go:build !js || !wasm || casino

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// BadugiInteractorIF is the boundary interface for Badugi use cases. The
// adapter layer depends on this, not the concrete type, so controllers can
// be tested with mock presenters.
type BadugiInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset initialises a fresh hand using the current config.
	Reset() string
	// ResetWithConfig swaps the config and initialises a fresh hand.
	// profileData, if non-empty, is imported as the meta-AI profile.
	ResetWithConfig(cfg domain.BadugiConfig, profileData []byte) string
	// GetConfig returns a copy of the current config.
	GetConfig() domain.BadugiConfig
	// Action executes a human betting action.
	Action(action, amount, humanPlayMs int) string
	// Exchange executes a draw with the given hand indices.
	Exchange(indices []int, humanPlayMs int) string
	// Stand passes on the current draw.
	Stand(humanPlayMs int) string
	// Hint returns the hint output for the current hand.
	Hint() string
	// ActionLog returns the log output for the current hand.
	ActionLog() string
}

// BadugiInteractor wires the Badugi domain game to a presenter.
type BadugiInteractor struct {
	GameBase[interfaces.BadugiGame]
	pp presenter.BadugiPresenter
}

// NewBadugiInteractor constructs the interactor. Panics if either dependency
// is nil (same contract as the other games; helps surface DI bugs at startup).
func NewBadugiInteractor(g interfaces.BadugiGame, pp presenter.BadugiPresenter) *BadugiInteractor {
	mustNotNil("BadugiInteractor", map[string]any{"g": g, "pp": pp})
	return &BadugiInteractor{
		GameBase: GameBase[interfaces.BadugiGame]{Game: g},
		pp:       pp,
	}
}

// Reset starts a fresh hand with the current config.
func (bi *BadugiInteractor) Reset() string {
	return execAndPresent(bi.Game, bi.pp, bi.Game.Reset)
}

// GetConfig returns the current game config.
func (bi *BadugiInteractor) GetConfig() domain.BadugiConfig {
	return bi.Game.GetConfig()
}

// ResetWithConfig validates the supplied config, applies it, and re-deals.
// An invalid config is reported as a presenter error without touching game state.
func (bi *BadugiInteractor) ResetWithConfig(cfg domain.BadugiConfig, profileData []byte) string {
	if err := cfg.Validate(); err != nil {
		return bi.pp.Output(bi.Game, err)
	}
	bi.Game.SetConfig(cfg)
	err := bi.Game.Reset()
	if len(profileData) > 0 {
		_ = bi.Game.ImportProfile(profileData)
	}
	return bi.pp.Output(bi.Game, err)
}

// Action performs a human betting action.
func (bi *BadugiInteractor) Action(action, amount, humanPlayMs int) string {
	return execAndPresent(bi.Game, bi.pp, func() error {
		return bi.Game.PlayerAction(action, amount, humanPlayMs)
	})
}

// Exchange performs a human draw with the given indices.
func (bi *BadugiInteractor) Exchange(indices []int, humanPlayMs int) string {
	return execAndPresent(bi.Game, bi.pp, func() error {
		return bi.Game.PlayerExchange(indices, humanPlayMs)
	})
}

// Stand stands pat for the current draw.
func (bi *BadugiInteractor) Stand(humanPlayMs int) string {
	return execAndPresent(bi.Game, bi.pp, func() error {
		return bi.Game.PlayerStand(humanPlayMs)
	})
}

// ActionLog returns the log output for the current hand.
func (bi *BadugiInteractor) ActionLog() string {
	return bi.pp.ActionLogOutput(bi.Game)
}

// Hint returns the hint output for the current hand.
func (bi *BadugiInteractor) Hint() string {
	return bi.pp.HintOutput(bi.Game)
}

// RestoreBadugiInteractor deserialises JSON into a BadugiInteractor. Used by
// the KV-backed worker session provider.
func RestoreBadugiInteractor(data []byte, pp presenter.BadugiPresenter) (*BadugiInteractor, error) {
	return restoreAndBuild[domain.Badugi](data, func(g *domain.Badugi) *BadugiInteractor {
		return &BadugiInteractor{
			GameBase: GameBase[interfaces.BadugiGame]{Game: g},
			pp:       pp,
		}
	})
}
