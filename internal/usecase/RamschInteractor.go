//go:build !js || !wasm || extra3

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// RamschInteractorIF Ramsch interactor interface.
type RamschInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset starts a fresh game.
	Reset() string
	// ResetWithConfig resets the game with a new configuration.
	ResetWithConfig(cfg domain.RamschConfig) string
	// Play plays a card from the human's hand.
	Play(cardIndex int) string
	// NextTrick advances to the next trick.
	NextTrick() string
	// NextRound scores the round and advances to the next.
	NextRound() string
	// GetConfig returns the current configuration.
	GetConfig() domain.RamschConfig
	// Hint returns the hint for the human player.
	Hint() string
	// ActionLog returns the round's action log.
	ActionLog() string
}

// RamschInteractor Ramsch interactor implementation.
type RamschInteractor struct {
	GameBase[interfaces.RamschGame]
	sp presenter.RamschPresenter
}

// NewRamschInteractor constructs a RamschInteractor.
func NewRamschInteractor(s interfaces.RamschGame, sp presenter.RamschPresenter) *RamschInteractor {
	mustNotNil("RamschInteractor", map[string]any{"s": s, "sp": sp})
	return &RamschInteractor{GameBase: GameBase[interfaces.RamschGame]{Game: s}, sp: sp}
}

// Reset starts a fresh game and auto-runs CPU phases until the human is needed.
func (si *RamschInteractor) Reset() string {
	si.Game.Reset()
	si.runCpuAutoPhases()
	return si.sp.Output(si.Game, nil)
}

// ResetWithConfig resets with a new validated configuration.
func (si *RamschInteractor) ResetWithConfig(cfg domain.RamschConfig) string {
	return resetWithValidatedConfig(si.Game, si.sp, cfg, si.Game.SetConfig, si.Reset)
}

// Play plays a card from the human's hand.
func (si *RamschInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(si.Game, si.sp); blocked {
		return out
	}
	if err := si.Game.PlayerPlay(cardIndex); err != nil {
		return si.sp.Output(si.Game, err)
	}
	si.runCpuTurns()
	return si.sp.Output(si.Game, nil)
}

// NextTrick advances to the next trick and runs CPU turns.
func (si *RamschInteractor) NextTrick() string {
	if out, blocked := guardGameEnd(si.Game, si.sp); blocked {
		return out
	}
	si.Game.NextTrick()
	si.runCpuTurns()
	return si.sp.Output(si.Game, nil)
}

// NextRound scores the round and starts the next.
func (si *RamschInteractor) NextRound() string {
	si.Game.ScoreRound()
	if out, blocked := guardGameEnd(si.Game, si.sp); blocked {
		return out
	}
	si.Game.NextRound()
	si.runCpuAutoPhases()
	return si.sp.Output(si.Game, nil)
}

// GetConfig returns the current configuration.
func (si *RamschInteractor) GetConfig() domain.RamschConfig { return si.Game.GetConfig() }

// Hint returns the hint output.
func (si *RamschInteractor) Hint() string { return si.sp.HintOutput(si.Game) }

// ActionLog returns the action-log output.
func (si *RamschInteractor) ActionLog() string { return si.sp.ActionLogOutput(si.Game) }

// runCpuAutoPhases drives the CPU turns until the human is needed or the round
// ends.
//
// **入札も宣言も無いので、動かすのはトリックだけ。** Skat から来た
// runCpuBids / runCpuDeclarerPhases はこのゲームには相当する局面が無い。
func (si *RamschInteractor) runCpuAutoPhases() {
	if si.Game.GetPhase() == domain.RamschPhasePlay {
		si.runCpuTurns()
	}
}

// runCpuTurns runs CPU trick-play turns.
func (si *RamschInteractor) runCpuTurns() {
	runCpuTurnsLoop(si.Game, trickPhases[domain.RamschPhase]{
		play:     domain.RamschPhasePlay,
		trickEnd: domain.RamschPhaseTrickEnd,
		roundEnd: domain.RamschPhaseRoundEnd,
		gameEnd:  domain.RamschPhaseGameEnd,
	})
}

// RestoreRamschInteractor deserialises a RamschInteractor from JSON.
func RestoreRamschInteractor(data []byte, sp presenter.RamschPresenter) (*RamschInteractor, error) {
	return restoreAndBuild[domain.Ramsch](data, func(g *domain.Ramsch) *RamschInteractor {
		return &RamschInteractor{GameBase: GameBase[interfaces.RamschGame]{Game: g}, sp: sp}
	})
}
