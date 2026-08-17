package presenter_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func newShitheadForPresenter() *domain.Shithead {
	return domain.NewDefaultShithead()
}

func TestShitheadCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.ShitheadCuiPresenter)

	t.Run("initial reset state", func(t *testing.T) {
		s := newShitheadForPresenter()
		s.Reset()
		result := p.Output(s, nil)
		assert.Contains(t, result, "Shithead")
		assert.Contains(t, result, "山札:")
		assert.Contains(t, result, "場札: なし")
		assert.Contains(t, result, "手番:")
	})

	t.Run("facedown blind turn lists indexed slots", func(t *testing.T) {
		s := newShitheadForPresenter()
		s.Reset() // currentTurn = 0 (the human)
		// Empty the human's hand and face-up piles so the current source is the
		// face-down (blind) pile, leaving three concealed slots.
		human := s.GetPlayer(0)
		human.Reset()
		human.AddFaceDown(domain.NewCard(domain.CardDesignSpade, 1, false))
		human.AddFaceDown(domain.NewCard(domain.CardDesignHeart, 2, false))
		human.AddFaceDown(domain.NewCard(domain.CardDesignClover, 3, false))

		result := p.Output(s, nil)
		// All three blind slots are listed with concealed (??) faces.
		assert.Contains(t, result, "[0]??")
		assert.Contains(t, result, "[1]??")
		assert.Contains(t, result, "[2]??")
	})

	t.Run("error message rendered", func(t *testing.T) {
		s := newShitheadForPresenter()
		s.Reset()
		result := p.Output(s, errors.New("boom"))
		assert.Contains(t, result, "boom")
	})

	t.Run("game end rendered", func(t *testing.T) {
		s := newShitheadForPresenter()
		s.Reset()
		// Drive game until end; budget is generous to handle pickup loops.
		for i := 0; i < 50000 && !s.GetGameEndFlag(); i++ {
			if s.IsHumanTurn() {
				_ = s.PlayerPlay(nil)
			} else {
				s.CpuPlay()
			}
		}
		if !s.GetGameEndFlag() {
			t.Skip("game did not end within step budget")
		}
		result := p.Output(s, nil)
		assert.Contains(t, result, "ゲーム終了")
	})
}

func TestShitheadCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.ShitheadCuiPresenter)
	s := newShitheadForPresenter()
	s.Reset()
	out := p.ActionLogOutput(s)
	assert.NotEmpty(t, out)
}

// #5577: Web は特殊札に絵文字バッジと説明を出しているのに、CUI には注記も、
// Go 側ロケールの文言すら無かった。どのランクが特殊かは設定で変わる。
func TestShitheadCuiPresenter_MarksTheMagicCards(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.ShitheadCuiPresenter)

	build := func(mutate func(*domain.ShitheadConfig)) string {
		s := newShitheadForPresenter()
		cfg := s.GetConfig()
		mutate(&cfg)
		s.SetConfig(cfg)
		s.Reset()
		human := s.GetPlayer(0)
		human.Reset()
		// 2 (特殊になりうる) と 5 (どの設定でも普通の札)。
		human.AddCard(domain.NewCard(domain.CardDesignSpade, 2, true))
		human.AddCard(domain.NewCard(domain.CardDesignHeart, 5, true))
		return p.Output(s, nil)
	}

	on := build(func(c *domain.ShitheadConfig) { c.MagicTwo = true })
	mark := i18n.T("shithead.magicMark")
	assert.Contains(t, on, "SPADE 2"+mark)
	// **普通の札には付けないこと。**全部に付ける実装でも「含む」検査だけなら通る。
	assert.NotContains(t, on, "HEART 5"+mark)
	// 効果も出ること。印だけでは何が起きるか分からない。
	assert.Contains(t, on, i18n.T("shithead.magicEffectTwo"))

	// 設定で無効なら何も出ない (受け入れ条件2)。
	off := build(func(c *domain.ShitheadConfig) {
		c.MagicTwo, c.MagicSeven, c.MagicEight, c.MagicTen = false, false, false, false
	})
	assert.NotContains(t, off, "SPADE 2"+mark)
	assert.NotContains(t, off, i18n.T("shithead.magicEffectTwo"))
	// 凡例ごと消えること。空の「特殊札(*): 」は意味が無い。
	assert.NotContains(t, off, strings.SplitN(i18n.T("shithead.magicLegend"), "{{", 2)[0])
}

// 有効にした効果だけが凡例に並ぶこと。全部並べる実装だと、無効な規則を教える。
func TestShitheadCuiPresenter_LegendListsOnlyTheEnabledEffects(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)

	s := newShitheadForPresenter()
	cfg := s.GetConfig()
	cfg.MagicTwo, cfg.MagicSeven, cfg.MagicEight, cfg.MagicTen = true, false, false, true
	s.SetConfig(cfg)
	s.Reset()

	out := new(presenter.ShitheadCuiPresenter).Output(s, nil)
	assert.Contains(t, out, i18n.T("shithead.magicEffectTwo"))
	assert.Contains(t, out, i18n.T("shithead.magicEffectTen"))
	assert.NotContains(t, out, i18n.T("shithead.magicEffectSeven"))
	assert.NotContains(t, out, i18n.T("shithead.magicEffectEight"))
}
