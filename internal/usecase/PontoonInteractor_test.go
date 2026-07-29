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

func newMockPontoonGame() *interfaces.MockPontoonGame {
	return new(interfaces.MockPontoonGame)
}

func newMockPontoonPresenter() *presenter.MockPontoonPresenter {
	return new(presenter.MockPontoonPresenter)
}

func TestNewPontoonInteractor(t *testing.T) {
	assert.NotNil(t, NewPontoonInteractor(newMockPontoonGame(), newMockPontoonPresenter()))
}

func TestNewPontoonInteractorPanicsOnNil(t *testing.T) {
	assert.Panics(t, func() { NewPontoonInteractor(nil, newMockPontoonPresenter()) })
	assert.Panics(t, func() { NewPontoonInteractor(newMockPontoonGame(), nil) })
}

func TestPontoonInteractorReset(t *testing.T) {
	g := newMockPontoonGame()
	p := newMockPontoonPresenter()
	i := NewPontoonInteractor(g, p)

	g.On("Reset").Return()
	p.On("Output", g, nil).Return("reset_output")

	assert.Equal(t, "reset_output", i.Reset())
	g.AssertCalled(t, "Reset")
}

// Every action forwards its arguments and routes the error to the presenter.
func TestPontoonInteractorActions(t *testing.T) {
	cases := []struct {
		name   string
		method string
		args   []any
		call   func(*PontoonInteractor) string
	}{
		{"bet", "PlaceBet", []any{100}, func(i *PontoonInteractor) string { return i.Bet(100) }},
		{"deal as banker", "StartAsBanker", nil, func(i *PontoonInteractor) string { return i.Deal() }},
		{"stick", "Stick", nil, func(i *PontoonInteractor) string { return i.Stick() }},
		{"twist", "Twist", nil, func(i *PontoonInteractor) string { return i.Twist() }},
		{"buy", "Buy", []any{50}, func(i *PontoonInteractor) string { return i.Buy(50) }},
		{"split", "Split", nil, func(i *PontoonInteractor) string { return i.Split() }},
		{"banker twist", "BankerTwist", nil, func(i *PontoonInteractor) string { return i.BankerTwist() }},
		{"banker stay", "BankerStay", nil, func(i *PontoonInteractor) string { return i.BankerStay() }},
	}
	for _, tc := range cases {
		t.Run(tc.name+" success", func(t *testing.T) {
			g := newMockPontoonGame()
			p := newMockPontoonPresenter()
			i := NewPontoonInteractor(g, p)
			g.On(tc.method, tc.args...).Return(nil)
			p.On("Output", g, nil).Return("ok_output")
			assert.Equal(t, "ok_output", tc.call(i))
			g.AssertCalled(t, tc.method, tc.args...)
		})
		t.Run(tc.name+" error", func(t *testing.T) {
			g := newMockPontoonGame()
			p := newMockPontoonPresenter()
			i := NewPontoonInteractor(g, p)
			err := errors.New("invalid")
			g.On(tc.method, tc.args...).Return(err)
			p.On("Output", g, err).Return("error_output")
			assert.Equal(t, "error_output", tc.call(i))
		})
	}
}

func TestPontoonInteractorActionLog(t *testing.T) {
	g := newMockPontoonGame()
	p := newMockPontoonPresenter()
	i := NewPontoonInteractor(g, p)

	p.On("ActionLogOutput", mock.Anything).Return("log_output")

	assert.Equal(t, "log_output", i.ActionLog())
}

func TestPontoonInteractorSnapshot(t *testing.T) {
	d := domain.NewDefaultPontoon()
	d.Reset()
	i := NewPontoonInteractor(d, newMockPontoonPresenter())

	data, err := i.Snapshot()
	require.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestRestorePontoonInteractor(t *testing.T) {
	t.Run("round-trip preserves state", func(t *testing.T) {
		d := domain.NewDefaultPontoon()
		d.Reset()
		p := newMockPontoonPresenter()
		data, err := NewPontoonInteractor(d, p).Snapshot()
		require.NoError(t, err)

		restored, err := RestorePontoonInteractor(data, p)
		require.NoError(t, err)
		assert.NotNil(t, restored)
	})

	t.Run("invalid json returns error", func(t *testing.T) {
		_, err := RestorePontoonInteractor([]byte("not json"), newMockPontoonPresenter())
		assert.Error(t, err)
	})
}
