//go:build !js || !wasm || extra3

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// SkatInteractorIF Skat interactor interface.
type SkatInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset starts a fresh game.
	Reset() string
	// ResetWithConfig resets the game with a new configuration.
	ResetWithConfig(cfg domain.SkatConfig) string
	// Bid applies an accept/pass bid step from the human.
	Bid(accept bool) string
	// PickSkat applies the declarer's skat-pickup choice.
	PickSkat(pickup bool) string
	// Discard applies the declarer's discard choice.
	Discard(idxA, idxB int) string
	// DeclareGame applies the declarer's game-type choice.
	DeclareGame(gameType domain.SkatGameType, trumpSuit int) string
	// Play plays a card from the human's hand.
	Play(cardIndex int) string
	// NextTrick advances to the next trick.
	NextTrick() string
	// NextRound scores the round and advances to the next.
	NextRound() string
	// GetConfig returns the current configuration.
	GetConfig() domain.SkatConfig
	// Hint returns the hint for the human player.
	Hint() string
	// ActionLog returns the round's action log.
	ActionLog() string
}

// SkatInteractor Skat interactor implementation.
type SkatInteractor struct {
	GameBase[interfaces.SkatGame]
	sp presenter.SkatPresenter
}

// NewSkatInteractor constructs a SkatInteractor.
func NewSkatInteractor(s interfaces.SkatGame, sp presenter.SkatPresenter) *SkatInteractor {
	mustNotNil("SkatInteractor", map[string]any{"s": s, "sp": sp})
	return &SkatInteractor{GameBase: GameBase[interfaces.SkatGame]{Game: s}, sp: sp}
}

// Reset starts a fresh game and auto-runs CPU phases until the human is needed.
func (si *SkatInteractor) Reset() string {
	si.Game.Reset()
	si.runCpuAutoPhases()
	return si.sp.Output(si.Game, nil)
}

// ResetWithConfig resets with a new validated configuration.
func (si *SkatInteractor) ResetWithConfig(cfg domain.SkatConfig) string {
	return resetWithValidatedConfig(si.Game, si.sp, cfg, si.Game.SetConfig, si.Reset)
}

// Bid applies the accept/pass bid step.
func (si *SkatInteractor) Bid(accept bool) string {
	if out, blocked := guardGameEnd(si.Game, si.sp); blocked {
		return out
	}
	if err := si.Game.PlayerBid(accept); err != nil {
		return si.sp.Output(si.Game, err)
	}
	si.runCpuAutoPhases()
	return si.sp.Output(si.Game, nil)
}

// PickSkat applies the declarer's skat-pickup decision.
func (si *SkatInteractor) PickSkat(pickup bool) string {
	if out, blocked := guardGameEnd(si.Game, si.sp); blocked {
		return out
	}
	if err := si.Game.PlayerPickSkat(pickup); err != nil {
		return si.sp.Output(si.Game, err)
	}
	si.runCpuAutoPhases()
	return si.sp.Output(si.Game, nil)
}

// Discard applies the declarer's discard.
func (si *SkatInteractor) Discard(idxA, idxB int) string {
	if out, blocked := guardGameEnd(si.Game, si.sp); blocked {
		return out
	}
	if err := si.Game.PlayerDiscard(idxA, idxB); err != nil {
		return si.sp.Output(si.Game, err)
	}
	si.runCpuAutoPhases()
	return si.sp.Output(si.Game, nil)
}

// DeclareGame applies the declarer's game choice.
func (si *SkatInteractor) DeclareGame(gameType domain.SkatGameType, trumpSuit int) string {
	if out, blocked := guardGameEnd(si.Game, si.sp); blocked {
		return out
	}
	if err := si.Game.PlayerDeclareGame(gameType, trumpSuit); err != nil {
		return si.sp.Output(si.Game, err)
	}
	si.runCpuAutoPhases()
	return si.sp.Output(si.Game, nil)
}

// Play plays a card from the human's hand.
func (si *SkatInteractor) Play(cardIndex int) string {
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
func (si *SkatInteractor) NextTrick() string {
	if out, blocked := guardGameEnd(si.Game, si.sp); blocked {
		return out
	}
	si.Game.NextTrick()
	si.runCpuTurns()
	return si.sp.Output(si.Game, nil)
}

// NextRound scores the round and starts the next.
func (si *SkatInteractor) NextRound() string {
	si.Game.ScoreRound()
	if out, blocked := guardGameEnd(si.Game, si.sp); blocked {
		return out
	}
	si.Game.NextRound()
	si.runCpuAutoPhases()
	return si.sp.Output(si.Game, nil)
}

// GetConfig returns the current configuration.
func (si *SkatInteractor) GetConfig() domain.SkatConfig { return si.Game.GetConfig() }

// Hint returns the hint output.
func (si *SkatInteractor) Hint() string { return si.sp.HintOutput(si.Game) }

// ActionLog returns the action-log output.
func (si *SkatInteractor) ActionLog() string { return si.sp.ActionLogOutput(si.Game) }

// runCpuAutoPhases drives all CPU-only phases until the human is needed or
// the round ends.
func (si *SkatInteractor) runCpuAutoPhases() {
	si.runCpuBids()
	si.runCpuDeclarerPhases()
	if si.Game.GetPhase() == domain.SkatPhasePlay {
		si.runCpuTurns()
	}
}

// runCpuBids runs CPU bid steps until the human's turn or the bid resolves.
func (si *SkatInteractor) runCpuBids() {
	runCpuBidsLoop(si.Game, domain.SkatPhaseBid)
}

// runCpuDeclarerPhases drives skat-pickup, discard, and game-declaration when
// the declarer is a CPU.
func (si *SkatInteractor) runCpuDeclarerPhases() {
	for i := 0; i < MaxCpuIterations; i++ {
		if si.Game.GetGameEndFlag() {
			return
		}
		phase := si.Game.GetPhase()
		switch phase {
		case domain.SkatPhaseSkatPickup:
			if si.Game.IsHumanDeclarerTurn() {
				return
			}
			si.Game.CpuPickSkat()
		case domain.SkatPhaseDiscard:
			if si.Game.IsHumanDeclarerTurn() {
				return
			}
			si.Game.CpuDiscard()
		case domain.SkatPhaseGameDeclaration:
			if si.Game.IsHumanDeclarerTurn() {
				return
			}
			si.Game.CpuDeclareGame()
		default:
			return
		}
	}
}

// runCpuTurns runs CPU trick-play turns.
func (si *SkatInteractor) runCpuTurns() {
	runCpuTurnsLoop(si.Game, trickPhases[domain.SkatPhase]{
		play:     domain.SkatPhasePlay,
		trickEnd: domain.SkatPhaseTrickEnd,
		roundEnd: domain.SkatPhaseRoundEnd,
		gameEnd:  domain.SkatPhaseGameEnd,
	})
}

// RestoreSkatInteractor deserialises a SkatInteractor from JSON.
func RestoreSkatInteractor(data []byte, sp presenter.SkatPresenter) (*SkatInteractor, error) {
	return restoreAndBuild[domain.Skat](data, func(g *domain.Skat) *SkatInteractor {
		return &SkatInteractor{GameBase: GameBase[interfaces.SkatGame]{Game: g}, sp: sp}
	})
}
