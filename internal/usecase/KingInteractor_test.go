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

const kingMockOutput = `{"phase":"play"}`

func newKingMocks() (*interfaces.MockKingGame, *presenter.MockKingPresenter) {
	return new(interfaces.MockKingGame), new(presenter.MockKingPresenter)
}

func TestNewKingInteractor_NilGuards(t *testing.T) {
	kpMock := new(presenter.MockKingPresenter)
	t.Run("panics when kg is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "KingInteractor: kg must not be nil", func() {
			usecase.NewKingInteractor(nil, kpMock)
		})
	})
	t.Run("panics when kp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockKingGame)
		assert.PanicsWithValue(t, "KingInteractor: kp must not be nil", func() {
			usecase.NewKingInteractor(gameMock, nil)
		})
	})
}

func TestKingInteractor_Reset(t *testing.T) {
	gm, kp := newKingMocks()
	kp.On("Output", mock.Anything, mock.Anything).Return(kingMockOutput)
	gm.On("Reset").Return()
	gm.On("GetGameEndFlag").Return(false)
	gm.On("IsHumanTurn").Return(true)

	ki := usecase.NewKingInteractor(gm, kp)
	assert.Equal(t, kingMockOutput, ki.Reset())
	gm.AssertCalled(t, "Reset")
}

func TestKingInteractor_ResetWithConfig(t *testing.T) {
	gm, kp := newKingMocks()
	kp.On("Output", mock.Anything, mock.Anything).Return(kingMockOutput)
	cfg := domain.KingConfig{CpuDifficulty: domain.KingDifficultyHard}
	gm.On("SetConfig", cfg).Return()
	gm.On("Reset").Return()
	gm.On("GetGameEndFlag").Return(false)
	gm.On("IsHumanTurn").Return(true)

	ki := usecase.NewKingInteractor(gm, kp)
	assert.Equal(t, kingMockOutput, ki.ResetWithConfig(cfg))
	gm.AssertCalled(t, "SetConfig", cfg)
}

func TestKingInteractor_ResetWithConfigInvalid(t *testing.T) {
	gm, kp := newKingMocks()
	kp.On("Output", mock.Anything, mock.Anything).Return(kingMockOutput)
	ki := usecase.NewKingInteractor(gm, kp)
	// Invalid difficulty triggers validation failure -> Output with error, no Reset.
	out := ki.ResetWithConfig(domain.KingConfig{CpuDifficulty: 99})
	assert.Equal(t, kingMockOutput, out)
	gm.AssertNotCalled(t, "Reset")
}

func TestKingInteractor_NextDeal(t *testing.T) {
	gm, kp := newKingMocks()
	kp.On("Output", mock.Anything, mock.Anything).Return(kingMockOutput)
	gm.On("GetGameEndFlag").Return(false)
	gm.On("NextDeal").Return()
	gm.On("IsHumanTurn").Return(true)

	ki := usecase.NewKingInteractor(gm, kp)
	assert.Equal(t, kingMockOutput, ki.NextDeal())
	gm.AssertCalled(t, "NextDeal")
}

func TestKingInteractor_NextDeal_GameEnded(t *testing.T) {
	gm, kp := newKingMocks()
	kp.On("Output", mock.Anything, mock.Anything).Return(kingMockOutput)
	gm.On("GetGameEndFlag").Return(true)

	ki := usecase.NewKingInteractor(gm, kp)
	assert.Equal(t, kingMockOutput, ki.NextDeal())
	gm.AssertNotCalled(t, "NextDeal")
}

func TestKingInteractor_SelectContract_Success(t *testing.T) {
	gm, kp := newKingMocks()
	kp.On("Output", mock.Anything, mock.Anything).Return(kingMockOutput)
	gm.On("GetGameEndFlag").Return(false)
	gm.On("SelectContract", domain.KingContractKingTrump, domain.CardDesignSpade).Return(nil)
	gm.On("IsHumanTurn").Return(true)

	ki := usecase.NewKingInteractor(gm, kp)
	assert.Equal(t, kingMockOutput, ki.SelectContract(domain.KingContractKingTrump, domain.CardDesignSpade))
	gm.AssertCalled(t, "SelectContract", domain.KingContractKingTrump, domain.CardDesignSpade)
}

func TestKingInteractor_SelectContract_Error(t *testing.T) {
	gm, kp := newKingMocks()
	kp.On("Output", mock.Anything, mock.Anything).Return(kingMockOutput)
	gm.On("GetGameEndFlag").Return(false)
	gm.On("SelectContract", 0, -1).Return(errors.New("bad"))

	ki := usecase.NewKingInteractor(gm, kp)
	assert.Equal(t, kingMockOutput, ki.SelectContract(0, -1))
	// runCpuTurns should not be invoked after an error.
	gm.AssertNotCalled(t, "CpuPlay")
}

