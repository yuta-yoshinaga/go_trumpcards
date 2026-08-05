package presenter_test

import (
	"errors"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newPishtiForPresenter() *domain.Pishti {
	players := []*domain.PishtiPlayer{
		domain.NewPishtiPlayer(true),
		domain.NewPishtiPlayer(false),
		domain.NewPishtiPlayer(false),
		domain.NewPishtiPlayer(false),
	}
	return domain.NewPishti(domain.NewTrumpCards(0), players, domain.DefaultPishtiConfig())
}

func TestPishtiCuiPresenter_Output(t *testing.T) {
	p := new(presenter.PishtiCuiPresenter)

	t.Run("initial state includes header", func(t *testing.T) {
		g := newPishtiForPresenter()
		out := p.Output(g, nil)
		assert.Contains(t, out, "Pişti")
	})

	t.Run("pile with top card rendered", func(t *testing.T) {
		g := newPishtiForPresenter()
		g.SetPile([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 5, false)})
		out := p.Output(g, nil)
		assert.Contains(t, out, "場:")
	})

	t.Run("empty pile rendered", func(t *testing.T) {
		g := newPishtiForPresenter()
		g.SetPile([]*domain.Card{})
		out := p.Output(g, nil)
		assert.Contains(t, out, "場:")
	})

	t.Run("error displayed", func(t *testing.T) {
		g := newPishtiForPresenter()
		out := p.Output(g, errors.New("boom"))
		assert.Contains(t, out, "boom")
	})

	t.Run("game end result", func(t *testing.T) {
		g := newPishtiForPresenter()
		g.GetPlayer(0).AddCaptured([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)})
		g.SetGameEndFlag(true)
		out := p.Output(g, nil)
		assert.Contains(t, out, "ゲーム終了")
	})
}

func TestPishtiCuiPresenter_ActionLog(t *testing.T) {
	p := new(presenter.PishtiCuiPresenter)
	g := newPishtiForPresenter()
	assert.NotEmpty(t, p.ActionLogOutput(g))
}

// **対局中の優劣を数値で出す。**ピシュティ賞と捕獲枚数を別々に出すだけでは、
// 複数プレイヤー分を毎回暗算することになる (#4892)。
func TestPishtiCuiPresenter_ShowsProvisionalScores(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(false) // リーダーの強調まで見る
	defer color.SetNoColor(orig)

	p := new(presenter.PishtiCuiPresenter)
	g := newPishtiForPresenter()
	g.Reset()
	card := func() *domain.Card { return domain.NewCard(domain.CardDesignSpade, 2, false) }
	give := func(seat, n int) {
		cards := make([]*domain.Card, n)
		for i := range cards {
			cards[i] = card()
		}
		g.GetPlayer(seat).AddCaptured(cards)
	}

	// 単独リーダー → +3 が乗り、その行が強調される。
	give(0, 5)
	give(1, 3)
	out := p.Output(g, nil)
	assert.Contains(t, out, "暫定: "+strconv.Itoa(domain.PishtiScoreMostCards)+"点")
	assert.Contains(t, out, "カード点は集計時に加算")

	// **同数になったら誰にも +3 は付かない** (受け入れ条件2)。
	give(1, 2)
	tied := p.Output(g, nil)
	assert.NotContains(t, tied, "暫定: "+strconv.Itoa(domain.PishtiScoreMostCards)+"点")
	assert.Contains(t, tied, "暫定: 0点")

	// **ゲーム終了後は最終スコアに切り替わる** (受け入れ条件3)。
	g.SetGameEndFlagForTest(true)
	ended := p.Output(g, nil)
	assert.NotContains(t, ended, "暫定")
	assert.NotContains(t, ended, "カード点は集計時に加算")
}
