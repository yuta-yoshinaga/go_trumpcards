package presenter_test

import (
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func TestGutsCuiPresenter_OutputDeclarePhase(t *testing.T) {
	g := domain.NewDefaultGuts()
	p := new(presenter.GutsCuiPresenter)
	out := p.Output(g, nil)
	assert.NotEmpty(t, out)
	// Round line + declare prompt should be present.
	assert.Contains(t, out, "ラウンド")
	assert.Contains(t, out, "宣言")
}

func TestGutsCuiPresenter_OutputError(t *testing.T) {
	g := domain.NewDefaultGuts()
	p := new(presenter.GutsCuiPresenter)
	out := p.Output(g, errors.New("boom"))
	assert.Contains(t, out, "boom")
}

func TestGutsCuiPresenter_OutputResultLose(t *testing.T) {
	g := gutsWebResultGame() // human loses to a CPU pair (shared helper)
	p := new(presenter.GutsCuiPresenter)
	out := p.Output(g, nil)
	assert.NotEmpty(t, out)
}

func TestGutsCuiPresenter_OutputResultCarry(t *testing.T) {
	g := domain.NewDefaultGuts()
	for i := 0; i < g.GetPlayerCnt(); i++ {
		g.GetPlayer(i).SetIn(false)
	}
	g.SettleForTest()
	p := new(presenter.GutsCuiPresenter)
	out := p.Output(g, nil)
	assert.NotEmpty(t, out)
	// The carry line reports the carried-over pot and the consecutive-carry count.
	assert.Equal(t, 1, g.GetCarryCount())
	assert.Contains(t, out, strconv.Itoa(g.GetCarryPot()))
	assert.Contains(t, out, "持ち越し")
}

func TestGutsCuiPresenter_OutputGameEnd(t *testing.T) {
	g := domain.NewDefaultGuts()
	cfg := g.GetConfig()
	cfg.TargetRounds = 1
	g.SetConfig(cfg)
	g.Reset()
	require.NoError(t, g.Declare(true))
	require.True(t, g.GetGameEndFlag())
	p := new(presenter.GutsCuiPresenter)
	assert.NotEmpty(t, p.Output(g, nil))
}

func TestGutsCuiPresenter_HintOutput(t *testing.T) {
	g := domain.NewDefaultGuts()
	p := new(presenter.GutsCuiPresenter)
	assert.NotEmpty(t, p.HintOutput(g))
}

func TestGutsCuiPresenter_HintOutputNone(t *testing.T) {
	g := domain.NewDefaultGuts()
	require.NoError(t, g.Declare(true)) // result phase → GetHint returns nil
	p := new(presenter.GutsCuiPresenter)
	assert.NotEmpty(t, p.HintOutput(g))
}

func TestGutsCuiPresenter_ActionLog(t *testing.T) {
	g := domain.NewDefaultGuts()
	require.NoError(t, g.Declare(true))
	p := new(presenter.GutsCuiPresenter)
	assert.NotEmpty(t, p.ActionLogOutput(g))
}

// #5697: Web は宣言フェーズで手役と勝ち目の目安を**常時**出しているのに、CUI は
// 「i か o を選んでください」だけで、同じ診断は hint コマンドを打たないと出なかった。
func TestGutsCuiPresenter_DeclareGuide(t *testing.T) {
	p := new(presenter.GutsCuiPresenter)

	handOf := func(a, b *domain.Card) *domain.Guts {
		g := domain.NewDefaultGuts()
		g.Reset()
		g.SetPhase(domain.GutsPhaseDeclare)
		human := g.GetPlayer(0)
		human.Reset()
		human.AddCard(a)
		human.AddCard(b)
		return g
	}

	t.Run("names the hand and the tier for a pair", func(t *testing.T) {
		out := p.Output(handOf(
			domain.NewCard(domain.CardDesignSpade, 9, false),
			domain.NewCard(domain.CardDesignHeart, 9, false)), nil)

		assert.Contains(t, out, i18n.Tf("guts.declareGuide",
			"hand", i18n.T("guts.guideHandPair"),
			"tier", i18n.T("guts.guideTierHigh")))
	})

	// K/A のノーペアだけが medium。CPU の in 基準 (J 以上) とは違う。
	t.Run("calls a jack-high hand low, not medium", func(t *testing.T) {
		out := p.Output(handOf(
			domain.NewCard(domain.CardDesignSpade, 11, false),
			domain.NewCard(domain.CardDesignHeart, 4, false)), nil)

		assert.Contains(t, out, i18n.Tf("guts.declareGuide",
			"hand", i18n.T("guts.guideHandHighCard"),
			"tier", i18n.T("guts.guideTierLow")))
	})

	t.Run("says nothing once the round is resolved", func(t *testing.T) {
		g := handOf(
			domain.NewCard(domain.CardDesignSpade, 9, false),
			domain.NewCard(domain.CardDesignHeart, 9, false))
		g.SetPhase(domain.GutsPhaseResult)

		out := p.Output(g, nil)

		assert.NotContains(t, out, strings.Split(i18n.T("guts.declareGuide"), "{{")[0])
	})
}
