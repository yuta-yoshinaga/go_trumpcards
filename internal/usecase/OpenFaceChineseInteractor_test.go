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

const ofcMockOutput = `{"phase":0}`

func newOfcMocks() (*presenter.MockOpenFaceChinesePresenter, *interfaces.MockOpenFaceChineseGame) {
	sp := new(presenter.MockOpenFaceChinesePresenter)
	sp.On("Output", mock.Anything, mock.Anything).Return(ofcMockOutput)
	sp.On("HintOutput", mock.Anything).Return("hint")
	sp.On("ActionLogOutput", mock.Anything).Return("log")
	return sp, new(interfaces.MockOpenFaceChineseGame)
}

func TestNewOpenFaceChineseInteractor_NilGuards(t *testing.T) {
	sp := new(presenter.MockOpenFaceChinesePresenter)
	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "OpenFaceChineseInteractor: g must not be nil", func() {
			usecase.NewOpenFaceChineseInteractor(nil, sp)
		})
	})
	t.Run("panics when sp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockOpenFaceChineseGame)
		assert.PanicsWithValue(t, "OpenFaceChineseInteractor: sp must not be nil", func() {
			usecase.NewOpenFaceChineseInteractor(gameMock, nil)
		})
	})
}

func TestOpenFaceChineseInteractor_Reset(t *testing.T) {
	sp, gameMock := newOfcMocks()
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.OpenFaceChinesePhasePlacing)
	gameMock.On("IsHumanTurn").Return(true)

	ti := usecase.NewOpenFaceChineseInteractor(gameMock, sp)
	assert.Equal(t, ofcMockOutput, ti.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestOpenFaceChineseInteractor_ResetWithConfig(t *testing.T) {
	sp, gameMock := newOfcMocks()
	cfg := domain.OpenFaceChineseConfig{CpuDifficulty: domain.OpenFaceChineseCpuDifficultyHard, PlayerCount: 2, TargetRounds: 4}
	gameMock.On("SetConfig", cfg).Return()
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.OpenFaceChinesePhasePlacing)
	gameMock.On("IsHumanTurn").Return(true)

	ti := usecase.NewOpenFaceChineseInteractor(gameMock, sp)
	assert.Equal(t, ofcMockOutput, ti.ResetWithConfig(cfg))
	gameMock.AssertCalled(t, "SetConfig", cfg)
}

func TestOpenFaceChineseInteractor_ResetWithConfigInvalid(t *testing.T) {
	sp, gameMock := newOfcMocks()
	ti := usecase.NewOpenFaceChineseInteractor(gameMock, sp)
	bad := domain.OpenFaceChineseConfig{CpuDifficulty: domain.OpenFaceChineseCpuDifficultyNormal, PlayerCount: 1, TargetRounds: 4}
	assert.Equal(t, ofcMockOutput, ti.ResetWithConfig(bad))
	gameMock.AssertNotCalled(t, "Reset")
}

func TestOpenFaceChineseInteractor_Place(t *testing.T) {
	sp, gameMock := newOfcMocks()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlace", 2).Return(nil)
	gameMock.On("GetPhase").Return(domain.OpenFaceChinesePhasePlacing)

	ti := usecase.NewOpenFaceChineseInteractor(gameMock, sp)
	assert.Equal(t, ofcMockOutput, ti.Place(2))
	gameMock.AssertCalled(t, "PlayerPlace", 2)
}

func TestOpenFaceChineseInteractor_PlaceBlockedWhenNotHuman(t *testing.T) {
	sp, gameMock := newOfcMocks()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(false)

	ti := usecase.NewOpenFaceChineseInteractor(gameMock, sp)
	assert.Equal(t, ofcMockOutput, ti.Place(0))
	gameMock.AssertNotCalled(t, "PlayerPlace", mock.Anything)
}

func TestOpenFaceChineseInteractor_PlaceError(t *testing.T) {
	sp, gameMock := newOfcMocks()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlace", 0).Return(errors.New("row full"))

	ti := usecase.NewOpenFaceChineseInteractor(gameMock, sp)
	assert.Equal(t, ofcMockOutput, ti.Place(0))
}

func TestOpenFaceChineseInteractor_NextRound(t *testing.T) {
	sp, gameMock := newOfcMocks()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("NextRound").Return()
	gameMock.On("GetPhase").Return(domain.OpenFaceChinesePhasePlacing)
	gameMock.On("IsHumanTurn").Return(true)

	ti := usecase.NewOpenFaceChineseInteractor(gameMock, sp)
	assert.Equal(t, ofcMockOutput, ti.NextRound())
	gameMock.AssertCalled(t, "NextRound")
}

func TestOpenFaceChineseInteractor_NextRoundBlockedWhenEnded(t *testing.T) {
	sp, gameMock := newOfcMocks()
	gameMock.On("GetGameEndFlag").Return(true)

	ti := usecase.NewOpenFaceChineseInteractor(gameMock, sp)
	assert.Equal(t, ofcMockOutput, ti.NextRound())
	gameMock.AssertNotCalled(t, "NextRound")
}

func TestOpenFaceChineseInteractor_AdvanceCpuLoop(t *testing.T) {
	sp, gameMock := newOfcMocks()
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.OpenFaceChinesePhasePlacing)
	// First IsHumanTurn=false triggers a CpuPlay, then true stops the loop.
	gameMock.On("IsHumanTurn").Return(false).Once()
	gameMock.On("CpuPlay").Return().Once()
	gameMock.On("IsHumanTurn").Return(true)

	ti := usecase.NewOpenFaceChineseInteractor(gameMock, sp)
	ti.Reset()
	gameMock.AssertCalled(t, "CpuPlay")
}

func TestOpenFaceChineseInteractor_HintAndLogAndConfig(t *testing.T) {
	sp, gameMock := newOfcMocks()
	cfg := domain.DefaultOpenFaceChineseConfig()
	gameMock.On("GetConfig").Return(cfg)

	ti := usecase.NewOpenFaceChineseInteractor(gameMock, sp)
	assert.Equal(t, "hint", ti.Hint())
	assert.Equal(t, "log", ti.ActionLog())
	assert.Equal(t, cfg, ti.GetConfig())
}

func TestOpenFaceChineseInteractor_Snapshot(t *testing.T) {
	sp := new(presenter.MockOpenFaceChinesePresenter)
	g := domain.NewDefaultOpenFaceChinese()
	g.Reset()
	ti := usecase.NewOpenFaceChineseInteractor(g, sp)
	data, err := ti.Snapshot()
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	restored, err := usecase.RestoreOpenFaceChineseInteractor(data, sp)
	assert.NoError(t, err)
	assert.NotNil(t, restored)
}

func TestRestoreOpenFaceChineseInteractor_Invalid(t *testing.T) {
	sp := new(presenter.MockOpenFaceChinesePresenter)
	_, err := usecase.RestoreOpenFaceChineseInteractor([]byte("{invalid"), sp)
	assert.Error(t, err)
}
