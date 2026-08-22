package presenter_test

import (
	"encoding/json"
	"errors"
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
		assert.Contains(t, out, "Ristikontra")
		// クローン元のタイトルが残っていないこと。
		assert.NotContains(t, out, "Pişti")
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
// TestRistikontraCuiPresenter_ShowsProvisionalScores は、途中経過が
// **チームの獲得枚数**として出ることを見る。
//
// クローン元のピシュティは席ごとの近似スコア (確定ボーナス + 最多捕獲 +3) を
// 出していた。リスティコントラは枚数がそのまま結果なので、席には自分の
// チームの合計が入る —— パートナーが取った枚数も自分の行に乗る。
func TestRistikontraCuiPresenter_ShowsProvisionalScores(t *testing.T) {
	p := new(presenter.RistikontraCuiPresenter)
	g := domain.NewDefaultRistikontra()
	g.Reset()
	give := func(seat, n int) {
		cards := make([]*domain.Card, n)
		for i := range cards {
			cards[i] = domain.NewCard(domain.CardDesignSpade, 2, false)
		}
		g.GetPlayer(seat).AddCaptured(cards)
	}

	give(0, 5) // チーム 0
	give(1, 3) // チーム 1
	out := p.Output(g, nil)
	assert.Contains(t, out, "暫定: 5点", "席 0 はチーム 0 の合計 5 を出す")
	assert.Contains(t, out, "暫定: 3点", "席 1 はチーム 1 の合計 3 を出す")

	// **パートナーの捕獲が自分の行に乗る。**
	give(2, 4)
	partnered := p.Output(g, nil)
	assert.Contains(t, partnered, "暫定: 9点", "チーム 0 は 5+4 = 9")

	// クローン元の近似スコアの注記はもう出さない。
	assert.NotContains(t, partnered, "カード点は集計時に加算")

	// **ゲーム終了後は最終スコアに切り替わる。**
	g.SetGameEndFlagForTest(true)
	ended := p.Output(g, nil)
	assert.NotContains(t, ended, "暫定")
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

	// **印が付くのは同ランクだけ。** クローン元のピシュティはジャックを万能の
	// 捕獲札として印を付けていたが、このゲームのジャックはただの札。印を信じて
	// 出したのに取れない、という嘘の合図になる。
	t.Run("marks the same-rank card and never an off-rank jack", func(t *testing.T) {
		g := seed([]*domain.Card{
			card(domain.CardDesignSpade, 7),  // 0: 場のトップと同ランク → 取れる
			card(domain.CardDesignHeart, 11), // 1: ジャックだがランク違い → 取れない
			card(domain.CardDesignClover, 3), // 2: 取れない
		}, card(domain.CardDesignDiamond, 7), 0)

		out := p.Output(g, nil)

		assert.Contains(t, out, "[0]SPADE 7"+presenter.CuiLegalMark)
		assert.NotContains(t, out, "[1]"+color.Red("HEART 11")+presenter.CuiLegalMark)
		assert.NotContains(t, out, "[2]CLOVER 3"+presenter.CuiLegalMark)
	})

	// **自分の手番でないときは出さない。**Web も isHumanTurn を条件にしている。
	t.Run("marks nothing when it is not your turn", func(t *testing.T) {
		g := seed([]*domain.Card{card(domain.CardDesignSpade, 7)}, card(domain.CardDesignDiamond, 7), 1)

		out := p.Output(g, nil)

		assert.NotContains(t, out, "[0]SPADE 7"+presenter.CuiLegalMark)
	})

	// **場が空なら何も取れない。** ピシュティならジャックが取れたが、
	// 同ランク条件は場のトップが無いと成立しない。
	t.Run("marks nothing onto an empty pile", func(t *testing.T) {
		g := seed([]*domain.Card{
			card(domain.CardDesignSpade, 7),
			card(domain.CardDesignHeart, 11),
		}, nil, 0)

		out := p.Output(g, nil)

		assert.NotContains(t, out, "[0]SPADE 7"+presenter.CuiLegalMark)
		assert.NotContains(t, out, "[1]"+color.Red("HEART 11")+presenter.CuiLegalMark)
	})

	t.Run("explains what the mark means", func(t *testing.T) {
		g := seed([]*domain.Card{card(domain.CardDesignHeart, 7)},
			card(domain.CardDesignDiamond, 7), 0)

		assert.Contains(t, p.Output(g, nil), i18n.T("ristikontra.captureLegend"))
	})
}
