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

const tarocchiniMockOutput = `{"phase":0}`

func newTarocchiniMocks() (*presenter.MockTarocchiniPresenter, *interfaces.MockTarocchiniGame) {
	tp := new(presenter.MockTarocchiniPresenter)
	tp.On("Output", mock.Anything, mock.Anything).Return(tarocchiniMockOutput)
	tp.On("HintOutput", mock.Anything).Return("hint")
	tp.On("ActionLogOutput", mock.Anything).Return("log")
	return tp, new(interfaces.MockTarocchiniGame)
}

// settled は「スカルト済みでプレイフェーズ、人間の手番」という定常状態を仕込む。
func settled(m *interfaces.MockTarocchiniGame) {
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.TarocchiniPhasePlay)
	m.On("IsHumanScartoTurn").Return(true)
	m.On("IsHumanTurn").Return(true)
}

func TestNewTarocchiniInteractor_NilGuards(t *testing.T) {
	tp := new(presenter.MockTarocchiniPresenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "TarocchiniInteractor: g must not be nil", func() {
			usecase.NewTarocchiniInteractor(nil, tp)
		})
	})
	t.Run("panics when tp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockTarocchiniGame)
		assert.PanicsWithValue(t, "TarocchiniInteractor: tp must not be nil", func() {
			usecase.NewTarocchiniInteractor(gameMock, nil)
		})
	})
}

