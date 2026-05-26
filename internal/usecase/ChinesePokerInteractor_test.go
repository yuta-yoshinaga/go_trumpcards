//go:build test

package usecase

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func TestNewChinesePokerInteractor(t *testing.T) {
	cp := domain.NewDefaultChinesePoker()
	pp := new(presenter.MockChinesePokerPresenter)
	pp.On("Output", cp, nil).Return("{}")
	ci := NewChinesePokerInteractor(cp, pp)
	assert.NotNil(t, ci)
}

func TestNewChinesePokerInteractor_NilPanics(t *testing.T) {
	assert.Panics(t, func() { NewChinesePokerInteractor(nil, nil) })
}

func TestChinesePokerInteractor_Reset(t *testing.T) {
	cp := domain.NewDefaultChinesePoker()
	pp := new(presenter.MockChinesePokerPresenter)
	pp.On("Output", cp, nil).Return("{}")
	ci := NewChinesePokerInteractor(cp, pp)
	result := ci.Reset()
	assert.Equal(t, "{}", result)
}

func TestChinesePokerInteractor_Bet(t *testing.T) {
	cp := domain.NewDefaultChinesePoker()
	pp := new(presenter.MockChinesePokerPresenter)
	pp.On("Output", cp, nil).Return("{}")
	ci := NewChinesePokerInteractor(cp, pp)
	result := ci.Bet(100)
	assert.Contains(t, result, "{")
}

func TestChinesePokerInteractor_Snapshot(t *testing.T) {
	cp := domain.NewDefaultChinesePoker()
	pp := new(presenter.MockChinesePokerPresenter)
	pp.On("Output", cp, nil).Return("{}")
	ci := NewChinesePokerInteractor(cp, pp)
	data, err := ci.Snapshot()
	require.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestRestoreChinesePokerInteractor(t *testing.T) {
	cp := domain.NewDefaultChinesePoker()
	pp := new(presenter.MockChinesePokerPresenter)
	pp.On("Output", cp, nil).Return("{}")
	ci := NewChinesePokerInteractor(cp, pp)
	data, err := ci.Snapshot()
	require.NoError(t, err)

	pp2 := new(presenter.MockChinesePokerPresenter)
	ci2, err := RestoreChinesePokerInteractor(data, pp2)
	require.NoError(t, err)
	assert.NotNil(t, ci2)
	_ = ci
}

func TestChinesePokerInteractor_ActionLog(t *testing.T) {
	cp := domain.NewDefaultChinesePoker()
	pp := new(presenter.MockChinesePokerPresenter)
	pp.On("Output", cp, nil).Return("{}")
	pp.On("ActionLogOutput", cp).Return("[]")
	ci := NewChinesePokerInteractor(cp, pp)
	result := ci.ActionLog()
	assert.Equal(t, "[]", result)
}
