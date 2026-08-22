//go:build test && (!js || !wasm || casino)

package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// newThreeCardRummyInteractorForTest wires a fresh mock pair to the interactor
// under test. Every test gets its own pair — testify's expectations are
// stateful, and a shared mock would let one test satisfy another's assertion.
func newThreeCardRummyInteractorForTest(t *testing.T) (
	*interfaces.MockThreeCardRummyGame,
	*presenter.MockThreeCardRummyPresenter,
	*ThreeCardRummyInteractor,
) {
	t.Helper()
	mockGame := new(interfaces.MockThreeCardRummyGame)
	mockPresenter := new(presenter.MockThreeCardRummyPresenter)
	return mockGame, mockPresenter, NewThreeCardRummyInteractor(mockGame, mockPresenter)
}

func TestNewThreeCardRummyInteractor(t *testing.T) {
	_, _, ti := newThreeCardRummyInteractorForTest(t)
	assert.NotNil(t, ti)
}

func TestNewThreeCardRummyInteractor_NilPanics(t *testing.T) {
	mockPresenter := new(presenter.MockThreeCardRummyPresenter)
	assert.Panics(t, func() { NewThreeCardRummyInteractor(nil, mockPresenter) })

	mockGame := new(interfaces.MockThreeCardRummyGame)
	assert.Panics(t, func() { NewThreeCardRummyInteractor(mockGame, nil) })
}

func TestThreeCardRummyInteractor_Reset(t *testing.T) {
	mockGame, mockPresenter, ti := newThreeCardRummyInteractorForTest(t)

	mockGame.On("Reset").Return()
	mockPresenter.On("Output", mockGame, nil).Return("reset output")

	assert.Equal(t, "reset output", ti.Reset())
	mockGame.AssertCalled(t, "Reset")
	mockPresenter.AssertExpectations(t)
}

func TestThreeCardRummyInteractor_Bet(t *testing.T) {
	// **アンテとローボーナスは別の賭け。** 引数を入れ替えると testify が
	// 未登録呼び出しで落ちるので、順序もここで固定される。
	t.Run("passes both stakes through in order", func(t *testing.T) {
		mockGame, mockPresenter, ti := newThreeCardRummyInteractorForTest(t)

		mockGame.On("Bet", 100, 50).Return(nil)
		mockPresenter.On("Output", mockGame, nil).Return("bet output")

		assert.Equal(t, "bet output", ti.Bet(100, 50))
		mockGame.AssertCalled(t, "Bet", 100, 50)
		mockPresenter.AssertExpectations(t)
	})

	// 断られたベットも提示に回る。握りつぶすと「なぜ賭けられないのか」が
	// プレイヤーに届かない。
	t.Run("surfaces the rejection", func(t *testing.T) {
		mockGame, mockPresenter, ti := newThreeCardRummyInteractorForTest(t)

		wantErr := errors.New("Insufficient chips.")
		mockGame.On("Bet", 100, 0).Return(wantErr)
		mockPresenter.On("Output", mockGame, mock.MatchedBy(func(e error) bool {
			return errors.Is(e, wantErr)
		})).Return("error output")

		assert.Equal(t, "error output", ti.Bet(100, 0))
		mockPresenter.AssertExpectations(t)
	})
}

func TestThreeCardRummyInteractor_Rebet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockGame, mockPresenter, ti := newThreeCardRummyInteractorForTest(t)

		mockGame.On("Rebet").Return(nil)
		mockPresenter.On("Output", mockGame, nil).Return("rebet output")

		assert.Equal(t, "rebet output", ti.Rebet())
		mockGame.AssertCalled(t, "Rebet")
		mockPresenter.AssertExpectations(t)
	})

	t.Run("surfaces the rejection", func(t *testing.T) {
		mockGame, mockPresenter, ti := newThreeCardRummyInteractorForTest(t)

		wantErr := errors.New("まだ賭けていないので再ベットできません")
		mockGame.On("Rebet").Return(wantErr)
		mockPresenter.On("Output", mockGame, wantErr).Return("rebet error output")

		assert.Equal(t, "rebet error output", ti.Rebet())
		mockPresenter.AssertExpectations(t)
	})
}

