package presenter_test

import (
	"encoding/json"
	"errors"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func newRistikontraForPresenter() *domain.Ristikontra {
	players := []*domain.RistikontraPlayer{
		domain.NewRistikontraPlayer(true),
		domain.NewRistikontraPlayer(false),
		domain.NewRistikontraPlayer(false),
		domain.NewRistikontraPlayer(false),
	}
	return domain.NewRistikontra(domain.NewTrumpCards(0), players, domain.DefaultRistikontraConfig())
}

func TestRistikontraCuiPresenter_Output(t *testing.T) {
	p := new(presenter.RistikontraCuiPresenter)

	t.Run("initial state includes header", func(t *testing.T) {
		g := newRistikontraForPresenter()
		out := p.Output(g, nil)
		assert.Contains(t, out, "Pişti")
	})

	t.Run("pile with top card rendered", func(t *testing.T) {
		g := newRistikontraForPresenter()
		g.SetPile([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 5, false)})
		out := p.Output(g, nil)
		assert.Contains(t, out, "場:")
	})

	t.Run("empty pile rendered", func(t *testing.T) {
		g := newRistikontraForPresenter()
		g.SetPile([]*domain.Card{})
		out := p.Output(g, nil)
		assert.Contains(t, out, "場:")
	})

	t.Run("error displayed", func(t *testing.T) {
		g := newRistikontraForPresenter()
		out := p.Output(g, errors.New("boom"))
		assert.Contains(t, out, "boom")
	})

	t.Run("game end result", func(t *testing.T) {
		g := newRistikontraForPresenter()
		g.GetPlayer(0).AddCaptured([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)})
		g.SetGameEndFlag(true)
		out := p.Output(g, nil)
		assert.Contains(t, out, "ゲーム終了")
	})
}

func TestRistikontraCuiPresenter_ActionLog(t *testing.T) {
	p := new(presenter.RistikontraCuiPresenter)
	g := newRistikontraForPresenter()
	assert.NotEmpty(t, p.ActionLogOutput(g))
}

// **対局中の優劣を数値で出す。**リスティコントラ賞と捕獲枚数を別々に出すだけでは、
// 複数プレイヤー分を毎回暗算することになる (#4892)。
func TestRistikontraCuiPresenter_ShowsProvisionalScores(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(false) // リーダーの強調まで見る
	defer color.SetNoColor(orig)

	p := new(presenter.RistikontraCuiPresenter)
	g := newRistikontraForPresenter()
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
	assert.Contains(t, out, "暫定: "+strconv.Itoa(domain.RistikontraScoreMostCards)+"点")
	assert.Contains(t, out, "カード点は集計時に加算")

	// **同数になったら誰にも +3 は付かない** (受け入れ条件2)。
	give(1, 2)
	tied := p.Output(g, nil)
	assert.NotContains(t, tied, "暫定: "+strconv.Itoa(domain.RistikontraScoreMostCards)+"点")
	assert.Contains(t, tied, "暫定: 0点")

	// **ゲーム終了後は最終スコアに切り替わる** (受け入れ条件3)。
	g.SetGameEndFlagForTest(true)
	ended := p.Output(g, nil)
	assert.NotContains(t, ended, "暫定")
	assert.NotContains(t, ended, "カード点は集計時に加算")
}

// ristikontraSetField は JSON 経由で Ristikontra の内部状態を差し替える (テスト用)。
func ristikontraSetField(g *domain.Ristikontra, fields map[string]any) {
	data, _ := json.Marshal(g)
	var raw map[string]json.RawMessage
	_ = json.Unmarshal(data, &raw)
	for k, v := range fields {
		raw[k], _ = json.Marshal(v)
	}
	newData, _ := json.Marshal(raw)
	_ = json.Unmarshal(newData, g)
}

// #5672: 場を取れるのは「場のトップと同ランクの札」と「ジャック(場を総取り)」の
// 2 条件。Web は該当札にリングを付けているのに、CUI は素の手札一覧で、毎ターン
// 場のトップと自分の手札を照合させていた。
func TestRistikontraCuiPresenter_MarksCapturingCards(t *testing.T) {
	p := new(presenter.RistikontraCuiPresenter)
	card := func(d, v int) *domain.Card { return domain.NewCard(d, v, false) }

	seed := func(hand []*domain.Card, pileTop *domain.Card, turn int) *domain.Ristikontra {
		g := newRistikontraForPresenter()
		human := g.GetPlayer(0)
		human.Reset()
		for _, c := range hand {
			human.AddCard(c)
		}
		fields := map[string]any{"ct": turn}
		if pileTop != nil {
			fields["pi"] = []map[string]any{{"d": pileTop.GetDesign(), "v": pileTop.GetValue(), "o": false}}
		}
		ristikontraSetField(g, fields)
		return g
	}

	t.Run("marks a same-rank card and a jack", func(t *testing.T) {
		g := seed([]*domain.Card{
			card(domain.CardDesignSpade, 7),  // 0: 場のトップと同ランク
			card(domain.CardDesignHeart, 11), // 1: ジャックは常に総取り
			card(domain.CardDesignClover, 3), // 2: 取れない
		}, card(domain.CardDesignDiamond, 7), 0)

		out := p.Output(g, nil)

		assert.Contains(t, out, "[0]SPADE 7"+presenter.CuiLegalMark)
		assert.Contains(t, out, "[1]"+color.Red("HEART 11")+presenter.CuiLegalMark)
		assert.NotContains(t, out, "[2]CLOVER 3"+presenter.CuiLegalMark)
	})

	// **自分の手番でないときは出さない。**Web も isHumanTurn を条件にしている。
	t.Run("marks nothing when it is not your turn", func(t *testing.T) {
		g := seed([]*domain.Card{card(domain.CardDesignSpade, 7)}, card(domain.CardDesignDiamond, 7), 1)

		out := p.Output(g, nil)

		assert.NotContains(t, out, "[0]SPADE 7"+presenter.CuiLegalMark)
	})

	// 場が空ならジャックだけが取れる (同ランク条件が成立しない)。
	t.Run("only the jack captures onto an empty pile", func(t *testing.T) {
		g := seed([]*domain.Card{
			card(domain.CardDesignSpade, 7),
			card(domain.CardDesignHeart, 11),
		}, nil, 0)

		out := p.Output(g, nil)

		assert.NotContains(t, out, "[0]SPADE 7"+presenter.CuiLegalMark)
		assert.Contains(t, out, "[1]"+color.Red("HEART 11")+presenter.CuiLegalMark)
	})

	t.Run("explains what the mark means", func(t *testing.T) {
		g := seed([]*domain.Card{card(domain.CardDesignHeart, 11)}, nil, 0)

		assert.Contains(t, p.Output(g, nil), i18n.T("ristikontra.captureLegend"))
	})
}
