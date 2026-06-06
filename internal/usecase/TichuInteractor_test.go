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

func newTestTichu() *domain.Tichu {
	players := []*domain.TichuPlayer{
		domain.NewTichuPlayer(true),
		domain.NewTichuPlayer(false),
		domain.NewTichuPlayer(false),
		domain.NewTichuPlayer(false),
	}
	return domain.NewTichu(domain.NewTrumpCards(domain.TichuJokerCount), players, domain.DefaultTichuConfig())
}

func TestNewTichuInteractor_NilGuards(t *testing.T) {
	tgpMock := new(presenter.MockTichuPresenter)
	t.Run("panics when tg is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "TichuInteractor: tg must not be nil", func() {
			usecase.NewTichuInteractor(nil, tgpMock)
		})
	})
	t.Run("panics when tgp is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "TichuInteractor: tgp must not be nil", func() {
			usecase.NewTichuInteractor(newTestTichu(), nil)
		})
	})
}

func TestTichuInteractor_RealGame(t *testing.T) {
	mockOutput := `{"players":[]}`
	tgpMock := new(presenter.MockTichuPresenter)
	tgpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	ti := usecase.NewTichuInteractor(newTestTichu(), tgpMock)

	assert.Equal(t, mockOutput, ti.Reset())
	assert.Equal(t, mockOutput, ti.Declare(0))
	assert.Equal(t, mockOutput, ti.Play([]int{}))
	assert.Equal(t, mockOutput, ti.ResetWithConfig(domain.DefaultTichuConfig()))
	assert.Equal(t, domain.DefaultTichuConfig(), ti.GetConfig())
}

func TestTichuInteractor_ActionLog(t *testing.T) {
	tgpMock := new(presenter.MockTichuPresenter)
	tgpMock.On("ActionLogOutput", mock.Anything).Return(`{"entries":[]}`)
	ti := usecase.NewTichuInteractor(newTestTichu(), tgpMock)
	assert.Equal(t, `{"entries":[]}`, ti.ActionLog())
}

func TestRestoreTichuInteractor(t *testing.T) {
	tgpMock := new(presenter.MockTichuPresenter)
	tgpMock.On("Output", mock.Anything, mock.Anything).Return(`{"players":[]}`)
	src := newTestTichu()
	src.Reset()
	data, err := json.Marshal(src)
	assert.NoError(t, err)
	ti, err := usecase.RestoreTichuInteractor(data, tgpMock)
	assert.NoError(t, err)
	assert.NotNil(t, ti)
}
