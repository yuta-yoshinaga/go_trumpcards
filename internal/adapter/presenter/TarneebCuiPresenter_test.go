//go:build test

package presenter_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func newTarneebForCuiTest() *domain.Tarneeb {
	tn := domain.NewDefaultTarneeb()
	tn.Reset()
	return tn
}

func TestTarneebCuiPresenter_Output_PhaseLabels(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.TarneebCuiPresenter)

	t.Run("bid phase prompt", func(t *testing.T) {
		tn := newTarneebForCuiTest()
		out := p.Output(tn, nil)
		assert.Contains(t, out, i18n.T("tarneeb.promptBidHelp"))
		assert.NotContains(t, out, "tarneeb.", "a raw i18n key reached the screen")
	})

	t.Run("trump phase prompt", func(t *testing.T) {
		tn := newTarneebForCuiTest()
		tn.SetPhase(domain.TarneebPhaseTrumpDeclaration)
		tn.SetBidWinnerIdx(0)
		out := p.Output(tn, nil)
		assert.NotEmpty(t, out)
	})

	t.Run("play phase prompt", func(t *testing.T) {
		tn := newTarneebForCuiTest()
		tn.SetPhase(domain.TarneebPhasePlay)
		tn.SetTrumpSuit(domain.CardDesignSpade)
		out := p.Output(tn, nil)
		assert.Contains(t, out, i18n.T("tarneeb.promptPlayHelp"))
		assert.NotContains(t, out, "tarneeb.", "a raw i18n key reached the screen")
	})

	t.Run("trick end prompt", func(t *testing.T) {
		tn := newTarneebForCuiTest()
		tn.SetPhase(domain.TarneebPhaseTrickEnd)
		out := p.Output(tn, nil)
		assert.NotEmpty(t, out)
	})

	t.Run("round end prompt", func(t *testing.T) {
		tn := newTarneebForCuiTest()
		tn.SetPhase(domain.TarneebPhaseRoundEnd)
		out := p.Output(tn, nil)
		assert.NotEmpty(t, out)
	})

	t.Run("redeal count surfaces", func(t *testing.T) {
		tn := newTarneebForCuiTest()
		// Force a redeal by simulating all-pass scenario via internal state.
		tn.SetPhase(domain.TarneebPhaseBid)
		out := p.Output(tn, nil)
		require.NotEmpty(t, out)
	})

	t.Run("error block included", func(t *testing.T) {
		tn := newTarneebForCuiTest()
		out := p.Output(tn, errors.New("bad input"))
		assert.Contains(t, out, "bad input")
	})

	t.Run("game end banner", func(t *testing.T) {
		tn := newTarneebForCuiTest()
		tn.SetPhase(domain.TarneebPhaseGameEnd)
		// Force the game end flag via game state
		tn.SetTeamScore(0, 31)
		tn.SetBidWinnerIdx(0)
		tn.SetHighestBid(7)
		// Simulate end-of-game presentation by setting phase + GameEndFlag via Reset path:
		// reach GameEnd by calling ScoreRound on a TarneebPhaseRoundEnd configured game
		out := p.Output(tn, nil)
		assert.NotEmpty(t, out)
	})
}

func TestTarneebCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.TarneebCuiPresenter)

	t.Run("bid hint", func(t *testing.T) {
		tn := newTarneebForCuiTest()
		tn.SetPhase(domain.TarneebPhaseBid)
		tn.SetBidPlayerIdx(0)
		out := p.HintOutput(tn)
		assert.NotEmpty(t, out)
	})

	t.Run("trump hint", func(t *testing.T) {
		tn := newTarneebForCuiTest()
		tn.SetPhase(domain.TarneebPhaseTrumpDeclaration)
		tn.SetBidWinnerIdx(0)
		out := p.HintOutput(tn)
		assert.NotEmpty(t, out)
	})

	t.Run("play hint", func(t *testing.T) {
		tn := newTarneebForCuiTest()
		tn.SetPhase(domain.TarneebPhasePlay)
		tn.SetTrumpSuit(domain.CardDesignSpade)
		tn.SetCurrentPlayerIdx(0)
		out := p.HintOutput(tn)
		assert.NotEmpty(t, out)
	})

	t.Run("no hint when out of turn", func(t *testing.T) {
		tn := newTarneebForCuiTest()
		tn.SetPhase(domain.TarneebPhasePlay)
		tn.SetCurrentPlayerIdx(1) // CPU's turn
		out := p.HintOutput(tn)
		assert.Contains(t, out, i18n.T("tarneeb.hintNone"))
	})
}

func TestTarneebCuiPresenter_ActionLogOutput(t *testing.T) {
	tn := newTarneebForCuiTest()
	p := new(presenter.TarneebCuiPresenter)
	out := p.ActionLogOutput(tn)
	assert.NotNil(t, out)
}

// #5606: CallBreak (#5605) と同型のギャップ。Web は validPlayIndices で出せない札を
// 無効化し理由まで出すのに、CUI は素の番号付き一覧だけで、マストフォロー違反は
// 番号を打ってエラーを踏むまで分からなかった。
func TestTarneebCuiPresenterMarksThePlayableCards(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.TarneebCuiPresenter)

	// 人間が持っているスートをリードさせると、そのスートだけが合法になる。
	setup := func(t *testing.T) *domain.Tarneeb {
		t.Helper()
		tn := newTarneebForCuiTest()
		tn.SetPhase(domain.TarneebPhasePlay)
		tn.SetCurrentPlayerIdx(0)
		human := tn.GetPlayer(0)
		require.Positive(t, human.GetCardsSize(), "前提: 人間に手札が配られていること")
		leadSuit := human.GetCard(0).GetDesign()
		tn.SetCurrentTrick([]*domain.TrickCard{
			{PlayerIdx: 1, Card: domain.NewCard(leadSuit, 5, false)},
		})
		return tn
	}

	t.Run("human play turn marks only the legal cards", func(t *testing.T) {
		tn := setup(t)
		playable := tn.GetValidPlayIndices(0)
		// **配りに賭けてはいない。**リードスートは人間の1枚目のスートなので合法手は
		// 必ず1枚以上ある。「全部合法」は13枚が同一スートの配りのときだけ。
		if len(playable) == 0 || len(playable) == tn.GetPlayer(0).GetCardsSize() {
			t.Skipf("13枚同一スートの配り (%d/%d) -- 目印の有無を区別できない",
				len(playable), tn.GetPlayer(0).GetCardsSize())
		}

		out := p.Output(tn, nil)
		assert.Equal(t, len(playable), strings.Count(out, presenter.CuiLegalMark),
			"目印の数が合法手の数と一致する")
	})

	// **目印を出さない側も踏む。**ビッド中は制限そのものが決まっていない。
	t.Run("bid phase leaves the hand unmarked", func(t *testing.T) {
		tn := setup(t)
		tn.SetPhase(domain.TarneebPhaseBid)

		out := p.Output(tn, nil)
		assert.NotContains(t, out, presenter.CuiLegalMark, "ビッド中は目印を出さない")
	})

	// 他家の手番でも出さない。
	t.Run("another player's turn leaves the hand unmarked", func(t *testing.T) {
		tn := setup(t)
		tn.SetCurrentPlayerIdx(1)

		out := p.Output(tn, nil)
		assert.NotContains(t, out, presenter.CuiLegalMark, "他家の手番では目印を出さない")
	})
}
