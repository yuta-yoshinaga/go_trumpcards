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

const aluetteMockOutput = `{"phase":0}`

func newAluetteMocks() (*presenter.MockAluettePresenter, *interfaces.MockAluetteGame) {
	tp := new(presenter.MockAluettePresenter)
	tp.On("Output", mock.Anything, mock.Anything).Return(aluetteMockOutput)
	tp.On("HintOutput", mock.Anything).Return("hint")
	tp.On("ActionLogOutput", mock.Anything).Return("log")
	return tp, new(interfaces.MockAluetteGame)
}

// aluetteSettled は「プレイフェーズ、人間の手番」という定常状態を仕込む。
func aluetteSettled(m *interfaces.MockAluetteGame) {
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.AluettePhasePlay)
	m.On("IsHumanTurn").Return(true)
}

func TestNewAluetteInteractor_NilGuards(t *testing.T) {
	tp := new(presenter.MockAluettePresenter)
	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "AluetteInteractor: g must not be nil", func() {
			usecase.NewAluetteInteractor(nil, tp)
		})
	})
	t.Run("panics when tp is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "AluetteInteractor: tp must not be nil", func() {
			usecase.NewAluetteInteractor(new(interfaces.MockAluetteGame), nil)
		})
	})
}

func TestAluetteInteractor_Reset(t *testing.T) {
	tp, gameMock := newAluetteMocks()
	gameMock.On("Reset").Return()
	aluetteSettled(gameMock)

	ci := usecase.NewAluetteInteractor(gameMock, tp)
	assert.Equal(t, aluetteMockOutput, ci.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestAluetteInteractor_ResetWithConfig(t *testing.T) {
	tp, gameMock := newAluetteMocks()
	cfg := domain.AluetteConfig{CpuDifficulty: domain.AluetteCpuDifficultyHard, TargetPoints: 10}
	gameMock.On("SetConfig", cfg).Return()
	gameMock.On("Reset").Return()
	aluetteSettled(gameMock)

	ci := usecase.NewAluetteInteractor(gameMock, tp)
	assert.Equal(t, aluetteMockOutput, ci.ResetWithConfig(cfg))
	gameMock.AssertCalled(t, "SetConfig", cfg)
}

func TestAluetteInteractor_ResetWithConfigInvalid(t *testing.T) {
	for name, cfg := range map[string]domain.AluetteConfig{
		"zero target":    {CpuDifficulty: domain.AluetteCpuDifficultyNormal, TargetPoints: 0},
		"bad difficulty": {CpuDifficulty: 99, TargetPoints: 6},
	} {
		t.Run(name, func(t *testing.T) {
			tp, gameMock := newAluetteMocks()
			ci := usecase.NewAluetteInteractor(gameMock, tp)
			assert.Equal(t, aluetteMockOutput, ci.ResetWithConfig(cfg))
			gameMock.AssertNotCalled(t, "Reset")
		})
	}
}

func TestAluetteInteractor_PlayResolvesTrick(t *testing.T) {
	tp, gameMock := newAluetteMocks()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("GetPhase").Return(domain.AluettePhaseTrickEnd).Once()
	gameMock.On("GetPhase").Return(domain.AluettePhaseRoundEnd)
	gameMock.On("PlayerPlay", 2).Return(nil)
	gameMock.On("ResolveTrick").Return()

	ci := usecase.NewAluetteInteractor(gameMock, tp)
	assert.Equal(t, aluetteMockOutput, ci.Play(2))
	gameMock.AssertCalled(t, "ResolveTrick")
}

func TestAluetteInteractor_PlayNoResolveWhenNotTrickEnd(t *testing.T) {
	tp, gameMock := newAluetteMocks()
	aluetteSettled(gameMock)
	gameMock.On("PlayerPlay", 1).Return(nil)

	ci := usecase.NewAluetteInteractor(gameMock, tp)
	assert.Equal(t, aluetteMockOutput, ci.Play(1))
	gameMock.AssertNotCalled(t, "ResolveTrick")
}

func TestAluetteInteractor_PlayError(t *testing.T) {
	tp, gameMock := newAluetteMocks()
	aluetteSettled(gameMock)
	gameMock.On("PlayerPlay", 9).Return(errors.New("out of range"))

	ci := usecase.NewAluetteInteractor(gameMock, tp)
	assert.Equal(t, aluetteMockOutput, ci.Play(9))
	gameMock.AssertNotCalled(t, "ResolveTrick")
}

func TestAluetteInteractor_PlayNotHumanTurn(t *testing.T) {
	tp, gameMock := newAluetteMocks()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(false)

	ci := usecase.NewAluetteInteractor(gameMock, tp)
	assert.Equal(t, aluetteMockOutput, ci.Play(0))
	gameMock.AssertNotCalled(t, "PlayerPlay", mock.Anything)
}

func TestAluetteInteractor_NextTrick(t *testing.T) {
	tp, gameMock := newAluetteMocks()
	gameMock.On("NextTrick").Return()
	aluetteSettled(gameMock)

	ci := usecase.NewAluetteInteractor(gameMock, tp)
	assert.Equal(t, aluetteMockOutput, ci.NextTrick())
	gameMock.AssertCalled(t, "NextTrick")
}

func TestAluetteInteractor_NextRound(t *testing.T) {
	tp, gameMock := newAluetteMocks()
	gameMock.On("ScoreRound").Return()
	gameMock.On("NextRound").Return()
	aluetteSettled(gameMock)

	ci := usecase.NewAluetteInteractor(gameMock, tp)
	assert.Equal(t, aluetteMockOutput, ci.NextRound())
	gameMock.AssertCalled(t, "NextRound")
}

func TestAluetteInteractor_NextRoundGameEnded(t *testing.T) {
	tp, gameMock := newAluetteMocks()
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(true)

	ci := usecase.NewAluetteInteractor(gameMock, tp)
	assert.Equal(t, aluetteMockOutput, ci.NextRound())
	gameMock.AssertNotCalled(t, "NextRound")
}

func TestAluetteInteractor_GetConfigHintActionLog(t *testing.T) {
	tp, gameMock := newAluetteMocks()
	cfg := domain.DefaultAluetteConfig()
	gameMock.On("GetConfig").Return(cfg)

	ci := usecase.NewAluetteInteractor(gameMock, tp)
	assert.Equal(t, cfg, ci.GetConfig())
	assert.Equal(t, "hint", ci.Hint())
	assert.Equal(t, "log", ci.ActionLog())
}

// **スカルトのループは無い。**配ったらすぐトリックが始まるので、CPU の
// プレイだけを人間の手番まで回せばよい。
func TestAluetteInteractor_AdvanceRunsCpuPlays(t *testing.T) {
	tp, gameMock := newAluetteMocks()
	gameMock.On("NextTrick").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.AluettePhasePlay)
	gameMock.On("IsHumanTurn").Return(false).Times(3)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("CpuPlay").Return()

	ci := usecase.NewAluetteInteractor(gameMock, tp)
	assert.Equal(t, aluetteMockOutput, ci.NextTrick())
	gameMock.AssertNumberOfCalls(t, "CpuPlay", 3)
}

func TestRestoreAluetteInteractor(t *testing.T) {
	tp := new(presenter.MockAluettePresenter)
	src := domain.NewDefaultAluette()
	src.Reset()
	data, err := src.MarshalJSON()
	assert.NoError(t, err)

	ci, err := usecase.RestoreAluetteInteractor(data, tp)
	assert.NoError(t, err)
	assert.NotNil(t, ci)

	_, err = usecase.RestoreAluetteInteractor([]byte(`{`), tp)
	assert.Error(t, err)
}
