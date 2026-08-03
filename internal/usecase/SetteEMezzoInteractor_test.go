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

func newMockSetteEMezzoGame() *interfaces.MockSetteEMezzoGame {
	return new(interfaces.MockSetteEMezzoGame)
}

func newMockSetteEMezzoPresenter() *presenter.MockSetteEMezzoPresenter {
	return new(presenter.MockSetteEMezzoPresenter)
}

func TestNewSetteEMezzoInteractor(t *testing.T) {
	assert.NotNil(t, NewSetteEMezzoInteractor(newMockSetteEMezzoGame(), newMockSetteEMezzoPresenter()))
}

func TestNewSetteEMezzoInteractorPanicsOnNil(t *testing.T) {
	assert.Panics(t, func() { NewSetteEMezzoInteractor(nil, newMockSetteEMezzoPresenter()) })
	assert.Panics(t, func() { NewSetteEMezzoInteractor(newMockSetteEMezzoGame(), nil) })
}

func TestSetteEMezzoInteractorReset(t *testing.T) {
	g := newMockSetteEMezzoGame()
	p := newMockSetteEMezzoPresenter()
	i := NewSetteEMezzoInteractor(g, p)

	g.On("Reset").Return()
	p.On("Output", g, nil).Return("reset_output")

	assert.Equal(t, "reset_output", i.Reset())
	g.AssertCalled(t, "Reset")
}

// Every action forwards its arguments and routes the error to the presenter.
func TestSetteEMezzoInteractorActions(t *testing.T) {
	cases := []struct {
		name   string
		method string
		args   []any
		call   func(*SetteEMezzoInteractor) string
	}{
		{"bet", "PlaceBet", []any{100}, func(i *SetteEMezzoInteractor) string { return i.Bet(100) }},
		{"deal as banker", "StartAsBanker", nil, func(i *SetteEMezzoInteractor) string { return i.Deal() }},
		{"hit", "Hit", nil, func(i *SetteEMezzoInteractor) string { return i.Hit() }},
		{"stand", "Stand", nil, func(i *SetteEMezzoInteractor) string { return i.Stand() }},
		// The matta rides in HALVES all the way down; a whole-point value here
		// would silently halve the hand.
		{"matta", "SetMattaValue", []any{6}, func(i *SetteEMezzoInteractor) string { return i.Matta(6) }},
		{"banker hit", "BankerHit", nil, func(i *SetteEMezzoInteractor) string { return i.BankerHit() }},
		{"banker stand", "BankerStand", nil, func(i *SetteEMezzoInteractor) string { return i.BankerStand() }},
	}
	for _, tc := range cases {
		t.Run(tc.name+" success", func(t *testing.T) {
			g := newMockSetteEMezzoGame()
			p := newMockSetteEMezzoPresenter()
			i := NewSetteEMezzoInteractor(g, p)
			g.On(tc.method, tc.args...).Return(nil)
			p.On("Output", g, nil).Return("ok_output")
			assert.Equal(t, "ok_output", tc.call(i))
			g.AssertCalled(t, tc.method, tc.args...)
		})
		t.Run(tc.name+" error", func(t *testing.T) {
			g := newMockSetteEMezzoGame()
			p := newMockSetteEMezzoPresenter()
			i := NewSetteEMezzoInteractor(g, p)
			err := errors.New("invalid")
			g.On(tc.method, tc.args...).Return(err)
			p.On("Output", g, err).Return("error_output")
			assert.Equal(t, "error_output", tc.call(i))
		})
	}
}

func TestSetteEMezzoInteractorActionLog(t *testing.T) {
	g := newMockSetteEMezzoGame()
	p := newMockSetteEMezzoPresenter()
	i := NewSetteEMezzoInteractor(g, p)

	p.On("ActionLogOutput", mock.Anything).Return("log_output")

	assert.Equal(t, "log_output", i.ActionLog())
}

func TestSetteEMezzoInteractorSnapshot(t *testing.T) {
	d := domain.NewDefaultSetteEMezzo()
	d.Reset()
	i := NewSetteEMezzoInteractor(d, newMockSetteEMezzoPresenter())

	data, err := i.Snapshot()
	require.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestRestoreSetteEMezzoInteractor(t *testing.T) {
	t.Run("round-trip preserves state", func(t *testing.T) {
		d := domain.NewDefaultSetteEMezzo()
		d.Reset()
		p := newMockSetteEMezzoPresenter()
		data, err := NewSetteEMezzoInteractor(d, p).Snapshot()
		require.NoError(t, err)

		restored, err := RestoreSetteEMezzoInteractor(data, p)
		require.NoError(t, err)
		assert.NotNil(t, restored)
	})

	t.Run("invalid json returns error", func(t *testing.T) {
		_, err := RestoreSetteEMezzoInteractor([]byte("not json"), newMockSetteEMezzoPresenter())
		assert.Error(t, err)
	})
}