func TestThreeCardRummyInteractor_Play(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockGame, mockPresenter, ti := newThreeCardRummyInteractorForTest(t)

		mockGame.On("Play").Return(nil)
		mockPresenter.On("Output", mockGame, nil).Return("play output")

		assert.Equal(t, "play output", ti.Play())
		mockGame.AssertCalled(t, "Play")
		mockPresenter.AssertExpectations(t)
	})

	t.Run("surfaces the rejection", func(t *testing.T) {
		mockGame, mockPresenter, ti := newThreeCardRummyInteractorForTest(t)

		wantErr := errors.New("Play is only allowed during the action phase.")
		mockGame.On("Play").Return(wantErr)
		mockPresenter.On("Output", mockGame, wantErr).Return("play error output")

		assert.Equal(t, "play error output", ti.Play())
		mockPresenter.AssertExpectations(t)
	})
}

func TestThreeCardRummyInteractor_Fold(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockGame, mockPresenter, ti := newThreeCardRummyInteractorForTest(t)

		mockGame.On("Fold").Return(nil)
		mockPresenter.On("Output", mockGame, nil).Return("fold output")

		assert.Equal(t, "fold output", ti.Fold())
		mockGame.AssertCalled(t, "Fold")
		mockPresenter.AssertExpectations(t)
	})

	t.Run("surfaces the rejection", func(t *testing.T) {
		mockGame, mockPresenter, ti := newThreeCardRummyInteractorForTest(t)

		wantErr := errors.New("Fold is only allowed during the action phase.")
		mockGame.On("Fold").Return(wantErr)
		mockPresenter.On("Output", mockGame, wantErr).Return("fold error output")

		assert.Equal(t, "fold error output", ti.Fold())
		mockPresenter.AssertExpectations(t)
	})
}

func TestThreeCardRummyInteractor_ActionLog(t *testing.T) {
	mockGame, mockPresenter, ti := newThreeCardRummyInteractorForTest(t)

	mockPresenter.On("ActionLogOutput", mockGame).Return("log output")

	assert.Equal(t, "log output", ti.ActionLog())
	mockPresenter.AssertExpectations(t)
}

func TestThreeCardRummyInteractor_Hint(t *testing.T) {
	mockGame, mockPresenter, ti := newThreeCardRummyInteractorForTest(t)

	mockPresenter.On("HintOutput", mockGame).Return("hint output")

	assert.Equal(t, "hint output", ti.Hint())
	mockPresenter.AssertExpectations(t)
}

// ヒントと棋譜はドメインを進めない。**読むだけの操作**なので、ここで
// Reset/Bet/Play/Fold のいずれかを呼んでいたら卓が勝手に動く。
func TestThreeCardRummyInteractor_ReadOnlyCallsDoNotAdvanceTheGame(t *testing.T) {
	mockGame, mockPresenter, ti := newThreeCardRummyInteractorForTest(t)

	mockPresenter.On("HintOutput", mockGame).Return("hint output")
	mockPresenter.On("ActionLogOutput", mockGame).Return("log output")

	ti.Hint()
	ti.ActionLog()

	mockGame.AssertNotCalled(t, "Reset")
	mockGame.AssertNotCalled(t, "Bet", mock.Anything, mock.Anything)
	mockGame.AssertNotCalled(t, "Rebet")
	mockGame.AssertNotCalled(t, "Play")
	mockGame.AssertNotCalled(t, "Fold")
	// Output は「状態が変わったとき」の提示口。読むだけの操作が通ったら、
	// 呼び出し側は棋譜の表示で盤面を上書きされる。
	mockPresenter.AssertNotCalled(t, "Output", mock.Anything, mock.Anything)
}
