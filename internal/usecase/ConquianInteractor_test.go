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

const cqMockOut = `{"phase":0}`

func cqNewPresenterMock() *presenter.MockConquianPresenter {
	p := new(presenter.MockConquianPresenter)
	p.On("Output", mock.Anything, mock.Anything).Return(cqMockOut)
	return p
}

// cqGameMockPlayable wires the minimal getters so guardNotPlayable + runCpuTurns
// treat the game as a live human turn.
func cqGameMockPlayable() *interfaces.MockConquianGame {
	m := new(interfaces.MockConquianGame)
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.ConquianPhaseDraw)
	m.On("IsHumanTurn").Return(true)
	return m
}

func TestNewConquianInteractor_NilGuards(t *testing.T) {
	pMock := cqNewPresenterMock()
	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "ConquianInteractor: g must not be nil", func() {
			usecase.NewConquianInteractor(nil, pMock)
		})
	})
	t.Run("panics when gp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockConquianGame)
		assert.PanicsWithValue(t, "ConquianInteractor: gp must not be nil", func() {
			usecase.NewConquianInteractor(gameMock, nil)
		})
	})
}

func TestConquianInteractor_Reset(t *testing.T) {
	pMock := cqNewPresenterMock()
	gameMock := cqGameMockPlayable()
	gameMock.On("Reset").Return()
	ci := usecase.NewConquianInteractor(gameMock, pMock)
	assert.Equal(t, cqMockOut, ci.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestConquianInteractor_ResetWithConfig(t *testing.T) {
	t.Run("valid config sets then resets", func(t *testing.T) {
		pMock := cqNewPresenterMock()
		gameMock := cqGameMockPlayable()
		gameMock.On("Reset").Return()
		gameMock.On("SetConfig", mock.Anything).Return()
		ci := usecase.NewConquianInteractor(gameMock, pMock)
		out := ci.ResetWithConfig(domain.DefaultConquianConfig())
		assert.Equal(t, cqMockOut, out)
		gameMock.AssertCalled(t, "SetConfig", domain.DefaultConquianConfig())
	})
	t.Run("invalid config returns error output without reset", func(t *testing.T) {
		pMock := cqNewPresenterMock()
		gameMock := new(interfaces.MockConquianGame)
		ci := usecase.NewConquianInteractor(gameMock, pMock)
		bad := domain.ConquianConfig{CpuDifficulty: 99, TargetWins: 1}
		out := ci.ResetWithConfig(bad)
		assert.Equal(t, cqMockOut, out)
		gameMock.AssertNotCalled(t, "Reset")
	})
}

func TestConquianInteractor_Actions(t *testing.T) {
	t.Run("DrawFromStock success", func(t *testing.T) {
		pMock := cqNewPresenterMock()
		gameMock := cqGameMockPlayable()
		gameMock.On("PlayerDrawFromStock").Return(nil)
		ci := usecase.NewConquianInteractor(gameMock, pMock)
		assert.Equal(t, cqMockOut, ci.DrawFromStock())
		gameMock.AssertCalled(t, "PlayerDrawFromStock")
	})
	t.Run("DrawFromStock error", func(t *testing.T) {
		pMock := cqNewPresenterMock()
		gameMock := cqGameMockPlayable()
		gameMock.On("PlayerDrawFromStock").Return(errors.New("boom"))
		ci := usecase.NewConquianInteractor(gameMock, pMock)
		assert.Equal(t, cqMockOut, ci.DrawFromStock())
	})
	t.Run("DrawFromDiscard success", func(t *testing.T) {
		pMock := cqNewPresenterMock()
		gameMock := cqGameMockPlayable()
		gameMock.On("PlayerDrawFromDiscard").Return(nil)
		ci := usecase.NewConquianInteractor(gameMock, pMock)
		assert.Equal(t, cqMockOut, ci.DrawFromDiscard())
		gameMock.AssertCalled(t, "PlayerDrawFromDiscard")
	})
	t.Run("DrawFromDiscard error", func(t *testing.T) {
		pMock := cqNewPresenterMock()
		gameMock := cqGameMockPlayable()
		gameMock.On("PlayerDrawFromDiscard").Return(errors.New("boom"))
		ci := usecase.NewConquianInteractor(gameMock, pMock)
		assert.Equal(t, cqMockOut, ci.DrawFromDiscard())
	})
	t.Run("Meld success", func(t *testing.T) {
		pMock := cqNewPresenterMock()
		gameMock := cqGameMockPlayable()
		gameMock.On("PlayerMeldWithTargets", mock.Anything, mock.Anything).Return(nil)
		ci := usecase.NewConquianInteractor(gameMock, pMock)
		assert.Equal(t, cqMockOut, ci.Meld([][]int{{0, 1, 2}}))
		gameMock.AssertCalled(t, "PlayerMeldWithTargets", [][]int{{0, 1, 2}}, []int(nil))
	})
	t.Run("Meld error", func(t *testing.T) {
		pMock := cqNewPresenterMock()
		gameMock := cqGameMockPlayable()
		gameMock.On("PlayerMeldWithTargets", mock.Anything, mock.Anything).Return(errors.New("boom"))
		ci := usecase.NewConquianInteractor(gameMock, pMock)
		assert.Equal(t, cqMockOut, ci.Meld(nil))
	})
	t.Run("Discard success", func(t *testing.T) {
		pMock := cqNewPresenterMock()
		gameMock := cqGameMockPlayable()
		gameMock.On("PlayerDiscard", 2).Return(nil)
		ci := usecase.NewConquianInteractor(gameMock, pMock)
		assert.Equal(t, cqMockOut, ci.Discard(2))
		gameMock.AssertCalled(t, "PlayerDiscard", 2)
	})
	t.Run("Discard error", func(t *testing.T) {
		pMock := cqNewPresenterMock()
		gameMock := cqGameMockPlayable()
		gameMock.On("PlayerDiscard", 0).Return(errors.New("boom"))
		ci := usecase.NewConquianInteractor(gameMock, pMock)
		assert.Equal(t, cqMockOut, ci.Discard(0))
	})
}

func TestConquianInteractor_NextRound(t *testing.T) {
	pMock := cqNewPresenterMock()
	gameMock := cqGameMockPlayable()
	gameMock.On("NextRound").Return()
	ci := usecase.NewConquianInteractor(gameMock, pMock)
	assert.Equal(t, cqMockOut, ci.NextRound())
	gameMock.AssertCalled(t, "NextRound")
}

func TestConquianInteractor_NextRound_GameEnded(t *testing.T) {
	pMock := cqNewPresenterMock()
	gameMock := new(interfaces.MockConquianGame)
	gameMock.On("GetGameEndFlag").Return(true)
	ci := usecase.NewConquianInteractor(gameMock, pMock)
	assert.Equal(t, cqMockOut, ci.NextRound())
	gameMock.AssertNotCalled(t, "NextRound")
}

func TestConquianInteractor_GameEndedGuard(t *testing.T) {
	pMock := cqNewPresenterMock()
	gameMock := new(interfaces.MockConquianGame)
	gameMock.On("GetGameEndFlag").Return(true)
	ci := usecase.NewConquianInteractor(gameMock, pMock)
	assert.Equal(t, cqMockOut, ci.DrawFromStock())
	gameMock.AssertNotCalled(t, "PlayerDrawFromStock")
}

func TestConquianInteractor_GetConfig(t *testing.T) {
	pMock := cqNewPresenterMock()
	gameMock := new(interfaces.MockConquianGame)
	gameMock.On("GetConfig").Return(domain.DefaultConquianConfig())
	ci := usecase.NewConquianInteractor(gameMock, pMock)
	assert.Equal(t, domain.DefaultConquianConfig(), ci.GetConfig())
}

func TestConquianInteractor_ActionLog(t *testing.T) {
	pMock := new(presenter.MockConquianPresenter)
	pMock.On("ActionLogOutput", mock.Anything).Return("log-out")
	gameMock := new(interfaces.MockConquianGame)
	ci := usecase.NewConquianInteractor(gameMock, pMock)
	assert.Equal(t, "log-out", ci.ActionLog())
}

func TestConquianInteractor_RunCpuTurns(t *testing.T) {
	pMock := cqNewPresenterMock()
	gameMock := new(interfaces.MockConquianGame)
	gameMock.On("Reset").Return()
	// 1st loop: not ended, meld phase, CPU turn → CpuPlay; then becomes human turn.
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.ConquianPhaseMeld)
	gameMock.On("IsHumanTurn").Return(false).Once()
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("CpuPlay").Return()
	ci := usecase.NewConquianInteractor(gameMock, pMock)
	ci.Reset()
	gameMock.AssertCalled(t, "CpuPlay")
}

func TestRestoreConquianInteractor(t *testing.T) {
	g := domain.NewDefaultConquian()
	g.Reset()
	data, err := g.MarshalJSON()
	assert.NoError(t, err)
	pMock := cqNewPresenterMock()
	ci, err := usecase.RestoreConquianInteractor(data, pMock)
	assert.NoError(t, err)
	assert.NotNil(t, ci)
}

// **延長先の指定がドメインまで届くこと (#4837)。**Web から来た extendTargets を
// 捨てていると、プレイヤーがどのメルドに足すかを選べないままになる。
func TestConquianInteractor_MeldWithTargets(t *testing.T) {
	pMock := cqNewPresenterMock()
	gameMock := cqGameMockPlayable()
	gameMock.On("PlayerMeldWithTargets", mock.Anything, mock.Anything).Return(nil)
	ci := usecase.NewConquianInteractor(gameMock, pMock)

	assert.Equal(t, cqMockOut, ci.MeldWithTargets([][]int{{0}}, []int{1}))
	gameMock.AssertCalled(t, "PlayerMeldWithTargets", [][]int{{0}}, []int{1})
}
