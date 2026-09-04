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

func TestPishtiWebPresenter_Output(t *testing.T) {
	p := new(presenter.PishtiWebPresenter)

	t.Run("initial state serialises", func(t *testing.T) {
		g := newPishtiForPresenter()
		for i := 0; i < 4; i++ {
			g.GetPlayer(i).AddCard(domain.NewCard(domain.CardDesignHeart, 5+i, false))
		}
		g.SetPile([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 9, false)})
		raw := p.Output(g, nil)
		var out controller.PishtiWebOutput
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
		g := newPishtiForPresenter()
		raw := p.Output(g, errors.New("boom"))
		var out controller.PishtiWebOutput
		require.NoError(t, json.Unmarshal([]byte(raw), &out))
		assert.Equal(t, "boom", out.Message)
	})

	t.Run("game end emits scores", func(t *testing.T) {
		g := newPishtiForPresenter()
		g.GetPlayer(0).AddCaptured([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)})
		g.SetGameEndFlag(true)
		g.SetWinners([]int{0})
		raw := p.Output(g, nil)
		var out controller.PishtiWebOutput
		require.NoError(t, json.Unmarshal([]byte(raw), &out))
		assert.True(t, out.GameEndFlag)
		assert.Equal(t, "pishti.result.scores", out.MessageCode)
		assert.Contains(t, out.MessageParams["scores"], "0:")
		assert.Equal(t, []int{0}, out.Winners)
		assert.NotEmpty(t, out.Message)
	})
}

func TestPishtiWebPresenter_ActionLog(t *testing.T) {
	p := new(presenter.PishtiWebPresenter)
	g := newPishtiForPresenter()
	assert.NotEmpty(t, p.ActionLogOutput(g))
}

// **暫定点はサーバが数える。**以前は画面が capturedCount と pistiBonus から
// 近似していて、実際の得点源 (A / J / ♣2 / ♦10) が終局まで見えなかった (#6468)。
func TestPishtiWebPresenter_ServesTheProvisionalScore(t *testing.T) {
	p := new(presenter.PishtiWebPresenter)
	g := newPishtiForPresenter()
	g.Reset()
	// 席 0 に ♦10 (=3点)。枚数の単独リーダーでもあるので +3 が乗る。
	g.GetPlayer(0).AddCaptured([]*domain.Card{domain.NewCard(domain.CardDesignDiamond, 10, false)})

	var out struct {
		Players []struct {
			ID               int `json:"id"`
			CapturedCount    int `json:"capturedCount"`
			ProvisionalScore int `json:"provisionalScore"`
		} `json:"players"`
	}
	require.NoError(t, json.Unmarshal([]byte(p.Output(g, nil)), &out))
	require.NotEmpty(t, out.Players)

	want := domain.PishtiScoreTenDiamonds + domain.PishtiScoreMostCards
	assert.Equal(t, want, out.Players[0].ProvisionalScore)
	// **枚数からは出せない値であること。**1 枚しか捕っていないのに 6 点なので、
	// capturedCount を足しただけの実装ではこの数にならない。
	assert.Equal(t, 1, out.Players[0].CapturedCount)
	// 何も捕っていない席は 0。
	assert.Equal(t, 0, out.Players[1].ProvisionalScore)
}
