//go:build test

package presenter

import (
	"encoding/json"
	"errors"
	"math/rand"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newGoofspielForWeb(t *testing.T) *domain.Goofspiel {
	t.Helper()
	g := domain.NewDefaultGoofspiel()
	g.SetRand(rand.New(rand.NewSource(1)))
	g.Reset()
	return g
}

func decodeGoofspiel(t *testing.T, str string) map[string]any {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.Unmarshal([]byte(str), &m))
	return m
}

func TestGoofspielWebPresenterOutput(t *testing.T) {
	p := new(GoofspielWebPresenter)
	g := newGoofspielForWeb(t)
	m := decodeGoofspiel(t, p.Output(g, nil))

	assert.Equal(t, float64(domain.GoofspielPhaseBid), m["phase"])
	assert.Equal(t, float64(-1), m["winnerIdx"])
	assert.Equal(t, float64(-1), m["lastWinnerIdx"], "まだ決着していない")
	assert.Zero(t, m["lastGained"])
	assert.Equal(t, float64(1), m["roundNumber"])
	assert.Equal(t, float64(domain.GoofspielRounds-1), m["prizeRemaining"])
	require.NotNil(t, m["currentPrize"])
	assert.Empty(t, m["carriedPrizes"])
	assert.Equal(t, m["currentPrize"].(map[string]any)["value"], m["prizeValue"])

	players := m["players"].([]any)
	require.Len(t, players, domain.GoofspielDefaultPlayerCnt)
	for i, raw := range players {
		pl := raw.(map[string]any)
		assert.Equal(t, float64(domain.GoofspielRounds), pl["cardCount"], "席 %d", i)
		assert.Zero(t, pl["score"])
		assert.False(t, pl["hasBid"].(bool))
		// **CPU の残り手札も公開します。** 使った札は場に出るので隠せていません。
		assert.Len(t, pl["cards"].([]any), domain.GoofspielRounds, "席 %d", i)
		assert.Nil(t, pl["revealedBid"], "入札はまだ公開されていない")
	}
}

// **伏せたことは見せますが、中身は公開まで見せません。**
func TestGoofspielWebPresenterHidesTheBidUntilItIsRevealed(t *testing.T) {
	p := new(GoofspielWebPresenter)
	g := newGoofspielForWeb(t)
	require.NoError(t, g.BidForTest(0, 5))

	m := decodeGoofspiel(t, p.Output(g, nil))
	human := m["players"].([]any)[0].(map[string]any)
	assert.True(t, human["hasBid"].(bool))
	assert.Nil(t, human["revealedBid"], "公開前は中身を出さない")
	assert.Equal(t, "goofspiel.waiting", m["messageCode"])
	assert.Empty(t, m["validPlays"], "伏せたあとは入札できない")

	// 全員が伏せて公開されると、中身が載る。
	require.NoError(t, g.BidForTest(1, 8))
	g.ResolveForTest()
	m = decodeGoofspiel(t, p.Output(g, nil))
	assert.Equal(t, float64(domain.GoofspielPhaseReveal), m["phase"])
	require.NotNil(t, m["players"].([]any)[0].(map[string]any)["revealedBid"])
	require.NotNil(t, m["players"].([]any)[1].(map[string]any)["revealedBid"])
}

// **最高額が賞札のランクぶん取る。**
func TestGoofspielWebPresenterReportsTheRoundWinner(t *testing.T) {
	p := new(GoofspielWebPresenter)
	g := newGoofspielForWeb(t)
	g.SetCurrentPrizeForTest(domain.NewCard(domain.GoofspielPrizeSuit(), 9, false))
	require.NoError(t, g.BidForTest(0, 4))
	require.NoError(t, g.BidForTest(1, 11))
	g.ResolveForTest()

	m := decodeGoofspiel(t, p.Output(g, nil))
	assert.Equal(t, float64(1), m["lastWinnerIdx"])
	assert.Equal(t, float64(9), m["lastGained"])
	assert.Equal(t, "goofspiel.round.cpu", m["messageCode"])
	assert.Equal(t, float64(9), m["players"].([]any)[1].(map[string]any)["score"])
}

// **同点は誰も取りません。** 勝者が居ない結果を言い分けます。
func TestGoofspielWebPresenterReportsATie(t *testing.T) {
	p := new(GoofspielWebPresenter)
	g := newGoofspielForWeb(t)
	g.SetCurrentPrizeForTest(domain.NewCard(domain.GoofspielPrizeSuit(), 9, false))
	require.NoError(t, g.BidForTest(0, 6))
	require.NoError(t, g.BidForTest(1, 6))
	g.ResolveForTest()

	m := decodeGoofspiel(t, p.Output(g, nil))
	assert.Equal(t, float64(-1), m["lastWinnerIdx"])
	assert.Zero(t, m["lastGained"])
	assert.Equal(t, "goofspiel.round.tie", m["messageCode"])
}

