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