func TestKingInteractor_SelectContract_GameEnded(t *testing.T) {
	gm, kp := newKingMocks()
	kp.On("Output", mock.Anything, mock.Anything).Return(kingMockOutput)
	gm.On("GetGameEndFlag").Return(true)

	ki := usecase.NewKingInteractor(gm, kp)
	assert.Equal(t, kingMockOutput, ki.SelectContract(0, -1))
	gm.AssertNotCalled(t, "SelectContract", mock.Anything, mock.Anything)
}

func TestKingInteractor_Play_Success(t *testing.T) {
	gm, kp := newKingMocks()
	kp.On("Output", mock.Anything, mock.Anything).Return(kingMockOutput)
	// guardNotPlayable: game not ended + human turn.
	gm.On("GetGameEndFlag").Return(false)
	gm.On("IsHumanTurn").Return(true)
	gm.On("PlayerPlay", 3).Return(nil)

	ki := usecase.NewKingInteractor(gm, kp)
	assert.Equal(t, kingMockOutput, ki.Play(3))
	gm.AssertCalled(t, "PlayerPlay", 3)
}

func TestKingInteractor_Play_NotPlayable(t *testing.T) {
	gm, kp := newKingMocks()
	kp.On("Output", mock.Anything, mock.Anything).Return(kingMockOutput)
	gm.On("GetGameEndFlag").Return(false)
	gm.On("IsHumanTurn").Return(false) // not the human's turn -> guard blocks

	ki := usecase.NewKingInteractor(gm, kp)
	assert.Equal(t, kingMockOutput, ki.Play(0))
	gm.AssertNotCalled(t, "PlayerPlay", mock.Anything)
}

func TestKingInteractor_RunCpuTurns_Loop(t *testing.T) {
	gm, kp := newKingMocks()
	kp.On("Output", mock.Anything, mock.Anything).Return(kingMockOutput)
	gm.On("Reset").Return()
	// First IsHumanTurn false triggers CpuPlay, then phase == dealEnd stops loop.
	gm.On("GetGameEndFlag").Return(false)
	gm.On("IsHumanTurn").Return(false)
	gm.On("GetPhase").Return(domain.KingPhaseDealEnd)

	ki := usecase.NewKingInteractor(gm, kp)
	ki.Reset()
	// Loop returns at dealEnd without calling CpuPlay.
	gm.AssertNotCalled(t, "CpuPlay")
}

func TestKingInteractor_RunCpuTurns_CallsCpuUntilHuman(t *testing.T) {
	gm, kp := newKingMocks()
	kp.On("Output", mock.Anything, mock.Anything).Return(kingMockOutput)
	gm.On("Reset").Return()
	gm.On("GetGameEndFlag").Return(false)
	// Not human turn first (triggers CpuPlay), then human turn (stops loop).
	gm.On("IsHumanTurn").Return(false).Once()
	gm.On("IsHumanTurn").Return(true)
	gm.On("GetPhase").Return(domain.KingPhasePlay)
	gm.On("CpuPlay").Return()

	ki := usecase.NewKingInteractor(gm, kp)
	ki.Reset()
	gm.AssertCalled(t, "CpuPlay")
}

func TestKingInteractor_HintAndLog(t *testing.T) {
	gm, kp := newKingMocks()
	kp.On("HintOutput", mock.Anything).Return("hint")
	kp.On("ActionLogOutput", mock.Anything).Return("log")

	ki := usecase.NewKingInteractor(gm, kp)
	assert.Equal(t, "hint", ki.Hint())
	assert.Equal(t, "log", ki.ActionLog())
}

func TestKingInteractor_GetConfig(t *testing.T) {
	gm, kp := newKingMocks()
	cfg := domain.KingConfig{CpuDifficulty: domain.KingDifficultyEasy}
	gm.On("GetConfig").Return(cfg)

	ki := usecase.NewKingInteractor(gm, kp)
	assert.Equal(t, cfg, ki.GetConfig())
}

func TestKingInteractor_SnapshotAndRestore(t *testing.T) {
	kp := new(presenter.MockKingPresenter)
	kp.On("Output", mock.Anything, mock.Anything).Return(kingMockOutput)

	// Real game -> snapshot -> restore.
	real := usecase.NewKingInteractor(domain.NewDefaultKing(), kp)
	real.Reset()
	data, err := real.Snapshot()
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	restored, err := usecase.RestoreKingInteractor(data, kp)
	assert.NoError(t, err)
	assert.NotNil(t, restored)

	// Bad data -> error.
	_, err = usecase.RestoreKingInteractor([]byte("not json"), kp)
	assert.Error(t, err)
}
