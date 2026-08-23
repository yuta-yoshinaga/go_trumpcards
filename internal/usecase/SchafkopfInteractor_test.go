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

const skMockOutput = `{"phase":0}`

func TestNewSchafkopfInteractor_NilGuards(t *testing.T) {
	spMock := new(presenter.MockSchafkopfPresenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "SchafkopfInteractor: g must not be nil", func() {
			usecase.NewSchafkopfInteractor(nil, spMock)
		})
	})
	t.Run("panics when sp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockSchafkopfGame)
		assert.PanicsWithValue(t, "SchafkopfInteractor: sp must not be nil", func() {
			usecase.NewSchafkopfInteractor(gameMock, nil)
		})
	})
}

func TestSchafkopfInteractor_Reset(t *testing.T) {
	spMock := new(presenter.MockSchafkopfPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(skMockOutput)
	gameMock := new(interfaces.MockSchafkopfGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.SchafkopfPhasePick)
	gameMock.On("IsHumanTurn").Return(true)

	si := usecase.NewSchafkopfInteractor(gameMock, spMock)
	assert.Equal(t, skMockOutput, si.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestSchafkopfInteractor_ResetWithConfig(t *testing.T) {
	spMock := new(presenter.MockSchafkopfPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(skMockOutput)
	gameMock := new(interfaces.MockSchafkopfGame)
	cfg := domain.SchafkopfConfig{
		CpuDifficulty: domain.SchafkopfCpuDifficultyHard,
		BaseChips:     2,
		StartChips:    20,
		TargetChips:   60,
	}
	gameMock.On("SetConfig", cfg).Return()
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.SchafkopfPhasePick)
	gameMock.On("IsHumanTurn").Return(true)

	si := usecase.NewSchafkopfInteractor(gameMock, spMock)
	assert.Equal(t, skMockOutput, si.ResetWithConfig(cfg))
	gameMock.AssertCalled(t, "SetConfig", cfg)
}

func TestSchafkopfInteractor_ResetWithConfigInvalid(t *testing.T) {
	spMock := new(presenter.MockSchafkopfPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(skMockOutput)
	gameMock := new(interfaces.MockSchafkopfGame)

	si := usecase.NewSchafkopfInteractor(gameMock, spMock)
	// TargetChips must be > StartChips; StartChips=10, TargetChips=5 is invalid
	bad := domain.SchafkopfConfig{
		CpuDifficulty: domain.SchafkopfCpuDifficultyNormal,
		BaseChips:     2,
		StartChips:    10,
		TargetChips:   5,
	}
	assert.Equal(t, skMockOutput, si.ResetWithConfig(bad))
	gameMock.AssertNotCalled(t, "Reset")
}

func TestSchafkopfInteractor_Declare(t *testing.T) {
	spMock := new(presenter.MockSchafkopfPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(skMockOutput)

	t.Run("declare advances CPU", func(t *testing.T) {
		gameMock := new(interfaces.MockSchafkopfGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerDeclare", true, domain.SchafkopfContractRufspiel, 0).Return(nil)
		gameMock.On("GetPhase").Return(domain.SchafkopfPhaseCall)
		gameMock.On("IsHumanTurn").Return(true)

		si := usecase.NewSchafkopfInteractor(gameMock, spMock)
		assert.Equal(t, skMockOutput, si.Declare(true, domain.SchafkopfContractRufspiel, 0))
		gameMock.AssertCalled(t, "PlayerDeclare", true, domain.SchafkopfContractRufspiel, 0)
	})

	t.Run("declare forwards the contract and its solo suit", func(t *testing.T) {
		// **契約とスートは素通しでなければならない。**ここを落とすと、宣言した
		// 契約と盤面の切り札が別物になる。
		gameMock := new(interfaces.MockSchafkopfGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerDeclare", true, domain.SchafkopfContractSolo, domain.CardDesignHeart).Return(nil)
		gameMock.On("GetPhase").Return(domain.SchafkopfPhasePlay)
		gameMock.On("IsHumanTurn").Return(true)

		si := usecase.NewSchafkopfInteractor(gameMock, spMock)
		assert.Equal(t, skMockOutput, si.Declare(true, domain.SchafkopfContractSolo, domain.CardDesignHeart))
		gameMock.AssertCalled(t, "PlayerDeclare", true, domain.SchafkopfContractSolo, domain.CardDesignHeart)
	})

	t.Run("declare returns error", func(t *testing.T) {
		gameMock := new(interfaces.MockSchafkopfGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerDeclare", false, domain.SchafkopfContractRufspiel, 0).Return(errors.New("cannot pass"))

		si := usecase.NewSchafkopfInteractor(gameMock, spMock)
		assert.Equal(t, skMockOutput, si.Declare(false, domain.SchafkopfContractRufspiel, 0))
	})

	t.Run("game ended blocks declare", func(t *testing.T) {
		gameMock := new(interfaces.MockSchafkopfGame)
		gameMock.On("GetGameEndFlag").Return(true)

		si := usecase.NewSchafkopfInteractor(gameMock, spMock)
		assert.Equal(t, skMockOutput, si.Declare(true, domain.SchafkopfContractRufspiel, 0))
		gameMock.AssertNotCalled(t, "PlayerDeclare", mock.Anything, mock.Anything, mock.Anything)
	})
}

func TestSchafkopfInteractor_Call(t *testing.T) {
	spMock := new(presenter.MockSchafkopfPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(skMockOutput)

	t.Run("call success", func(t *testing.T) {
		gameMock := new(interfaces.MockSchafkopfGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerCall", domain.CardDesignClover).Return(nil)
		gameMock.On("GetPhase").Return(domain.SchafkopfPhasePlay)
		gameMock.On("IsHumanTurn").Return(true)

		si := usecase.NewSchafkopfInteractor(gameMock, spMock)
		assert.Equal(t, skMockOutput, si.Call(domain.CardDesignClover))
		gameMock.AssertCalled(t, "PlayerCall", domain.CardDesignClover)
	})

	t.Run("call error", func(t *testing.T) {
		gameMock := new(interfaces.MockSchafkopfGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerCall", 0).Return(errors.New("invalid suit"))

		si := usecase.NewSchafkopfInteractor(gameMock, spMock)
		assert.Equal(t, skMockOutput, si.Call(0))
	})

	t.Run("game ended blocks call", func(t *testing.T) {
		gameMock := new(interfaces.MockSchafkopfGame)
		gameMock.On("GetGameEndFlag").Return(true)

		si := usecase.NewSchafkopfInteractor(gameMock, spMock)
		assert.Equal(t, skMockOutput, si.Call(domain.CardDesignSpade))
		gameMock.AssertNotCalled(t, "PlayerCall", mock.Anything)
	})
}

func TestSchafkopfInteractor_PlayResolvesTrick(t *testing.T) {
	spMock := new(presenter.MockSchafkopfPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(skMockOutput)
	gameMock := new(interfaces.MockSchafkopfGame)
	gameMock.On("GetGameEndFlag").Return(false)
	// Phase sequence:
	//   1st call: PhasePlay  (guard check in Play())
	//   2nd call: TrickEnd   (post-PlayerPlay check → triggers ResolveTrick + NextTrick)
	//   3rd call: TrickEnd   (post-ResolveTrick check inside Play(), resolves to RoundEnd path... use RoundEnd to exit)
	//   Remaining: RoundEnd  (runCpuTurns exits on RoundEnd)
	gameMock.On("GetPhase").Return(domain.SchafkopfPhasePlay).Once()
	gameMock.On("GetPhase").Return(domain.SchafkopfPhaseTrickEnd).Once()
	gameMock.On("GetPhase").Return(domain.SchafkopfPhaseRoundEnd)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 2).Return(nil)
	gameMock.On("ResolveTrick").Return()

	si := usecase.NewSchafkopfInteractor(gameMock, spMock)
	assert.Equal(t, skMockOutput, si.Play(2))
	gameMock.AssertCalled(t, "PlayerPlay", 2)
	gameMock.AssertCalled(t, "ResolveTrick")
}

func TestSchafkopfInteractor_PlayNoResolveWhenNotTrickEnd(t *testing.T) {
	spMock := new(presenter.MockSchafkopfPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(skMockOutput)
	gameMock := new(interfaces.MockSchafkopfGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.SchafkopfPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 1).Return(nil)

	si := usecase.NewSchafkopfInteractor(gameMock, spMock)
	assert.Equal(t, skMockOutput, si.Play(1))
	gameMock.AssertNotCalled(t, "ResolveTrick")
}

func TestSchafkopfInteractor_PlayError(t *testing.T) {
	spMock := new(presenter.MockSchafkopfPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(skMockOutput)
	gameMock := new(interfaces.MockSchafkopfGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.SchafkopfPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 9).Return(errors.New("boom"))

	si := usecase.NewSchafkopfInteractor(gameMock, spMock)
	assert.Equal(t, skMockOutput, si.Play(9))
	gameMock.AssertNotCalled(t, "ResolveTrick")
}

func TestSchafkopfInteractor_PlayNotHumanTurn(t *testing.T) {
	spMock := new(presenter.MockSchafkopfPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(skMockOutput)
	gameMock := new(interfaces.MockSchafkopfGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.SchafkopfPhasePlay)
	gameMock.On("IsHumanTurn").Return(false)

	si := usecase.NewSchafkopfInteractor(gameMock, spMock)
	assert.Equal(t, skMockOutput, si.Play(0))
	gameMock.AssertNotCalled(t, "PlayerPlay", mock.Anything)
}

func TestSchafkopfInteractor_NextTrick(t *testing.T) {
	spMock := new(presenter.MockSchafkopfPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(skMockOutput)
	gameMock := new(interfaces.MockSchafkopfGame)
	gameMock.On("NextTrick").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.SchafkopfPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	si := usecase.NewSchafkopfInteractor(gameMock, spMock)
	assert.Equal(t, skMockOutput, si.NextTrick())
	gameMock.AssertCalled(t, "NextTrick")
}

func TestSchafkopfInteractor_NextRound(t *testing.T) {
	spMock := new(presenter.MockSchafkopfPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(skMockOutput)
	gameMock := new(interfaces.MockSchafkopfGame)
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("NextRound").Return()
	gameMock.On("GetPhase").Return(domain.SchafkopfPhasePick)
	gameMock.On("IsHumanTurn").Return(true)

	si := usecase.NewSchafkopfInteractor(gameMock, spMock)
	assert.Equal(t, skMockOutput, si.NextRound())
	gameMock.AssertCalled(t, "NextRound")
}

func TestSchafkopfInteractor_NextRoundGameEnded(t *testing.T) {
	spMock := new(presenter.MockSchafkopfPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(skMockOutput)
	gameMock := new(interfaces.MockSchafkopfGame)
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(true)

	si := usecase.NewSchafkopfInteractor(gameMock, spMock)
	assert.Equal(t, skMockOutput, si.NextRound())
	gameMock.AssertNotCalled(t, "NextRound")
}

func TestSchafkopfInteractor_GetConfigHintActionLog(t *testing.T) {
	spMock := new(presenter.MockSchafkopfPresenter)
	spMock.On("HintOutput", mock.Anything).Return("hint")
	spMock.On("ActionLogOutput", mock.Anything).Return("log")
	gameMock := new(interfaces.MockSchafkopfGame)
	cfg := domain.DefaultSchafkopfConfig()
	gameMock.On("GetConfig").Return(cfg)

	si := usecase.NewSchafkopfInteractor(gameMock, spMock)
	assert.Equal(t, cfg, si.GetConfig())
	assert.Equal(t, "hint", si.Hint())
	assert.Equal(t, "log", si.ActionLog())
}

func TestRestoreSchafkopfInteractor(t *testing.T) {
	spMock := new(presenter.MockSchafkopfPresenter)
	src := domain.NewDefaultSchafkopf()
	src.Reset()
	data, err := src.MarshalJSON()
	assert.NoError(t, err)

	si, err := usecase.RestoreSchafkopfInteractor(data, spMock)
	assert.NoError(t, err)
	assert.NotNil(t, si)

	_, err = usecase.RestoreSchafkopfInteractor([]byte(`{`), spMock)
	assert.Error(t, err)
}