func TestTarocchiniInteractor_Reset(t *testing.T) {
	tp, gameMock := newTarocchiniMocks()
	gameMock.On("Reset").Return()
	settled(gameMock)

	ci := usecase.NewTarocchiniInteractor(gameMock, tp)
	assert.Equal(t, tarocchiniMockOutput, ci.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestTarocchiniInteractor_ResetWithConfig(t *testing.T) {
	tp, gameMock := newTarocchiniMocks()
	cfg := domain.TarocchiniConfig{CpuDifficulty: domain.TarocchiniCpuDifficultyHard, TargetRounds: 8}
	gameMock.On("SetConfig", cfg).Return()
	gameMock.On("Reset").Return()
	settled(gameMock)

	ci := usecase.NewTarocchiniInteractor(gameMock, tp)
	assert.Equal(t, tarocchiniMockOutput, ci.ResetWithConfig(cfg))
	gameMock.AssertCalled(t, "SetConfig", cfg)
}

// 局数がプレイヤー数の倍数でない設定はゲームに届いてはならない。
func TestTarocchiniInteractor_ResetWithConfigInvalid(t *testing.T) {
	for name, cfg := range map[string]domain.TarocchiniConfig{
		"zero rounds":                        {CpuDifficulty: domain.TarocchiniCpuDifficultyNormal, TargetRounds: 0},
		"not a multiple of the player count": {CpuDifficulty: domain.TarocchiniCpuDifficultyNormal, TargetRounds: 6},
		"bad difficulty":                     {CpuDifficulty: 99, TargetRounds: 4},
	} {
		t.Run(name, func(t *testing.T) {
			tp, gameMock := newTarocchiniMocks()
			ci := usecase.NewTarocchiniInteractor(gameMock, tp)
			assert.Equal(t, tarocchiniMockOutput, ci.ResetWithConfig(cfg))
			gameMock.AssertNotCalled(t, "Reset")
		})
	}
}

func TestTarocchiniInteractor_Discard(t *testing.T) {
	tp, gameMock := newTarocchiniMocks()
	settled(gameMock)
	gameMock.On("PlayerScarto", []int{0, 1}).Return(nil)

	ci := usecase.NewTarocchiniInteractor(gameMock, tp)
	assert.Equal(t, tarocchiniMockOutput, ci.Discard([]int{0, 1}))
	gameMock.AssertCalled(t, "PlayerScarto", []int{0, 1})
}

func TestTarocchiniInteractor_DiscardError(t *testing.T) {
	tp, gameMock := newTarocchiniMocks()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("PlayerScarto", mock.Anything).Return(errors.New("trumps cannot be discarded"))

	ci := usecase.NewTarocchiniInteractor(gameMock, tp)
	assert.Equal(t, tarocchiniMockOutput, ci.Discard([]int{0, 1}))
	gameMock.AssertNotCalled(t, "CpuScarto")
}

func TestTarocchiniInteractor_DiscardGameEnded(t *testing.T) {
	tp, gameMock := newTarocchiniMocks()
	gameMock.On("GetGameEndFlag").Return(true)

	ci := usecase.NewTarocchiniInteractor(gameMock, tp)
	assert.Equal(t, tarocchiniMockOutput, ci.Discard([]int{0, 1}))
	gameMock.AssertNotCalled(t, "PlayerScarto", mock.Anything)
}

func TestTarocchiniInteractor_PlayResolvesTrick(t *testing.T) {
	tp, gameMock := newTarocchiniMocks()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("IsHumanScartoTurn").Return(true)
	gameMock.On("GetPhase").Return(domain.TarocchiniPhaseTrickEnd).Once()
	gameMock.On("GetPhase").Return(domain.TarocchiniPhaseRoundEnd)
	gameMock.On("PlayerPlay", 2).Return(nil)
	gameMock.On("ResolveTrick").Return()

	ci := usecase.NewTarocchiniInteractor(gameMock, tp)
	assert.Equal(t, tarocchiniMockOutput, ci.Play(2))
	gameMock.AssertCalled(t, "ResolveTrick")
}

func TestTarocchiniInteractor_PlayNoResolveWhenNotTrickEnd(t *testing.T) {
	tp, gameMock := newTarocchiniMocks()
	settled(gameMock)
	gameMock.On("PlayerPlay", 1).Return(nil)

	ci := usecase.NewTarocchiniInteractor(gameMock, tp)
	assert.Equal(t, tarocchiniMockOutput, ci.Play(1))
	gameMock.AssertNotCalled(t, "ResolveTrick")
}

func TestTarocchiniInteractor_PlayError(t *testing.T) {
	tp, gameMock := newTarocchiniMocks()
	settled(gameMock)
	gameMock.On("PlayerPlay", 9).Return(errors.New("must follow"))

	ci := usecase.NewTarocchiniInteractor(gameMock, tp)
	assert.Equal(t, tarocchiniMockOutput, ci.Play(9))
	gameMock.AssertNotCalled(t, "ResolveTrick")
}

func TestTarocchiniInteractor_PlayNotHumanTurn(t *testing.T) {
	tp, gameMock := newTarocchiniMocks()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(false)

	ci := usecase.NewTarocchiniInteractor(gameMock, tp)
	assert.Equal(t, tarocchiniMockOutput, ci.Play(0))
	gameMock.AssertNotCalled(t, "PlayerPlay", mock.Anything)
}

func TestTarocchiniInteractor_NextTrick(t *testing.T) {
	tp, gameMock := newTarocchiniMocks()
	gameMock.On("NextTrick").Return()
	settled(gameMock)

	ci := usecase.NewTarocchiniInteractor(gameMock, tp)
	assert.Equal(t, tarocchiniMockOutput, ci.NextTrick())
	gameMock.AssertCalled(t, "NextTrick")
}

func TestTarocchiniInteractor_NextRound(t *testing.T) {
	tp, gameMock := newTarocchiniMocks()
	gameMock.On("ScoreRound").Return()
	gameMock.On("NextRound").Return()
	settled(gameMock)

	ci := usecase.NewTarocchiniInteractor(gameMock, tp)
	assert.Equal(t, tarocchiniMockOutput, ci.NextRound())
	gameMock.AssertCalled(t, "NextRound")
}

func TestTarocchiniInteractor_NextRoundGameEnded(t *testing.T) {
	tp, gameMock := newTarocchiniMocks()
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(true)

	ci := usecase.NewTarocchiniInteractor(gameMock, tp)
	assert.Equal(t, tarocchiniMockOutput, ci.NextRound())
	gameMock.AssertNotCalled(t, "NextRound")
}

func TestTarocchiniInteractor_GetConfigHintActionLog(t *testing.T) {
	tp, gameMock := newTarocchiniMocks()
	cfg := domain.DefaultTarocchiniConfig()
	gameMock.On("GetConfig").Return(cfg)

	ci := usecase.NewTarocchiniInteractor(gameMock, tp)
	assert.Equal(t, cfg, ci.GetConfig())
	assert.Equal(t, "hint", ci.Hint())
	assert.Equal(t, "log", ci.ActionLog())
}

// **CPU がディーラーの局はスカルトを回さないと前に進まない。**トリックのループ
// だけではスカルトフェーズで止まる。
func TestTarocchiniInteractor_AdvanceRunsTheCpuScarto(t *testing.T) {
	tp, gameMock := newTarocchiniMocks()
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.TarocchiniPhaseScarto).Twice()
	gameMock.On("GetPhase").Return(domain.TarocchiniPhasePlay)
	gameMock.On("IsHumanScartoTurn").Return(false).Once()
	gameMock.On("IsHumanScartoTurn").Return(true)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("CpuScarto").Return()

	ci := usecase.NewTarocchiniInteractor(gameMock, tp)
	assert.Equal(t, tarocchiniMockOutput, ci.Reset())
	gameMock.AssertCalled(t, "CpuScarto")
}

func TestTarocchiniInteractor_AdvanceRunsCpuPlays(t *testing.T) {
	tp, gameMock := newTarocchiniMocks()
	gameMock.On("NextTrick").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.TarocchiniPhasePlay)
	gameMock.On("IsHumanScartoTurn").Return(true)
	gameMock.On("IsHumanTurn").Return(false).Times(3)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("CpuPlay").Return()

	ci := usecase.NewTarocchiniInteractor(gameMock, tp)
	assert.Equal(t, tarocchiniMockOutput, ci.NextTrick())
	gameMock.AssertNumberOfCalls(t, "CpuPlay", 3)
}

func TestRestoreTarocchiniInteractor(t *testing.T) {
	tp := new(presenter.MockTarocchiniPresenter)
	src := domain.NewDefaultTarocchini()
	src.Reset()
	data, err := src.MarshalJSON()
	assert.NoError(t, err)

	ci, err := usecase.RestoreTarocchiniInteractor(data, tp)
	assert.NoError(t, err)
	assert.NotNil(t, ci)

	_, err = usecase.RestoreTarocchiniInteractor([]byte(`{`), tp)
	assert.Error(t, err)
}
