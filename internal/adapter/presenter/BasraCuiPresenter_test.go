package presenter_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestBasraCuiPresenter_Output_PlayPhase(t *testing.T) {
	g := domain.NewDefaultBasra()
	g.Reset()
	g.SetCurrentTurn(0)
	p := new(presenter.BasraCuiPresenter)
	out := p.Output(g, nil)
	assert.NotEmpty(t, out)
	assert.Contains(t, out, "[0]") // human indexed hand + indexed table
}

// **どの札で何を取れるかを見せる。**Web は選択中の札が捕獲できる場札を
// リングとチェックで示すのに、CUI はヒントを叩かない限り分からなかった (#4922)。
func TestBasraCuiPresenter_AnnotatesCaptureOptions(t *testing.T) {
	g := domain.NewDefaultBasra()
	g.Reset()
	g.SetPhase(domain.BasraPhasePlay)
	g.SetCurrentTurn(0)

	human := g.GetPlayer(0)
	human.Reset()
	human.AddCard(domain.NewCard(domain.CardDesignSpade, 5, false)) // 場の 5 を取れる
	human.AddCard(domain.NewCard(domain.CardDesignHeart, 9, false)) // 何も取れない
	g.SetTableCards([]*domain.Card{
		domain.NewCard(domain.CardDesignClover, 5, false),
		domain.NewCard(domain.CardDesignDiamond, 3, false),
	})

	out := new(presenter.BasraCuiPresenter).Output(g, nil)
	// ♠5 は場[0] の ♣5 を取れる。
	assert.Contains(t, out, "[0]SPADE 5 → 場[0]")
	// 何も取れない札には注記が付かない (受け入れ条件2)。
	assert.NotContains(t, out, "[1]HEART 9 →")
}

// CPU の手番でも人間の手札の注記は出す (手札はもともと公開されている)。
// 一方、捕獲できる札が 1 枚も無ければ行そのものを出さない。
func TestBasraCuiPresenter_NoCaptureLineWhenNothingCanBeTaken(t *testing.T) {
	g := domain.NewDefaultBasra()
	g.Reset()
	g.SetPhase(domain.BasraPhasePlay)
	g.SetCurrentTurn(0)

	human := g.GetPlayer(0)
	human.Reset()
	human.AddCard(domain.NewCard(domain.CardDesignHeart, 9, false))
	g.SetTableCards([]*domain.Card{domain.NewCard(domain.CardDesignDiamond, 3, false)})

	assert.NotContains(t, new(presenter.BasraCuiPresenter).Output(g, nil), "→ 場")
}

func TestBasraCuiPresenter_Output_AllPhases(t *testing.T) {
	p := new(presenter.BasraCuiPresenter)

	g := domain.NewDefaultBasra()
	g.Reset()
	g.SetCurrentTurn(0)
	assert.NotEmpty(t, p.Output(g, nil))

	g.SetPhase(domain.BasraPhaseGameEnd)
	assert.NotEmpty(t, p.Output(g, nil))

	// エラー出力。
	assert.NotEmpty(t, p.Output(g, errors.New("boom")))

	// 空テーブル表示。
	g.SetPhase(domain.BasraPhasePlay)
	g.SetTableCards(nil)
	assert.NotEmpty(t, p.Output(g, nil))
}

func TestBasraCuiPresenter_HintOutput(t *testing.T) {
	p := new(presenter.BasraCuiPresenter)

	t.Run("capture hint", func(t *testing.T) {
		g := domain.NewDefaultBasra()
		g.Reset()
		g.SetCurrentTurn(0)
		g.SetTableCards([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 5, false)})
		g.GetPlayer(0).Reset()
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
		assert.NotEmpty(t, p.HintOutput(g))
	})

	t.Run("no hint outside human turn", func(t *testing.T) {
		g := domain.NewDefaultBasra()
		g.Reset()
		g.SetCurrentTurn(1)
		assert.NotEmpty(t, p.HintOutput(g)) // hintNone message
	})
}

func TestBasraCuiPresenter_ActionLog(t *testing.T) {
	g := domain.NewDefaultBasra()
	g.Reset()
	p := new(presenter.BasraCuiPresenter)
	assert.NotEmpty(t, p.ActionLogOutput(g))
}
