//go:build test

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

func setupBeggarMyNeighbourTest() *domain.BeggarMyNeighbour {
	tc := domain.NewTrumpCards(0)
	players := []*domain.BeggarMyNeighbourPlayer{
		domain.NewBeggarMyNeighbourPlayer(true),
		domain.NewBeggarMyNeighbourPlayer(false),
	}
	g := domain.NewBeggarMyNeighbour(tc, players, domain.DefaultBeggarMyNeighbourConfig())
	g.Reset()
	return g
}

func TestBeggarMyNeighbourCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)

	p := new(presenter.BeggarMyNeighbourCuiPresenter)

	t.Run("initial state", func(t *testing.T) {
		g := setupBeggarMyNeighbourTest()
		result := p.Output(g, nil)
		assert.Contains(t, result, "Beggar-My-Neighbour")
		assert.Contains(t, result, "CPU:")
		assert.Contains(t, result, "あなた:")
		assert.Contains(t, result, "[場の山]")
	})

	t.Run("error", func(t *testing.T) {
		g := setupBeggarMyNeighbourTest()
		result := p.Output(g, errors.New("oops"))
		assert.Contains(t, result, "oops")
	})

	t.Run("win message", func(t *testing.T) {
		g := setupBeggarMyNeighbourTest()
		data, _ := json.Marshal(g)
		var raw map[string]json.RawMessage
		_ = json.Unmarshal(data, &raw)
		raw["gf"], _ = json.Marshal(true)
		raw["wi"], _ = json.Marshal(0)
		raw["ph"], _ = json.Marshal(domain.BeggarMyNeighbourPhaseGameEnd)
		newData, _ := json.Marshal(raw)
		_ = json.Unmarshal(newData, g)

		assert.Contains(t, p.Output(g, nil), "あなたの勝ちです")
	})

	t.Run("lose message", func(t *testing.T) {
		g := setupBeggarMyNeighbourTest()
		data, _ := json.Marshal(g)
		var raw map[string]json.RawMessage
		_ = json.Unmarshal(data, &raw)
		raw["gf"], _ = json.Marshal(true)
		raw["wi"], _ = json.Marshal(1)
		raw["ph"], _ = json.Marshal(domain.BeggarMyNeighbourPhaseGameEnd)
		newData, _ := json.Marshal(raw)
		_ = json.Unmarshal(newData, g)

		assert.Contains(t, p.Output(g, nil), "CPUの勝ちです")
	})

	t.Run("draw message", func(t *testing.T) {
		g := setupBeggarMyNeighbourTest()
		data, _ := json.Marshal(g)
		var raw map[string]json.RawMessage
		_ = json.Unmarshal(data, &raw)
		raw["gf"], _ = json.Marshal(true)
		raw["wi"], _ = json.Marshal(-1)
		raw["ph"], _ = json.Marshal(domain.BeggarMyNeighbourPhaseGameEnd)
		newData, _ := json.Marshal(raw)
		_ = json.Unmarshal(newData, g)

		assert.Contains(t, p.Output(g, nil), "引き分けです")
	})

	t.Run("pay penalty phase", func(t *testing.T) {
		g := setupBeggarMyNeighbourTest()
		data, _ := json.Marshal(g)
		var raw map[string]json.RawMessage
		_ = json.Unmarshal(data, &raw)
		raw["ph"], _ = json.Marshal(domain.BeggarMyNeighbourPhasePayPenalty)
		raw["pr"], _ = json.Marshal(3)
		newData, _ := json.Marshal(raw)
		_ = json.Unmarshal(newData, g)

		assert.Contains(t, p.Output(g, nil), "ペナルティ支払い中")
	})

	t.Run("collect phase", func(t *testing.T) {
		g := setupBeggarMyNeighbourTest()
		data, _ := json.Marshal(g)
		var raw map[string]json.RawMessage
		_ = json.Unmarshal(data, &raw)
		raw["ph"], _ = json.Marshal(domain.BeggarMyNeighbourPhaseCollect)
		newData, _ := json.Marshal(raw)
		_ = json.Unmarshal(newData, g)

		assert.Contains(t, p.Output(g, nil), "ペナルティを払いきった")
	})
}

func TestBeggarMyNeighbourCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.BeggarMyNeighbourCuiPresenter)
	g := setupBeggarMyNeighbourTest()
	assert.NotEmpty(t, p.ActionLogOutput(g))
}

