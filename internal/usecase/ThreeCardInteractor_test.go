package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func TestNewThreeCardInteractor(t *testing.T) {
	mockGame := new(interfaces.MockThreeCardGame)
	mockPresenter := new(presenter.MockThreeCardPresenter)
	ti := NewThreeCardInteractor(mockGame, mockPresenter)
	assert.NotNil(t, ti)
}

func TestNewThreeCardInteractor_NilPanics(t *testing.T) {
	mockPresenter := new(presenter.MockThreeCardPresenter)
	assert.Panics(t, func() { NewThreeCardInteractor(nil, mockPresenter) })

	mockGame := new(interfaces.MockThreeCardGame)
	assert.Panics(t, func() { NewThreeCardInteractor(mockGame, nil) })
}

func TestThreeCardInteractor_Reset(t *testing.T) {
	mockGame := new(interfaces.MockThreeCardGame)
	mockPresenter := new(presenter.MockThreeCardPresenter)
	ti := NewThreeCardInteractor(mockGame, mockPresenter)

	mockGame.On("Reset").Return()
	mockPresenter.On("Output", mockGame, nil).Return("reset output")

	result := ti.Reset()
	assert.Equal(t, "reset output", result)
	mockGame.AssertCalled(t, "Reset")
}

func TestThreeCardInteractor_Hint(t *testing.T) {
	mockGame := new(interfaces.MockThreeCardGame)
	mockPresenter := new(presenter.MockThreeCardPresenter)
	ti := NewThreeCardInteractor(mockGame, mockPresenter)

	mockPresenter.On("HintOutput", mockGame).Return("hint output")

	assert.Equal(t, "hint output", ti.Hint())
}

func TestThreeCardInteractor_Bet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockGame := new(interfaces.MockThreeCardGame)
		mockPresenter := new(presenter.MockThreeCardPresenter)
		ti := NewThreeCardInteractor(mockGame, mockPresenter)

		mockGame.On("Bet", 100, 50).Return(nil)
		mockPresenter.On("Output", mockGame, nil).Return("bet output")

		result := ti.Bet(100, 50)
		assert.Equal(t, "bet output", result)
	})

	t.Run("error", func(t *testing.T) {
		mockGame := new(interfaces.MockThreeCardGame)
		mockPresenter := new(presenter.MockThreeCardPresenter)
		ti := NewThreeCardInteractor(mockGame, mockPresenter)

		err := errors.New("test error")
		mockGame.On("Bet", 100, 0).Return(err)
		mockPresenter.On("Output", mockGame, mock.MatchedBy(func(e error) bool {
			return e != nil && e.Error() == "test error"
		})).Return("error output")

		result := ti.Bet(100, 0)
		assert.Equal(t, "error output", result)
	})
}

func TestThreeCardInteractor_Play(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockGame := new(interfaces.MockThreeCardGame)
		mockPresenter := new(presenter.MockThreeCardPresenter)
		ti := NewThreeCardInteractor(mockGame, mockPresenter)

		mockGame.On("Play").Return(nil)
		mockPresenter.On("Output", mockGame, nil).Return("play output")

		result := ti.Play()
		assert.Equal(t, "play output", result)
	})

	t.Run("error", func(t *testing.T) {
		mockGame := new(interfaces.MockThreeCardGame)
		mockPresenter := new(presenter.MockThreeCardPresenter)
		ti := NewThreeCardInteractor(mockGame, mockPresenter)

		err := errors.New("play error")
		mockGame.On("Play").Return(err)
		mockPresenter.On("Output", mockGame, mock.MatchedBy(func(e error) bool {
			return e != nil
		})).Return("play error output")

		result := ti.Play()
		assert.Equal(t, "play error output", result)
	})
}

func TestThreeCardInteractor_Fold(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockGame := new(interfaces.MockThreeCardGame)
		mockPresenter := new(presenter.MockThreeCardPresenter)
		ti := NewThreeCardInteractor(mockGame, mockPresenter)

		mockGame.On("Fold").Return(nil)
		mockPresenter.On("Output", mockGame, nil).Return("fold output")

		result := ti.Fold()
		assert.Equal(t, "fold output", result)
	})
}

func TestThreeCardInteractor_ActionLog(t *testing.T) {
	mockGame := new(interfaces.MockThreeCardGame)
	mockPresenter := new(presenter.MockThreeCardPresenter)
	ti := NewThreeCardInteractor(mockGame, mockPresenter)

	mockPresenter.On("ActionLogOutput", mockGame).Return("log output")

	result := ti.ActionLog()
	assert.Equal(t, "log output", result)
}

// #5513: 直前と同じ額で賭け直す。ドメインの Rebet をそのまま通し、結果を提示する。
func TestThreeCardInteractor_Rebet(t *testing.T) {
	mockGame := new(interfaces.MockThreeCardGame)
	mockPresenter := new(presenter.MockThreeCardPresenter)
	ti := NewThreeCardInteractor(mockGame, mockPresenter)

	mockGame.On("Rebet").Return(nil)
	mockPresenter.On("Output", mockGame, nil).Return("rebet output")

	assert.Equal(t, "rebet output", ti.Rebet())
	mockGame.AssertCalled(t, "Rebet")
}

// **エラーもそのまま提示に回る。** 握りつぶすと、資金不足で断られたことが
// プレイヤーに伝わらない。
func TestThreeCardInteractor_RebetSurfacesTheError(t *testing.T) {
	mockGame := new(interfaces.MockThreeCardGame)
	mockPresenter := new(presenter.MockThreeCardPresenter)
	ti := NewThreeCardInteractor(mockGame, mockPresenter)

	wantErr := errors.New("insufficient chips")
	mockGame.On("Rebet").Return(wantErr)
	mockPresenter.On("Output", mockGame, wantErr).Return("error output")

	assert.Equal(t, "error output", ti.Rebet())
}
