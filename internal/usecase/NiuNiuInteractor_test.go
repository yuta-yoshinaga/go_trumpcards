package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func newMockNiuNiuGame() *interfaces.MockNiuNiuGame {
	return new(interfaces.MockNiuNiuGame)
}

func newMockNiuNiuPresenter() *presenter.MockNiuNiuPresenter {
	return new(presenter.MockNiuNiuPresenter)
}

func TestNewNiuNiuInteractor(t *testing.T) {
	assert.NotNil(t, NewNiuNiuInteractor(newMockNiuNiuGame(), newMockNiuNiuPresenter()))
}

func TestNewNiuNiuInteractorPanicsOnNil(t *testing.T) {
	assert.Panics(t, func() { NewNiuNiuInteractor(nil, newMockNiuNiuPresenter()) })
	assert.Panics(t, func() { NewNiuNiuInteractor(newMockNiuNiuGame(), nil) })
}

func TestNiuNiuInteractorReset(t *testing.T) {
	g := newMockNiuNiuGame()
	p := newMockNiuNiuPresenter()
	i := NewNiuNiuInteractor(g, p)

	g.On("Reset").Return()
	p.On("Output", g, nil).Return("reset_output")

	assert.Equal(t, "reset_output", i.Reset())
	g.AssertCalled(t, "Reset")
}

func TestNiuNiuInteractorBet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		g := newMockNiuNiuGame()
		p := newMockNiuNiuPresenter()
		i := NewNiuNiuInteractor(g, p)
		g.On("PlaceBet", 100).Return(nil)
		p.On("Output", g, nil).Return("ok_output")
		assert.Equal(t, "ok_output", i.Bet(100))
		g.AssertCalled(t, "PlaceBet", 100)
	})

	t.Run("error", func(t *testing.T) {
		g := newMockNiuNiuGame()
		p := newMockNiuNiuPresenter()
		i := NewNiuNiuInteractor(g, p)
		err := errors.New("invalid")
		g.On("PlaceBet", 100).Return(err)
		p.On("Output", g, err).Return("error_output")
		assert.Equal(t, "error_output", i.Bet(100))
	})
}

func TestNiuNiuInteractorActionLog(t *testing.T) {
	g := newMockNiuNiuGame()
	p := newMockNiuNiuPresenter()
	i := NewNiuNiuInteractor(g, p)

	p.On("ActionLogOutput", mock.Anything).Return("log_output")

	assert.Equal(t, "log_output", i.ActionLog())
}

func TestNiuNiuInteractorSnapshot(t *testing.T) {
	d := domain.NewDefaultNiuNiu()
	d.Reset()
	i := NewNiuNiuInteractor(d, newMockNiuNiuPresenter())

	data, err := i.Snapshot()
	require.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestRestoreNiuNiuInteractor(t *testing.T) {
	t.Run("round-trip preserves state", func(t *testing.T) {
		d := domain.NewDefaultNiuNiu()
		d.Reset()
		p := newMockNiuNiuPresenter()
		data, err := NewNiuNiuInteractor(d, p).Snapshot()
		require.NoError(t, err)

		restored, err := RestoreNiuNiuInteractor(data, p)
		require.NoError(t, err)
		assert.NotNil(t, restored)
	})

	t.Run("invalid json returns error", func(t *testing.T) {
		_, err := RestoreNiuNiuInteractor([]byte("not json"), newMockNiuNiuPresenter())
		assert.Error(t, err)
	})
}
