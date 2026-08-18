package presenter_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func TestBarbuCuiPresenter_Output(t *testing.T) {
	b := domain.NewDefaultBarbu()
	b.Reset()
	p := new(presenter.BarbuCuiPresenter)
	out := p.Output(b, nil)
	assert.NotEmpty(t, out)
	assert.Contains(t, out, "1/28") // deal line
	assert.Contains(t, out, "[0]")  // human indexed hand
}

func TestBarbuCuiPresenter_HintOutput(t *testing.T) {
	p := new(presenter.BarbuCuiPresenter)

	t.Run("dominoes playable", func(t *testing.T) {
		b := domain.BarbuTestNew(domain.DefaultBarbuConfig())
		b.BarbuTestSetContract(domain.BarbuContractDominoes, -1)
		b.BarbuTestSetPhase(domain.BarbuPhasePlay)
		b.BarbuTestSetCurrentPlayer(0)
		b.BarbuTestSetHand(0, []*domain.Card{bcard(domain.CardDesignSpade, 7)})
		assert.Contains(t, p.HintOutput(b), "置けるカード")
	})

	t.Run("dominoes pass when nothing playable", func(t *testing.T) {
		b := domain.BarbuTestNew(domain.DefaultBarbuConfig())
		b.BarbuTestSetContract(domain.BarbuContractDominoes, -1)
		b.BarbuTestSetPhase(domain.BarbuPhasePlay)
		b.BarbuTestSetCurrentPlayer(0)
		// No 7 and an empty table: nothing can be placed.
		b.BarbuTestSetHand(0, []*domain.Card{bcard(domain.CardDesignSpade, 5)})
		assert.Contains(t, p.HintOutput(b), "パス")
	})

	t.Run("trick legal follow", func(t *testing.T) {
		b := domain.BarbuTestNew(domain.DefaultBarbuConfig())
		b.BarbuTestSetContract(domain.BarbuContractNoTricks, 0)
		b.BarbuTestSetPhase(domain.BarbuPhasePlay)
		b.BarbuTestSetCurrentPlayer(0)
		b.BarbuTestSetHand(0, []*domain.Card{bcard(domain.CardDesignHeart, 5), bcard(domain.CardDesignSpade, 9)})
		b.BarbuTestSetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 3, Card: bcard(domain.CardDesignHeart, 9)}})
		assert.Contains(t, p.HintOutput(b), "合法手")
	})

	t.Run("trick leading (empty trick): all cards legal", func(t *testing.T) {
		b := domain.BarbuTestNew(domain.DefaultBarbuConfig())
		b.BarbuTestSetContract(domain.BarbuContractNoTricks, 0)
		b.BarbuTestSetPhase(domain.BarbuPhasePlay)
		b.BarbuTestSetCurrentPlayer(0)
		b.BarbuTestSetHand(0, []*domain.Card{bcard(domain.CardDesignHeart, 5), bcard(domain.CardDesignSpade, 9)})
		assert.Contains(t, p.HintOutput(b), "合法手")
	})

	t.Run("trick void in lead suit: all cards legal", func(t *testing.T) {
		b := domain.BarbuTestNew(domain.DefaultBarbuConfig())
		b.BarbuTestSetContract(domain.BarbuContractNoTricks, 0)
		b.BarbuTestSetPhase(domain.BarbuPhasePlay)
		b.BarbuTestSetCurrentPlayer(0)
		// Hand has no hearts, lead is a heart -> void -> every card is legal.
		b.BarbuTestSetHand(0, []*domain.Card{bcard(domain.CardDesignSpade, 5), bcard(domain.CardDesignClover, 9)})
		b.BarbuTestSetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 3, Card: bcard(domain.CardDesignHeart, 9)}})
		assert.Contains(t, p.HintOutput(b), "合法手")
	})

	t.Run("trick with a nil lead entry falls back to all legal", func(t *testing.T) {
		b := domain.BarbuTestNew(domain.DefaultBarbuConfig())
		b.BarbuTestSetContract(domain.BarbuContractNoTricks, 0)
		b.BarbuTestSetPhase(domain.BarbuPhasePlay)
		b.BarbuTestSetCurrentPlayer(0)
		b.BarbuTestSetHand(0, []*domain.Card{bcard(domain.CardDesignSpade, 5)})
		b.BarbuTestSetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 3, Card: nil}})
		assert.Contains(t, p.HintOutput(b), "合法手")
	})

	t.Run("none outside the play phase", func(t *testing.T) {
		b := domain.BarbuTestNew(domain.DefaultBarbuConfig())
		b.BarbuTestSetPhase(domain.BarbuPhaseSelectContract)
		assert.Contains(t, p.HintOutput(b), "ヒントはありません")
	})
}