// **持ち越しは「今回の賞が増える」こと。** 見えないと計算が合いません。
func TestGoofspielWebPresenterCarriesThePot(t *testing.T) {
	p := new(GoofspielWebPresenter)
	cfg := domain.DefaultGoofspielConfig()
	cfg.TieRule = domain.GoofspielTieCarryOver
	g := domain.NewGoofspiel(nil, cfg)
	g.SetRand(rand.New(rand.NewSource(1)))
	g.Reset()
	g.SetCurrentPrizeForTest(domain.NewCard(domain.GoofspielPrizeSuit(), 9, false))
	require.NoError(t, g.BidForTest(0, 6))
	require.NoError(t, g.BidForTest(1, 6))
	g.ResolveForTest()
	require.NoError(t, g.NextRound())

	m := decodeGoofspiel(t, p.Output(g, nil))
	assert.Len(t, m["carriedPrizes"].([]any), 1)
	prize := m["currentPrize"].(map[string]any)["value"].(float64)
	assert.Equal(t, prize+9, m["prizeValue"], "持ち越しぶんが上乗せされる")
}

func TestGoofspielWebPresenterMessages(t *testing.T) {
	p := new(GoofspielWebPresenter)

	t.Run("エラーはそのまま返す", func(t *testing.T) {
		m := decodeGoofspiel(t, p.Output(newGoofspielForWeb(t), errors.New("boom")))
		assert.Equal(t, "boom", m["message"])
		assert.Empty(t, m["messageCode"])
	})

	// **賞札はシャッフルされます。** 具体的な数字を書くと配り依存になるので、
	// いま出ている賞札そのものと突き合わせます。
	t.Run("入札の場面は懸かっている点を出す", func(t *testing.T) {
		g := newGoofspielForWeb(t)
		m := decodeGoofspiel(t, p.Output(g, nil))
		assert.Equal(t, "goofspiel.bid", m["messageCode"])
		want := strconv.Itoa(int(m["currentPrize"].(map[string]any)["value"].(float64)))
		assert.Equal(t, want, m["messageParams"].(map[string]any)["n"])
	})

	t.Run("投了すると相手の勝ち", func(t *testing.T) {
		g := newGoofspielForWeb(t)
		g.GiveUp()
		assert.Equal(t, "goofspiel.result.cpu", decodeGoofspiel(t, p.Output(g, nil))["messageCode"])
	})

	t.Run("いちばん多く取ったら勝ち", func(t *testing.T) {
		g := newGoofspielForWeb(t)
		for !g.GetGameEndFlag() {
			require.NoError(t, g.PlayerBid(g.GetPlayer(0).GetCardsSize()-1))
			if g.GetPhase() == domain.GoofspielPhaseReveal && !g.GetGameEndFlag() {
				require.NoError(t, g.NextRound())
			}
		}
		m := decodeGoofspiel(t, p.Output(g, nil))
		assert.Contains(t, []any{"goofspiel.result.you", "goofspiel.result.cpu"}, m["messageCode"])
	})
}

func TestGoofspielWebPresenterHintOutput(t *testing.T) {
	p := new(GoofspielWebPresenter)
	g := newGoofspielForWeb(t)

	hint := decodeGoofspiel(t, p.HintOutput(g))["hint"].(map[string]any)
	assert.NotNil(t, hint["cardIndex"])
	assert.NotEmpty(t, hint["reason"])

	// 受動ヒントは Output() でも埋まる (#4483)。
	assert.NotNil(t, decodeGoofspiel(t, p.Output(g, nil))["hint"])

	require.NoError(t, g.PlayerBid(0))
	assert.Nil(t, decodeGoofspiel(t, p.HintOutput(g))["hint"], "伏せたあとは助言しない")

	g.GiveUp()
	assert.Nil(t, decodeGoofspiel(t, p.HintOutput(g))["hint"], "終局後は助言しない")
}

// **棋譜は終局まで伏せる。**
func TestGoofspielWebPresenterActionLogOutput(t *testing.T) {
	p := new(GoofspielWebPresenter)
	g := newGoofspielForWeb(t)
	assert.Empty(t, decodeGoofspiel(t, p.ActionLogOutput(g))["entries"])

	g.GiveUp()
	assert.NotEmpty(t, decodeGoofspiel(t, p.ActionLogOutput(g))["entries"])
}
