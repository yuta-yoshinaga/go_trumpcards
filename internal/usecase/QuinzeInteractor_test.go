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

func newMockQuinzeGame() *interfaces.MockQuinzeGame {
	return new(interfaces.MockQuinzeGame)
}

func newMockQuinzePresenter() *presenter.MockQuinzePresenter {
	return new(presenter.MockQuinzePresenter)
}

func TestNewQuinzeInteractor(t *testing.T) {
	assert.NotNil(t, NewQuinzeInteractor(newMockQuinzeGame(), newMockQuinzePresenter()))
}

func TestNewQuinzeInteractorPanicsOnNil(t *testing.T) {
	assert.Panics(t, func() { NewQuinzeInteractor(nil, newMockQuinzePresenter()) })
	assert.Panics(t, func() { NewQuinzeInteractor(newMockQuinzeGame(), nil) })
}

func TestQuinzeInteractorReset(t *testing.T) {
	g := newMockQuinzeGame()
	p := newMockQuinzePresenter()
	i := NewQuinzeInteractor(g, p)

	g.On("Reset").Return()
	p.On("Output", g, nil).Return("reset_output")

	assert.Equal(t, "reset_output", i.Reset())
	g.AssertCalled(t, "Reset")
}

// Every action forwards its arguments and routes the error to the presenter.
func TestQuinzeInteractorActions(t *testing.T) {
	cases := []struct {
		name   string
		method string
		args   []any
		call   func(*QuinzeInteractor) string
	}{
		{"bet", "PlaceBet", []any{100}, func(i *QuinzeInteractor) string { return i.Bet(100) }},
		{"deal as banker", "StartAsBanker", nil, func(i *QuinzeInteractor) string { return i.Deal() }},
		{"hit", "Hit", nil, func(i *QuinzeInteractor) string { return i.Hit() }},
		{"stand", "Stand", nil, func(i *QuinzeInteractor) string { return i.Stand() }},
		{"banker hit", "BankerHit", nil, func(i *QuinzeInteractor) string { return i.BankerHit() }},
		{"banker stand", "BankerStand", nil, func(i *QuinzeInteractor) string { return i.BankerStand() }},
	}
	for _, tc := range cases {
		t.Run(tc.name+" success", func(t *testing.T) {
			g := newMockQuinzeGame()
			p := newMockQuinzePresenter()
			i := NewQuinzeInteractor(g, p)
			g.On(tc.method, tc.args...).Return(nil)
			p.On("Output", g, nil).Return("ok_output")
			assert.Equal(t, "ok_output", tc.call(i))
			g.AssertCalled(t, tc.method, tc.args...)
		})
		t.Run(tc.name+" error", func(t *testing.T) {
			g := newMockQuinzeGame()
			p := newMockQuinzePresenter()
			i := NewQuinzeInteractor(g, p)
			err := errors.New("invalid")
			g.On(tc.method, tc.args...).Return(err)
			p.On("Output", g, err).Return("error_output")
			assert.Equal(t, "error_output", tc.call(i))
		})
	}
}

func TestQuinzeInteractorActionLog(t *testing.T) {
	g := newMockQuinzeGame()
	p := newMockQuinzePresenter()
	i := NewQuinzeInteractor(g, p)

	p.On("ActionLogOutput", mock.Anything).Return("log_output")

	assert.Equal(t, "log_output", i.ActionLog())
}

func TestQuinzeInteractorSnapshot(t *testing.T) {
	d := domain.NewDefaultQuinze()
	d.Reset()
	i := NewQuinzeInteractor(d, newMockQuinzePresenter())

	data, err := i.Snapshot()
	require.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestRestoreQuinzeInteractor(t *testing.T) {
	t.Run("round-trip preserves state", func(t *testing.T) {
		d := domain.NewDefaultQuinze()
		d.Reset()
		p := newMockQuinzePresenter()
		data, err := NewQuinzeInteractor(d, p).Snapshot()
		require.NoError(t, err)

		restored, err := RestoreQuinzeInteractor(data, p)
		require.NoError(t, err)
		assert.NotNil(t, restored)
	})

	t.Run("invalid json returns error", func(t *testing.T) {
		_, err := RestoreQuinzeInteractor([]byte("not json"), newMockQuinzePresenter())
		assert.Error(t, err)
	})
}
