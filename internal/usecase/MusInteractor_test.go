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

const musMockOutput = `{"phase":0}`

func TestNewMusInteractor_NilGuards(t *testing.T) {
	spMock := new(presenter.MockMusPresenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "MusInteractor: g must not be nil", func() {
			usecase.NewMusInteractor(nil, spMock)
		})
	})
	t.Run("panics when sp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockMusGame)
		assert.PanicsWithValue(t, "MusInteractor: sp must not be nil", func() {
			usecase.NewMusInteractor(gameMock, nil)
		})
	})
}

func TestMusInteractor_Reset(t *testing.T) {
	spMock := new(presenter.MockMusPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(musMockOutput)
	gameMock := new(interfaces.MockMusGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.MusPhaseMus)
	gameMock.On("IsHumanTurn").Return(true)

	mi := usecase.NewMusInteractor(gameMock, spMock)
	assert.Equal(t, musMockOutput, mi.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestMusInteractor_ResetWithConfig(t *testing.T) {
	spMock := new(presenter.MockMusPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(musMockOutput)
	gameMock := new(interfaces.MockMusGame)
	cfg := domain.MusConfig{
		CpuDifficulty:   domain.MusCpuDifficultyHard,
		TargetAmarrakos: 50,
	}
	gameMock.On("SetConfig", cfg).Return()
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.MusPhaseMus)
	gameMock.On("IsHumanTurn").Return(true)

	mi := usecase.NewMusInteractor(gameMock, spMock)
	assert.Equal(t, musMockOutput, mi.ResetWithConfig(cfg))
	gameMock.AssertCalled(t, "SetConfig", cfg)
}

func TestMusInteractor_ResetWithConfigInvalid(t *testing.T) {
	spMock := new(presenter.MockMusPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(musMockOutput)
	gameMock := new(interfaces.MockMusGame)

	mi := usecase.NewMusInteractor(gameMock, spMock)
	bad := domain.MusConfig{
		CpuDifficulty:   domain.MusCpuDifficulty(99),
		TargetAmarrakos: 40,
	}
	assert.Equal(t, musMockOutput, mi.ResetWithConfig(bad))
	gameMock.AssertNotCalled(t, "Reset")
}

func TestMusInteractor_Mus(t *testing.T) {
	spMock := new(presenter.MockMusPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(musMockOutput)

	t.Run("mus true advances CPU", func(t *testing.T) {
		gameMock := new(interfaces.MockMusGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerMus", true).Return(nil)
		gameMock.On("GetPhase").Return(domain.MusPhaseMus)
		gameMock.On("IsHumanTurn").Return(true)

		mi := usecase.NewMusInteractor(gameMock, spMock)
		assert.Equal(t, musMockOutput, mi.Mus(true))
		gameMock.AssertCalled(t, "PlayerMus", true)
	})

	t.Run("mus returns error", func(t *testing.T) {
		gameMock := new(interfaces.MockMusGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerMus", false).Return(errors.New("wrong phase"))

		mi := usecase.NewMusInteractor(gameMock, spMock)
		assert.Equal(t, musMockOutput, mi.Mus(false))
	})

	t.Run("game ended blocks mus", func(t *testing.T) {
		gameMock := new(interfaces.MockMusGame)
		gameMock.On("GetGameEndFlag").Return(true)

		mi := usecase.NewMusInteractor(gameMock, spMock)
		assert.Equal(t, musMockOutput, mi.Mus(true))
		gameMock.AssertNotCalled(t, "PlayerMus", mock.Anything)
	})
}

func TestMusInteractor_Discard(t *testing.T) {
	spMock := new(presenter.MockMusPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(musMockOutput)
	indices := []int{0, 2}

	t.Run("discard success", func(t *testing.T) {
		gameMock := new(interfaces.MockMusGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerDiscard", indices).Return(nil)
		gameMock.On("GetPhase").Return(domain.MusPhaseDiscard)
		gameMock.On("IsHumanTurn").Return(true)

		mi := usecase.NewMusInteractor(gameMock, spMock)
		assert.Equal(t, musMockOutput, mi.Discard(indices))
		gameMock.AssertCalled(t, "PlayerDiscard", indices)
	})

	t.Run("discard error", func(t *testing.T) {
		gameMock := new(interfaces.MockMusGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerDiscard", indices).Return(errors.New("invalid discard"))

		mi := usecase.NewMusInteractor(gameMock, spMock)
		assert.Equal(t, musMockOutput, mi.Discard(indices))
	})

	t.Run("game ended blocks discard", func(t *testing.T) {
		gameMock := new(interfaces.MockMusGame)
		gameMock.On("GetGameEndFlag").Return(true)

		mi := usecase.NewMusInteractor(gameMock, spMock)
		assert.Equal(t, musMockOutput, mi.Discard(indices))
		gameMock.AssertNotCalled(t, "PlayerDiscard", mock.Anything)
	})
}

func TestMusInteractor_Bet(t *testing.T) {
	spMock := new(presenter.MockMusPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(musMockOutput)

	t.Run("paso success", func(t *testing.T) {
		gameMock := new(interfaces.MockMusGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerBet", domain.MusActionPaso, 0).Return(nil)
		gameMock.On("GetPhase").Return(domain.MusPhaseGrande)
		gameMock.On("IsHumanTurn").Return(true)

		mi := usecase.NewMusInteractor(gameMock, spMock)
		assert.Equal(t, musMockOutput, mi.Bet(domain.MusActionPaso, 0))
		gameMock.AssertCalled(t, "PlayerBet", domain.MusActionPaso, 0)
	})

	t.Run("bet error", func(t *testing.T) {
		gameMock := new(interfaces.MockMusGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerBet", domain.MusActionEnvido, 2).Return(errors.New("invalid bet"))

		mi := usecase.NewMusInteractor(gameMock, spMock)
		assert.Equal(t, musMockOutput, mi.Bet(domain.MusActionEnvido, 2))
	})

	t.Run("game ended blocks bet", func(t *testing.T) {
		gameMock := new(interfaces.MockMusGame)
		gameMock.On("GetGameEndFlag").Return(true)

		mi := usecase.NewMusInteractor(gameMock, spMock)
		assert.Equal(t, musMockOutput, mi.Bet(domain.MusActionPaso, 0))
		gameMock.AssertNotCalled(t, "PlayerBet", mock.Anything, mock.Anything)
	})
}

func TestMusInteractor_NextRound(t *testing.T) {
	spMock := new(presenter.MockMusPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(musMockOutput)
	gameMock := new(interfaces.MockMusGame)
	gameMock.On("NextRound").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.MusPhaseMus)
	gameMock.On("IsHumanTurn").Return(true)

	mi := usecase.NewMusInteractor(gameMock, spMock)
	assert.Equal(t, musMockOutput, mi.NextRound())
	gameMock.AssertCalled(t, "NextRound")
}

func TestMusInteractor_RunCpuTurns_Showdown(t *testing.T) {
	spMock := new(presenter.MockMusPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(musMockOutput)
	gameMock := new(interfaces.MockMusGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("PlayerBet", domain.MusActionPaso, 0).Return(nil)
	// Phase sequence: bet phase (human turn) → Showdown → RoundEnd
	gameMock.On("GetPhase").Return(domain.MusPhaseShowdown).Once()
	gameMock.On("GetPhase").Return(domain.MusPhaseRoundEnd)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("Showdown").Return()

	mi := usecase.NewMusInteractor(gameMock, spMock)
	// Trigger runCpuTurns by calling Reset (which calls runCpuTurns after Reset)
	gameMock.On("Reset").Return()
	assert.Equal(t, musMockOutput, mi.Reset())
	gameMock.AssertCalled(t, "Showdown")
}

func TestMusInteractor_GetConfigHintActionLog(t *testing.T) {
	spMock := new(presenter.MockMusPresenter)
	spMock.On("HintOutput", mock.Anything).Return("hint")
	spMock.On("ActionLogOutput", mock.Anything).Return("log")
	gameMock := new(interfaces.MockMusGame)
	cfg := domain.DefaultMusConfig()
	gameMock.On("GetConfig").Return(cfg)

	mi := usecase.NewMusInteractor(gameMock, spMock)
	assert.Equal(t, cfg, mi.GetConfig())
	assert.Equal(t, "hint", mi.Hint())
	assert.Equal(t, "log", mi.ActionLog())
}

func TestRestoreMusInteractor(t *testing.T) {
	spMock := new(presenter.MockMusPresenter)
	src := domain.NewDefaultMus()
	src.Reset()
	data, err := src.MarshalJSON()
	assert.NoError(t, err)

	mi, err := usecase.RestoreMusInteractor(data, spMock)
	assert.NoError(t, err)
	assert.NotNil(t, mi)

	_, err = usecase.RestoreMusInteractor([]byte(`{`), spMock)
	assert.Error(t, err)
}
