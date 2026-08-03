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

const minchiateMockOutput = `{"phase":0}`

func newMinchiateMocks() (*presenter.MockMinchiatePresenter, *interfaces.MockMinchiateGame) {
	tp := new(presenter.MockMinchiatePresenter)
	tp.On("Output", mock.Anything, mock.Anything).Return(minchiateMockOutput)
	tp.On("HintOutput", mock.Anything).Return("hint")
	tp.On("ActionLogOutput", mock.Anything).Return("log")
	return tp, new(interfaces.MockMinchiateGame)
}

// minchiateSettled は「スカルト済みでプレイフェーズ、人間の手番」という定常状態を仕込む。
//
// **共通名 settled は使えない。**Tarocchini が同じパッケージで同名の helper を
// 持っているため衝突する。
func minchiateSettled(m *interfaces.MockMinchiateGame) {
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.MinchiatePhasePlay)
	m.On("IsHumanScartoTurn").Return(true)
	m.On("IsHumanTurn").Return(true)
}

func TestNewMinchiateInteractor_NilGuards(t *testing.T) {
	tp := new(presenter.MockMinchiatePresenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "MinchiateInteractor: g must not be nil", func() {
			usecase.NewMinchiateInteractor(nil, tp)
		})
	})
	t.Run("panics when tp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockMinchiateGame)
		assert.PanicsWithValue(t, "MinchiateInteractor: tp must not be nil", func() {
			usecase.NewMinchiateInteractor(gameMock, nil)
		})
	})
}