func TestBarbuCuiPresenter_ContractAndTrick(t *testing.T) {
	b := domain.BarbuTestNew(domain.DefaultBarbuConfig())
	b.BarbuTestSetContract(domain.BarbuContractTrumps, domain.CardDesignSpade)
	b.BarbuTestSetPhase(domain.BarbuPhasePlay)
	b.BarbuTestSetCurrentPlayer(0)
	b.BarbuTestSetHand(0, []*domain.Card{bcard(domain.CardDesignSpade, 5)})
	b.BarbuTestSetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 3, Card: bcard(domain.CardDesignHeart, 9)}})
	p := new(presenter.BarbuCuiPresenter)
	out := p.Output(b, nil)
	assert.NotEmpty(t, out)
}

func TestBarbuCuiPresenter_Dominoes(t *testing.T) {
	b := domain.BarbuTestNew(domain.DefaultBarbuConfig())
	b.BarbuTestSetContract(domain.BarbuContractDominoes, -1)
	b.BarbuTestSetPhase(domain.BarbuPhasePlay)
	b.BarbuTestSetCurrentPlayer(0)
	b.BarbuTestSetHand(0, []*domain.Card{bcard(domain.CardDesignSpade, 7)})
	var table [5]uint16
	table[domain.CardDesignSpade] = 1 << 7
	b.BarbuTestSetTablePlaced(table)
	p := new(presenter.BarbuCuiPresenter)
	out := p.Output(b, nil)
	assert.NotEmpty(t, out)
}

func TestBarbuCuiPresenter_ErrorAndGameEnd(t *testing.T) {
	b := domain.NewDefaultBarbu()
	b.Reset()
	p := new(presenter.BarbuCuiPresenter)
	assert.NotEmpty(t, p.Output(b, errors.New("boom")))

	b.BarbuTestSetGameEnd(true)
	out := p.Output(b, nil)
	assert.NotEmpty(t, out)
}

func TestBarbuCuiPresenter_ActionLog(t *testing.T) {
	b := domain.NewDefaultBarbu()
	b.Reset()
	p := new(presenter.BarbuCuiPresenter)
	assert.NotEmpty(t, p.ActionLogOutput(b))
}

// #5621: Web は dealHistory からディール×プレイヤーの得点表 (bb-score-matrix) を
// 出すのに、CUI は合計しか出さず、7 ディールの経過を振り返る手段が無かった。
func TestBarbuCuiPresenterShowsThePerDealBreakdown(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)

	b := domain.BarbuTestNew(domain.DefaultBarbuConfig())
	b.BarbuTestSetGameEnd(true)
	b.BarbuTestAppendDealHistory(&domain.BarbuDealDetail{
		Contract: domain.BarbuContractNoHearts, TrumpSuit: -1, DealerIdx: 0,
		Gained: map[int]int{0: -10, 1: 0, 2: -20, 3: 0},
	})
	b.BarbuTestAppendDealHistory(&domain.BarbuDealDetail{
		Contract: domain.BarbuContractTrumps, TrumpSuit: domain.CardDesignSpade, DealerIdx: 1,
		Gained: map[int]int{0: 5, 1: 15, 2: 0, 3: 10},
	})

	out := new(presenter.BarbuCuiPresenter).Output(b, nil)

	// 契約名は既存のラベルを使う (画面の他の場所と同じ表記)。
	assert.Contains(t, out, i18n.T("barbu.cNoHearts"))
	assert.Contains(t, out, i18n.T("barbu.cTrumps"))
	// **トランプ契約はスートも出す。**同じ契約でもスートで別のディールになる。
	// 表記は他の行と同じ barbuTrumpLabel 由来 (SPADE 等)。
	assert.Regexp(t, i18n.T("barbu.cTrumps")+` \S`, out)
	// 各プレイヤーの得失点。合計 (scoreEntry) とは別に、ディールごとの数字が出る。
	assert.Contains(t, out, "-20")
	assert.Contains(t, out, "15")
	// 合計行は従来どおり残っている。
	assert.Contains(t, out, i18n.T("barbu.gameEnd"))
}

// 履歴が無いうちは表そのものを出さない (空の表は「0点だった」と読める)。
func TestBarbuCuiPresenterOmitsTheBreakdownWithoutHistory(t *testing.T) {
	b := domain.BarbuTestNew(domain.DefaultBarbuConfig())
	b.BarbuTestSetGameEnd(true)

	out := new(presenter.BarbuCuiPresenter).Output(b, nil)
	assert.NotContains(t, out, i18n.T("barbu.dealBreakdownHeader"))
}
