package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// newCaribbeanDrawDrawFixture wires an interactor whose game mock records the
// exact slice handed to Draw.
//
// **記録しないと素通りを検査できない。** `On("Draw", mock.Anything)` だけでは
// 「Draw が呼ばれた」ことしか分からず、添字を捨てた実装でも黙って通る。
func newCaribbeanDrawDrawFixture(t *testing.T, out string, retErr error) (
	*CaribbeanDrawInteractor, *interfaces.MockCaribbeanDrawGame, *[]int, *bool,
) {
	t.Helper()
	mockGame := new(interfaces.MockCaribbeanDrawGame)
	mockPresenter := new(presenter.MockCaribbeanDrawPresenter)
	ci := NewCaribbeanDrawInteractor(mockGame, mockPresenter)

	var got []int
	called := false
	mockGame.On("Draw", mock.Anything).Run(func(args mock.Arguments) {
		called = true
		got, _ = args.Get(0).([]int)
	}).Return(retErr)

	if retErr == nil {
		mockPresenter.On("Output", mockGame, nil).Return(out)
	} else {
		mockPresenter.On("Output", mockGame, mock.MatchedBy(func(e error) bool {
			return e != nil && e.Error() == retErr.Error()
		})).Return(out)
	}
	return ci, mockGame, &got, &called
}

// TestCaribbeanDrawInteractor_Draw covers the draw phase the clone source
// (Caribbean Stud) does not have: the interactor is a pass-through, so the only
// thing it can get wrong is the payload.
func TestCaribbeanDrawInteractor_Draw(t *testing.T) {
	t.Run("passes the indices through to the domain untouched", func(t *testing.T) {
		ci, mockGame, got, called := newCaribbeanDrawDrawFixture(t, "draw output", nil)

		result := ci.Draw([]int{0, 2})

		assert.Equal(t, "draw output", result)
		require.True(t, *called, "Draw must reach the domain")
		// **順序も枚数もそのまま。** ここで並べ替えたり切り詰めたりすると、
		// プレイヤーが指した札とは別の札が消える。
		assert.Equal(t, []int{0, 2}, *got)
		mockGame.AssertCalled(t, "Draw", []int{0, 2})
	})

	t.Run("passes the maximum two indices through", func(t *testing.T) {
		ci, _, got, called := newCaribbeanDrawDrawFixture(t, "draw output", nil)

		ci.Draw([]int{4, 1})

		require.True(t, *called)
		assert.Equal(t, []int{4, 1}, *got)
	})

	t.Run("stands pat on a nil index list", func(t *testing.T) {
		ci, _, got, called := newCaribbeanDrawDrawFixture(t, "stand pat output", nil)

		result := ci.Draw(nil)

		assert.Equal(t, "stand pat output", result)
		require.True(t, *called, "standing pat still has to advance the phase")
		assert.Empty(t, *got)
	})

	t.Run("stands pat on an empty index list", func(t *testing.T) {
		ci, _, got, called := newCaribbeanDrawDrawFixture(t, "stand pat output", nil)

		result := ci.Draw([]int{})

		assert.Equal(t, "stand pat output", result)
		require.True(t, *called)
		assert.Empty(t, *got)
	})

	// **拒否は presenter 経由で返る。** エラーを握り潰すと、範囲外の添字を
	// 打ったプレイヤーには「何も起きなかった」ようにしか見えない。
	t.Run("surfaces a domain rejection through the presenter", func(t *testing.T) {
		wrongPhase := errors.New("Draw is only allowed during the draw phase.")
		ci, _, got, called := newCaribbeanDrawDrawFixture(t, "error output", wrongPhase)

		result := ci.Draw([]int{1})

		assert.Equal(t, "error output", result)
		require.True(t, *called)
		assert.Equal(t, []int{1}, *got)
	})

	t.Run("surfaces an out-of-range rejection", func(t *testing.T) {
		outOfRange := errors.New("Card index out of range.")
		ci, _, _, _ := newCaribbeanDrawDrawFixture(t, "range error output", outOfRange)

		assert.Equal(t, "range error output", ci.Draw([]int{9}))
	})
}

// TestCaribbeanDrawInteractor_DrawIsOnTheInterface pins Draw to the interface the
// controllers depend on; dropping it there would strand every caller.
func TestCaribbeanDrawInteractor_DrawIsOnTheInterface(t *testing.T) {
	mockGame := new(interfaces.MockCaribbeanDrawGame)
	mockPresenter := new(presenter.MockCaribbeanDrawPresenter)
	var ci CaribbeanDrawInteractorIF = NewCaribbeanDrawInteractor(mockGame, mockPresenter)

	mockGame.On("Draw", []int{3}).Return(nil)
	mockPresenter.On("Output", mockGame, nil).Return("iface output")

	assert.Equal(t, "iface output", ci.Draw([]int{3}))
	mockGame.AssertCalled(t, "Draw", []int{3})
}
