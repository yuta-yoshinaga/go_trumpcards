//go:build test

package usecase_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// gameEndedCariocaMocks builds mocks where the game has already ended so every
// playable command is short-circuited by the guard.
func gameEndedCariocaMocks() (*presenter.MockCariocaPresenter, *interfaces.MockCariocaGame) {
	pMock := new(presenter.MockCariocaPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(cariocaMockOutput)
	gameMock := new(interfaces.MockCariocaGame)
	gameMock.On("GetGameEndFlag").Return(true)
	return pMock, gameMock
}

// TestCariocaInteractor_GuardBlocksWhenGameEnded covers the blocked branch of
// guardNotPlayable for every playable command.
func TestCariocaInteractor_GuardBlocksWhenGameEnded(t *testing.T) {
	t.Run("DrawFromDiscard", func(t *testing.T) {
		p, g := gameEndedCariocaMocks()
		ci := usecase.NewCariocaInteractor(g, p)
		assert.Equal(t, cariocaMockOutput, ci.DrawFromDiscard())
		g.AssertNotCalled(t, "PlayerDrawFromDiscard")
	})
	t.Run("MeldContract", func(t *testing.T) {
		p, g := gameEndedCariocaMocks()
		ci := usecase.NewCariocaInteractor(g, p)
		assert.Equal(t, cariocaMockOutput, ci.MeldContract([][]int{{0, 1, 2}}))
		g.AssertNotCalled(t, "PlayerMeldContract", mock.Anything)
	})
	t.Run("MeldExtra", func(t *testing.T) {
		p, g := gameEndedCariocaMocks()
		ci := usecase.NewCariocaInteractor(g, p)
		assert.Equal(t, cariocaMockOutput, ci.MeldExtra([]int{0, 1, 2}))
		g.AssertNotCalled(t, "PlayerMeldExtra", mock.Anything)
	})
	t.Run("Layoff", func(t *testing.T) {
		p, g := gameEndedCariocaMocks()
		ci := usecase.NewCariocaInteractor(g, p)
		assert.Equal(t, cariocaMockOutput, ci.Layoff(1, 0, 2))
		g.AssertNotCalled(t, "PlayerLayoff", mock.Anything, mock.Anything, mock.Anything)
	})
	t.Run("Discard", func(t *testing.T) {
		p, g := gameEndedCariocaMocks()
		ci := usecase.NewCariocaInteractor(g, p)
		assert.Equal(t, cariocaMockOutput, ci.Discard(0))
		g.AssertNotCalled(t, "PlayerDiscard", mock.Anything)
	})
}

// TestCariocaInteractor_CommandErrorBranches covers the domain-error branch of
// each command (err != nil → Output(game, err), no CPU turns).
func TestCariocaInteractor_CommandErrorBranches(t *testing.T) {
	t.Run("DrawFromDiscard error", func(t *testing.T) {
		p, g := setupCariocaMocks()
		g.On("PlayerDrawFromDiscard").Return(errors.New("boom"))
		ci := usecase.NewCariocaInteractor(g, p)
		assert.Equal(t, cariocaMockOutput, ci.DrawFromDiscard())
		g.AssertCalled(t, "PlayerDrawFromDiscard")
	})
	t.Run("MeldExtra error", func(t *testing.T) {
		p, g := setupCariocaMocks()
		g.On("PlayerMeldExtra", mock.Anything).Return(errors.New("bad"))
		ci := usecase.NewCariocaInteractor(g, p)
		assert.Equal(t, cariocaMockOutput, ci.MeldExtra([]int{0, 1, 2}))
	})
	t.Run("Layoff error", func(t *testing.T) {
		p, g := setupCariocaMocks()
		g.On("PlayerLayoff", 1, 0, 2).Return(errors.New("bad"))
		ci := usecase.NewCariocaInteractor(g, p)
		assert.Equal(t, cariocaMockOutput, ci.Layoff(1, 0, 2))
	})
	t.Run("Discard error", func(t *testing.T) {
		p, g := setupCariocaMocks()
		g.On("PlayerDiscard", 0).Return(errors.New("bad"))
		ci := usecase.NewCariocaInteractor(g, p)
		assert.Equal(t, cariocaMockOutput, ci.Discard(0))
	})
}

// TestCariocaInteractor_RealGameDrivesCpuTurns exercises runCpuTurns and the
// success paths against a real domain game (Easy CPU), driving whole rounds so
// the CPU AI branches execute.
func TestCariocaInteractor_RealGameDrivesCpuTurns(t *testing.T) {
	g := domain.NewDefaultCarioca()
	cfg := g.GetConfig()
	cfg.CpuDifficulty = domain.CariocaCpuDifficultyEasy
	g.SetConfig(cfg)

	pMock := new(presenter.MockCariocaPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(cariocaMockOutput)
	pMock.On("ActionLogOutput", mock.Anything).Return("log")

	ci := usecase.NewCariocaInteractor(g, pMock)
	assert.Equal(t, cariocaMockOutput, ci.Reset())

	for i := 0; i < 6000 && !g.GetGameEndFlag(); i++ {
		switch g.GetPhase() {
		case domain.CariocaPhaseRoundEnd:
			ci.NextRound()
		case domain.CariocaPhaseDraw:
			if g.IsHumanTurn() {
				ci.DrawFromStock()
			} else {
				// Should not normally happen (runCpuTurns drains CPU turns),
				// but guard against a stuck loop.
				g.CpuPlay()
			}
		case domain.CariocaPhasePlay:
			if g.IsHumanTurn() && g.GetPlayer(0).GetCardsSize() > 0 {
				ci.Discard(0)
			} else {
				g.CpuPlay()
			}
		}
	}
	// Coverage-driven: the loop exercises every interactor command + runCpuTurns.
	// Random Easy play isn't guaranteed to complete a contract within the budget
	// (a turn is draw-1/discard-1 and an empty stock recycles), so assert a
	// consistent state rather than a hard game end.
	assert.GreaterOrEqual(t, g.GetPhase(), domain.CariocaPhaseDraw)
	assert.LessOrEqual(t, g.GetPhase(), domain.CariocaPhaseGameEnd)
	// ActionLog path.
	assert.Equal(t, "log", ci.ActionLog())
}

// TestCariocaInteractor_ResetWithConfig_EasyRunsCpu covers ResetWithConfig with a
// valid Easy-difficulty config against a real game (drives runCpuTurns via Reset).
func TestCariocaInteractor_ResetWithConfig_EasyRunsCpu(t *testing.T) {
	g := domain.NewDefaultCarioca()
	pMock := new(presenter.MockCariocaPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(cariocaMockOutput)
	ci := usecase.NewCariocaInteractor(g, pMock)

	cfg := domain.DefaultCariocaConfig()
	cfg.CpuDifficulty = domain.CariocaCpuDifficultyEasy
	cfg.PlayerCount = 5
	assert.Equal(t, cariocaMockOutput, ci.ResetWithConfig(cfg))
	assert.Equal(t, 5, g.GetPlayerCnt())
}