// **引き分け打ち切りがいつ来るかを出す。**自動進行するゲームで、あとどれだけかが
// 分からないと待つほかない (#4896)。
func TestBeggarMyNeighbourCuiPresenter_ShowsTheRoundProgress(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(false) // 強調 (太字/色) の有無まで見る
	defer color.SetNoColor(orig)

	p := new(presenter.BeggarMyNeighbourCuiPresenter)
	g := setupBeggarMyNeighbourTest()
	cfg := g.GetConfig()

	// 序盤: 数字は出るが強調しない。
	g.SetRoundsPlayedForTest(1)
	out := p.Output(g, nil)
	assert.Contains(t, out, "ラウンド: 1 / "+strconv.Itoa(cfg.MaxRounds))
	assert.NotContains(t, out, color.Yellow("ラウンド: 1 / "+strconv.Itoa(cfg.MaxRounds)))

	// **9 割を超えたら強調する** (受け入れ条件2)。
	near := cfg.MaxRounds * 9 / 10
	g.SetRoundsPlayedForTest(near)
	assert.Contains(t, p.Output(g, nil),
		color.Yellow(i18n.Tf("beggarmyneighbour.roundProgress",
			"played", strconv.Itoa(near), "max", strconv.Itoa(cfg.MaxRounds))))

	// **上限を変えたら即座に反映される** (受け入れ条件3)。
	cfg.MaxRounds = 500
	g.SetConfig(cfg)
	g.SetRoundsPlayedForTest(1)
	assert.Contains(t, p.Output(g, nil), "ラウンド: 1 / 500")
}

// #5682: Web は 52 枚のうちどちらが優勢かを色分けバーで常時出しているのに、CUI は
// 山札・捨て札・合計の生数値だけだった。**autoplay で何百ラウンドも回せるゲーム**
// なので、最終局面を CUI で追う人ほど「今どちらが勝っているか」を素早く読みたい。
func TestBeggarMyNeighbourCuiPresenter_ShowsTheShare(t *testing.T) {
	p := new(presenter.BeggarMyNeighbourCuiPresenter)
	card := func(v int) *domain.Card { return domain.NewCard(domain.CardDesignSpade, v, false) }

	seed := func(humanCards, cpuCards int) *domain.BeggarMyNeighbour {
		g := setupBeggarMyNeighbourTest()
		for _, idx := range []int{0, 1} {
			pl := g.GetPlayer(idx)
			pl.ResetPiles()
			n := humanCards
			if idx == 1 {
				n = cpuCards
			}
			for i := 0; i < n; i++ {
				pl.AddToDrawPile(card(1))
			}
		}
		return g
	}

	t.Run("reports the human share of the held cards", func(t *testing.T) {
		// 39 対 13 = 75%。
		out := p.Output(seed(39, 13), nil)

		assert.Contains(t, out, i18n.Tf("beggarmyneighbour.share", "you", "75", "cpu", "25"))
	})

	// **0 枚でも 0 除算しない** (受け入れ条件2)。両者 0 は 50/50 とする (Web と同じ)。
	t.Run("does not divide by zero when nobody holds a card", func(t *testing.T) {
		out := p.Output(seed(0, 0), nil)

		assert.Contains(t, out, i18n.Tf("beggarmyneighbour.share", "you", "50", "cpu", "50"))
	})

	t.Run("reads 100/0 once one side holds everything", func(t *testing.T) {
		out := p.Output(seed(52, 0), nil)

		assert.Contains(t, out, i18n.Tf("beggarmyneighbour.share", "you", "100", "cpu", "0"))
	})

	// **Web と同じ四捨五入。**2/3 = 66.67% を切り捨てると 66 になり、Web の
	// Math.round (67) とずれる。
	t.Run("rounds like the web page does", func(t *testing.T) {
		out := p.Output(seed(2, 1), nil)

		assert.Contains(t, out, i18n.Tf("beggarmyneighbour.share", "you", "67", "cpu", "33"))
	})

	// 既存の生数値の行は残す。
	t.Run("keeps the raw pile counts", func(t *testing.T) {
		out := p.Output(seed(39, 13), nil)

		assert.Contains(t, out, i18n.Tf("beggarmyneighbour.humanStats",
			"draw", "39", "discard", "0", "total", "39"))
	})
}
