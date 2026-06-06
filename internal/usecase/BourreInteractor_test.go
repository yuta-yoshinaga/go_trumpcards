package usecase_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func TestNewBourreInteractor_NilGuards(t *testing.T) {
	bpMock := new(presenter.MockBourrePresenter)
	t.Run("panics when bg is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "BourreInteractor: bg must not be nil", func() {
			usecase.NewBourreInteractor(nil, bpMock)
		})
	})
	t.Run("panics when bp is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "BourreInteractor: bp must not be nil", func() {
			usecase.NewBourreInteractor(domain.NewDefaultBourre(), nil)
		})
	})
}

func TestBourreInteractor_RealGame(t *testing.T) {
	mockOutput := `{"players":[]}`
	bpMock := new(presenter.MockBourrePresenter)
	bpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	bi := usecase.NewBourreInteractor(domain.NewDefaultBourre(), bpMock)

	assert.Equal(t, mockOutput, bi.Reset())
	assert.Equal(t, mockOutput, bi.Decide(true))
	// Draw / Play は人間の手番でなければガードされ、同じ出力を返す
	assert.Equal(t, mockOutput, bi.Draw([]int{}))
	assert.Equal(t, mockOutput, bi.Play(0))
	assert.Equal(t, mockOutput, bi.NextHand())
	assert.Equal(t, mockOutput, bi.ResetWithConfig(domain.DefaultBourreConfig()))
	assert.Equal(t, domain.DefaultBourreConfig(), bi.GetConfig())
}

func TestBourreInteractor_ActionLog(t *testing.T) {
	bpMock := new(presenter.MockBourrePresenter)
	bpMock.On("ActionLogOutput", mock.Anything).Return(`{"entries":[]}`)
	bi := usecase.NewBourreInteractor(domain.NewDefaultBourre(), bpMock)
	assert.Equal(t, `{"entries":[]}`, bi.ActionLog())
}

func TestRestoreBourreInteractor(t *testing.T) {
	bpMock := new(presenter.MockBourrePresenter)
	bpMock.On("Output", mock.Anything, mock.Anything).Return(`{"players":[]}`)
	src := domain.NewDefaultBourre()
	src.Reset()
	data, err := json.Marshal(src)
	assert.NoError(t, err)
	bi, err := usecase.RestoreBourreInteractor(data, bpMock)
	assert.NoError(t, err)
	assert.NotNil(t, bi)
}
