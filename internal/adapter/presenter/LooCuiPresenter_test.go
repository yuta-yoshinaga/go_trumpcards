package presenter_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func TestLooCuiPresenter_Output_DecidePhase(t *testing.T) {
	g := domain.NewDefaultLoo()
	g.Reset()
	g.SetDecidePlayerIdx(0)
	p := new(presenter.LooCuiPresenter)
	out := p.Output(g, nil)
	assert.NotEmpty(t, out)
	assert.Contains(t, out, "[0]") // human indexed hand
}

// **参加の損得を出す。**Web は potRisk パネルで示すのに、CUI は現在のポット額
// しか見えず、取り分もルーの負担も暗算させていた (#4921)。
func TestLooCuiPresenter_DecidePhaseShowsThePotRisk(t *testing.T) {
	g := domain.NewDefaultLoo()
	g.Reset()
	g.SetDecidePlayerIdx(0)
	g.SetPotStart(37) // 5 で割り切れない額。端数の扱いが見える。

	out := new(presenter.LooCuiPresenter).Output(g, nil)
	// 37 / 5 = 7 (切り捨て)。**全トリック取っても入るのは 35。**端数の 2 は
	// ポットに残るので、「最大 +37」は実際より多く見せることになる。
	assert.Contains(t, out, "+35")
	assert.Contains(t, out, "+7")
	assert.NotContains(t, out, "+37")
	// 一方ペナルティはポット全額。
	assert.Contains(t, out, "-37")
	assert.NotContains(t, out, "7.4")
}

// **ポットが 0 のディールでも壊れない** (受け入れ条件2)。
func TestLooCuiPresenter_DecidePhasePotRiskWithAnEmptyPot(t *testing.T) {
	g := domain.NewDefaultLoo()
	g.Reset()
	g.SetDecidePlayerIdx(0)
	g.SetPotStart(0)

	out := new(presenter.LooCuiPresenter).Output(g, nil)
	assert.Contains(t, out, "+0")
	assert.Contains(t, out, "-0")
}

// ディサイド以外のフェーズには出さない。毎画面に出ると邪魔になる。
func TestLooCuiPresenter_PotRiskIsConfinedToTheDecidePhase(t *testing.T) {
	g := domain.NewDefaultLoo()
	g.Reset()
	g.SetPotStart(37)
	g.SetPhase(domain.LooPhasePlay)
	assert.NotContains(t, new(presenter.LooCuiPresenter).Output(g, nil), "参加した場合")
}

func TestLooCuiPresenter_Output_AllPhases(t *testing.T) {
	p := new(presenter.LooCuiPresenter)

	g := domain.NewDefaultLoo()
	g.Reset()

	g.SetPhase(domain.LooPhasePlay)
	g.SetTrumpSuit(domain.CardDesignHeart)
	g.SetCurrentTurn(0)
	g.GetPlayer(0).SetPlaying(true)
	g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 3, Card: bcard(domain.CardDesignHeart, 9)}})
	assert.NotEmpty(t, p.Output(g, nil))

	g.SetPhase(domain.LooPhaseTrickEnd)
	assert.NotEmpty(t, p.Output(g, nil))

	g.SetPhase(domain.LooPhaseRoundEnd)
	assert.NotEmpty(t, p.Output(g, nil))

	// エラー出力。
	assert.NotEmpty(t, p.Output(g, errors.New("boom")))
}

func TestLooCuiPresenter_Output_RoundEndShowsLooed(t *testing.T) {
	g := domain.NewDefaultLoo()
	g.Reset()
	// Two players in; player 1 takes zero tricks and is looed.
	g.GetPlayer(0).SetPlaying(true)
	g.GetPlayer(1).SetPlaying(true)
	g.GetPlayer(2).SetPlaying(false)
	g.GetPlayer(3).SetPlaying(false)
	g.SetRoundTricks([domain.LooPlayerCnt]int{domain.LooTrickCount, 0, 0, 0})
	g.SetPhase(domain.LooPhaseRoundEnd)
	g.ScoreRound()
	require.NotNil(t, g.GetLastDealDetail())
	require.Contains(t, g.GetLastDealDetail().Looed, 1)

	p := new(presenter.LooCuiPresenter)
	out := p.Output(g, nil)
	looedPrefix := strings.SplitN(i18n.T("loo.looedList"), "{{", 2)[0]
	assert.Contains(t, out, looedPrefix)
	// Player 0 swept the pot, so their chip delta is shown with a + sign.
	assert.Contains(t, out, "+")
}

func TestLooCuiPresenter_HintOutput(t *testing.T) {
	p := new(presenter.LooCuiPresenter)

	t.Run("decide hint", func(t *testing.T) {
		g := domain.NewDefaultLoo()
		g.Reset()
		g.SetDecidePlayerIdx(0)
		assert.NotEmpty(t, p.HintOutput(g))
	})

	t.Run("play hint", func(t *testing.T) {
		g := domain.NewDefaultLoo()
		g.Reset()
		g.SetPhase(domain.LooPhasePlay)
		g.SetTrumpSuit(domain.CardDesignHeart)
		g.SetCurrentTurn(0)
		g.SetLeadPlayerIdx(0)
		g.GetPlayer(0).SetPlaying(true)
		g.GetPlayer(0).Reset()
		g.GetPlayer(0).AddCard(bcard(domain.CardDesignHeart, 1))
		g.GetPlayer(0).AddCard(bcard(domain.CardDesignSpade, 2))
		assert.NotEmpty(t, p.HintOutput(g))
	})

	t.Run("no hint outside human turn", func(t *testing.T) {
		g := domain.NewDefaultLoo()
		g.Reset()
		g.SetPhase(domain.LooPhaseTrickEnd)
		assert.NotEmpty(t, p.HintOutput(g)) // hintNone message
	})
}

func TestLooCuiPresenter_ActionLog(t *testing.T) {
	g := domain.NewDefaultLoo()
	g.Reset()
	p := new(presenter.LooCuiPresenter)
	assert.NotEmpty(t, p.ActionLogOutput(g))
}
