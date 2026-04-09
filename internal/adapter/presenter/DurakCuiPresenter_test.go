//go:build test

package presenter_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func TestDurakCuiPresenter_Output(t *testing.T) {
	p := new(presenter.DurakCuiPresenter)

	setupGame := func() *domain.Durak {
		players := []*domain.DurakPlayer{
			domain.NewDurakPlayer(true),
			domain.NewDurakPlayer(false),
			domain.NewDurakPlayer(false),
			domain.NewDurakPlayer(false),
		}
		tc := domain.NewTrumpCardsShortDeck()
		d := domain.NewDurak(tc, players)
		d.Reset()
		return d
	}

	t.Run("initial state", func(t *testing.T) {
		d := setupGame()
		result := p.Output(d, nil)
		assert.Contains(t, result, "Durak")
		assert.Contains(t, result, "切り札")
	})

	t.Run("with error", func(t *testing.T) {
		d := setupGame()
		result := p.Output(d, domain.ErrInvalidCard)
		assert.Contains(t, result, "invalid card")
	})

	t.Run("game end human loses", func(t *testing.T) {
		d := setupGame()
		d.SetGameEndFlag(true)
		d.SetLoserIdx(0)
		d.SetPhase(domain.DurakPhaseGameEnd)
		result := p.Output(d, nil)
		assert.Contains(t, result, "ドゥラーク")
	})

	t.Run("game end CPU loses", func(t *testing.T) {
		d := setupGame()
		d.SetGameEndFlag(true)
		d.SetLoserIdx(1)
		d.SetPhase(domain.DurakPhaseGameEnd)
		result := p.Output(d, nil)
		assert.Contains(t, result, "CPU 1")
	})

	t.Run("game end draw", func(t *testing.T) {
		d := setupGame()
		d.SetGameEndFlag(true)
		d.SetLoserIdx(-1)
		d.SetPhase(domain.DurakPhaseGameEnd)
		result := p.Output(d, nil)
		assert.Contains(t, result, "引き分け")
	})

	t.Run("with table pairs", func(t *testing.T) {
		d := setupGame()
		d.SetTablePairs([]*domain.DurakTablePair{
			{Attack: domain.NewCard(domain.CardDesignSpade, 7, false)},
		})
		result := p.Output(d, nil)
		assert.Contains(t, result, "テーブル")
	})
}

func TestDurakCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.DurakCuiPresenter)
	gameMock := new(interfaces.MockDurakGame)
	gameMock.On("GetActionLog").Return([]*domain.ActionLogEntry{})
	gameMock.On("GetGameEndFlag").Return(false)

	result := p.ActionLogOutput(gameMock)
	assert.NotEmpty(t, result)
}