func TestMinchiateInteractor_Reset(t *testing.T) {
	tp, gameMock := newMinchiateMocks()
	gameMock.On("Reset").Return()
	minchiateSettled(gameMock)

	ci := usecase.NewMinchiateInteractor(gameMock, tp)
	assert.Equal(t, minchiateMockOutput, ci.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestMinchiateInteractor_ResetWithConfig(t *testing.T) {
	tp, gameMock := newMinchiateMocks()
	cfg := domain.MinchiateConfig{CpuDifficulty: domain.MinchiateCpuDifficultyHard, TargetRounds: 8}
	gameMock.On("SetConfig", cfg).Return()
	gameMock.On("Reset").Return()
	minchiateSettled(gameMock)

	ci := usecase.NewMinchiateInteractor(gameMock, tp)
	assert.Equal(t, minchiateMockOutput, ci.ResetWithConfig(cfg))
	gameMock.AssertCalled(t, "SetConfig", cfg)
}

// 局数がプレイヤー数の倍数でない設定はゲームに届いてはならない。
func TestMinchiateInteractor_ResetWithConfigInvalid(t *testing.T) {
	for name, cfg := range map[string]domain.MinchiateConfig{
		"zero rounds":                        {CpuDifficulty: domain.MinchiateCpuDifficultyNormal, TargetRounds: 0},
		"not a multiple of the player count": {CpuDifficulty: domain.MinchiateCpuDifficultyNormal, TargetRounds: 6},
		"bad difficulty":                     {CpuDifficulty: 99, TargetRounds: 4},
	} {
		t.Run(name, func(t *testing.T) {
			tp, gameMock := newMinchiateMocks()
			ci := usecase.NewMinchiateInteractor(gameMock, tp)
			assert.Equal(t, minchiateMockOutput, ci.ResetWithConfig(cfg))
			gameMock.AssertNotCalled(t, "Reset")
		})
	}
}

// minchiateSurplusIndices は捨てる枚数ぶんの位置を返す。
//
// **枚数は定数から出す。**Tarocchini から写した際は 2 枚のままだったが、
// Minchiate の余剰は 13 枚。テストに数字を直書きすると、読んだ人が誤った
// 枚数を仕様だと受け取る。
func minchiateSurplusIndices() []int {
	idx := make([]int, 0, domain.MinchiateSurplus)
	for i := 0; i < domain.MinchiateSurplus; i++ {
		idx = append(idx, i)
	}
	return idx
}

func TestMinchiateInteractor_Discard(t *testing.T) {
	tp, gameMock := newMinchiateMocks()
	minchiateSettled(gameMock)
	idx := minchiateSurplusIndices()
	gameMock.On("PlayerScarto", idx).Return(nil)

	ci := usecase.NewMinchiateInteractor(gameMock, tp)
	assert.Equal(t, minchiateMockOutput, ci.Discard(idx))
	gameMock.AssertCalled(t, "PlayerScarto", idx)
}

func TestMinchiateInteractor_DiscardError(t *testing.T) {
	tp, gameMock := newMinchiateMocks()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("PlayerScarto", mock.Anything).Return(errors.New("trumps cannot be discarded"))

	ci := usecase.NewMinchiateInteractor(gameMock, tp)
	assert.Equal(t, minchiateMockOutput, ci.Discard(minchiateSurplusIndices()))
	gameMock.AssertNotCalled(t, "CpuScarto")
}

func TestMinchiateInteractor_DiscardGameEnded(t *testing.T) {
	tp, gameMock := newMinchiateMocks()
	gameMock.On("GetGameEndFlag").Return(true)

	ci := usecase.NewMinchiateInteractor(gameMock, tp)
	assert.Equal(t, minchiateMockOutput, ci.Discard(minchiateSurplusIndices()))
	gameMock.AssertNotCalled(t, "PlayerScarto", mock.Anything)
}

func TestMinchiateInteractor_PlayResolvesTrick(t *testing.T) {
	tp, gameMock := newMinchiateMocks()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("IsHumanScartoTurn").Return(true)
	gameMock.On("GetPhase").Return(domain.MinchiatePhaseTrickEnd).Once()
	gameMock.On("GetPhase").Return(domain.MinchiatePhaseRoundEnd)
	gameMock.On("PlayerPlay", 2).Return(nil)
	gameMock.On("ResolveTrick").Return()

	ci := usecase.NewMinchiateInteractor(gameMock, tp)
	assert.Equal(t, minchiateMockOutput, ci.Play(2))
	gameMock.AssertCalled(t, "ResolveTrick")
}

func TestMinchiateInteractor_PlayNoResolveWhenNotTrickEnd(t *testing.T) {
	tp, gameMock := newMinchiateMocks()
	minchiateSettled(gameMock)
	gameMock.On("PlayerPlay", 1).Return(nil)

	ci := usecase.NewMinchiateInteractor(gameMock, tp)
	assert.Equal(t, minchiateMockOutput, ci.Play(1))
	gameMock.AssertNotCalled(t, "ResolveTrick")
}

func TestMinchiateInteractor_PlayError(t *testing.T) {
	tp, gameMock := newMinchiateMocks()
	minchiateSettled(gameMock)
	gameMock.On("PlayerPlay", 9).Return(errors.New("must follow"))

	ci := usecase.NewMinchiateInteractor(gameMock, tp)
	assert.Equal(t, minchiateMockOutput, ci.Play(9))
	gameMock.AssertNotCalled(t, "ResolveTrick")
}

func TestMinchiateInteractor_PlayNotHumanTurn(t *testing.T) {
	tp, gameMock := newMinchiateMocks()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(false)

	ci := usecase.NewMinchiateInteractor(gameMock, tp)
	assert.Equal(t, minchiateMockOutput, ci.Play(0))
	gameMock.AssertNotCalled(t, "PlayerPlay", mock.Anything)
}

func TestMinchiateInteractor_NextTrick(t *testing.T) {
	tp, gameMock := newMinchiateMocks()
	gameMock.On("NextTrick").Return()
	minchiateSettled(gameMock)

	ci := usecase.NewMinchiateInteractor(gameMock, tp)
	assert.Equal(t, minchiateMockOutput, ci.NextTrick())
	gameMock.AssertCalled(t, "NextTrick")
}

func TestMinchiateInteractor_NextRound(t *testing.T) {
	tp, gameMock := newMinchiateMocks()
	gameMock.On("ScoreRound").Return()
	gameMock.On("NextRound").Return()
	minchiateSettled(gameMock)

	ci := usecase.NewMinchiateInteractor(gameMock, tp)
	assert.Equal(t, minchiateMockOutput, ci.NextRound())
	gameMock.AssertCalled(t, "NextRound")
}

func TestMinchiateInteractor_NextRoundGameEnded(t *testing.T) {
	tp, gameMock := newMinchiateMocks()
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(true)

	ci := usecase.NewMinchiateInteractor(gameMock, tp)
	assert.Equal(t, minchiateMockOutput, ci.NextRound())
	gameMock.AssertNotCalled(t, "NextRound")
}

func TestMinchiateInteractor_GetConfigHintActionLog(t *testing.T) {
	tp, gameMock := newMinchiateMocks()
	cfg := domain.DefaultMinchiateConfig()
	gameMock.On("GetConfig").Return(cfg)

	ci := usecase.NewMinchiateInteractor(gameMock, tp)
	assert.Equal(t, cfg, ci.GetConfig())
	assert.Equal(t, "hint", ci.Hint())
	assert.Equal(t, "log", ci.ActionLog())
}

// **CPU がディーラーの局はスカルトを回さないと前に進まない。**トリックのループ
// だけではスカルトフェーズで止まる。
func TestMinchiateInteractor_AdvanceRunsTheCpuScarto(t *testing.T) {
	tp, gameMock := newMinchiateMocks()
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.MinchiatePhaseScarto).Twice()
	gameMock.On("GetPhase").Return(domain.MinchiatePhasePlay)
	gameMock.On("IsHumanScartoTurn").Return(false).Once()
	gameMock.On("IsHumanScartoTurn").Return(true)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("CpuScarto").Return()

	ci := usecase.NewMinchiateInteractor(gameMock, tp)
	assert.Equal(t, minchiateMockOutput, ci.Reset())
	gameMock.AssertCalled(t, "CpuScarto")
}

func TestMinchiateInteractor_AdvanceRunsCpuPlays(t *testing.T) {
	tp, gameMock := newMinchiateMocks()
	gameMock.On("NextTrick").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.MinchiatePhasePlay)
	gameMock.On("IsHumanScartoTurn").Return(true)
	gameMock.On("IsHumanTurn").Return(false).Times(3)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("CpuPlay").Return()

	ci := usecase.NewMinchiateInteractor(gameMock, tp)
	assert.Equal(t, minchiateMockOutput, ci.NextTrick())
	gameMock.AssertNumberOfCalls(t, "CpuPlay", 3)
}

func TestRestoreMinchiateInteractor(t *testing.T) {
	tp := new(presenter.MockMinchiatePresenter)
	src := domain.NewDefaultMinchiate()
	src.Reset()
	data, err := src.MarshalJSON()
	assert.NoError(t, err)

	ci, err := usecase.RestoreMinchiateInteractor(data, tp)
	assert.NoError(t, err)
	assert.NotNil(t, ci)

	_, err = usecase.RestoreMinchiateInteractor([]byte(`{`), tp)
	assert.Error(t, err)
}
