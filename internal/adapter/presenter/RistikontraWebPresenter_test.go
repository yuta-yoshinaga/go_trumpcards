package presenter_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestRistikontraWebPresenter_Output(t *testing.T) {
	p := new(presenter.RistikontraWebPresenter)

	t.Run("initial state serialises", func(t *testing.T) {
		g := newRistikontraForPresenter()
		for i := 0; i < 4; i++ {
			g.GetPlayer(i).AddCard(domain.NewCard(domain.CardDesignHeart, 5+i, false))
		}
		g.SetPile([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 9, false)})
		raw := p.Output(g, nil)
		var out controller.RistikontraWebOutput
		require.NoError(t, json.Unmarshal([]byte(raw), &out))
		assert.Len(t, out.Players, 4)
		assert.Equal(t, 0, out.CurrentTurn)
		assert.Equal(t, -1, out.LastCaptureIdx)
		assert.Equal(t, 4, out.Config.PlayerCnt)
		assert.Equal(t, 1, out.PileCount)
		assert.NotNil(t, out.PileTop)
		// human hand visible, CPU hands hidden
		assert.Len(t, out.Players[0].Cards, 1)
		assert.Len(t, out.Players[1].Cards, 0)
	})

	t.Run("error message", func(t *testing.T) {
		g := newRistikontraForPresenter()
		raw := p.Output(g, errors.New("boom"))
		var out controller.RistikontraWebOutput
		require.NoError(t, json.Unmarshal([]byte(raw), &out))
		assert.Equal(t, "boom", out.Message)
	})

	t.Run("game end emits scores", func(t *testing.T) {
		g := newRistikontraForPresenter()
		g.GetPlayer(0).AddCaptured([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)})
		g.SetGameEndFlag(true)
		g.SetWinners([]int{0})
		raw := p.Output(g, nil)
		var out controller.RistikontraWebOutput
		require.NoError(t, json.Unmarshal([]byte(raw), &out))
		assert.True(t, out.GameEndFlag)
		assert.Equal(t, "ristikontra.result.scores", out.MessageCode)
		assert.Contains(t, out.MessageParams["scores"], "0:")
		assert.Equal(t, []int{0}, out.Winners)
		assert.NotEmpty(t, out.Message)
	})
}

func TestRistikontraWebPresenter_ActionLog(t *testing.T) {
	p := new(presenter.RistikontraWebPresenter)
	g := newRistikontraForPresenter()
	assert.NotEmpty(t, p.ActionLogOutput(g))
}
