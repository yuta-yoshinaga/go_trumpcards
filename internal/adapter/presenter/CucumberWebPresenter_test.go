//go:build test

package presenter

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newCucumberForWeb(t *testing.T) *domain.Cucumber {
	t.Helper()
	c := domain.NewDefaultCucumber()
	c.Reset()
	return c
}

func decodeCucumber(t *testing.T, str string) map[string]any {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.Unmarshal([]byte(str), &m))
	return m
}

func TestCucumberWebPresenterOutput(t *testing.T) {
	p := new(CucumberWebPresenter)
	c := newCucumberForWeb(t)
	m := decodeCucumber(t, p.Output(c, nil))

	assert.Equal(t, float64(domain.CucumberPhasePlay), m["phase"])
	assert.Equal(t, float64(-1), m["winnerIdx"])
	assert.Equal(t, float64(-1), m["lastTrickWinnerIdx"], "まだ最終トリックは無い")
	assert.Zero(t, m["lastPenalty"])
	assert.Zero(t, m["highestInTrick"], "リードの前は基準が無い")
	assert.False(t, m["forced"].(bool))
	assert.Equal(t, float64(1), m["roundNumber"])

	players := m["players"].([]any)
	require.Len(t, players, domain.CucumberDefaultPlayerCnt)
	human := players[0].(map[string]any)
	assert.True(t, human["isHuman"].(bool))
	// **7 枚固定。** 人数で割りません。
	assert.Equal(t, float64(domain.CucumberHandSize), human["cardCount"])
	assert.Zero(t, human["penalty"])
	assert.Empty(t, players[1].(map[string]any)["cards"], "CPU の手札は伏せる")
	// リードなので全部出せる。
	assert.Len(t, m["validPlays"].([]any), domain.CucumberHandSize)
}

// **超える基準はワイヤに載る。** 盤面から数えさせません。
func TestCucumberWebPresenterCarriesTheThreshold(t *testing.T) {
	p := new(CucumberWebPresenter)
	c := newCucumberForWeb(t)
	c.SetCurrentPlayerIdxForTest(0)
	c.GiveHandForTest(0,
		domain.NewCard(domain.CardDesignSpade, 3, false),
		domain.NewCard(domain.CardDesignHeart, 10, false))
	c.SetCurrentTrickForTest([]*domain.TrickCard{
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignDiamond, 9, false)},
	})

	m := decodeCucumber(t, p.Output(c, nil))
	assert.Equal(t, float64(9), m["highestInTrick"])
	assert.Equal(t, []any{float64(1)}, m["validPlays"], "10 だけが合法")
	assert.False(t, m["forced"].(bool), "1 枚だが更新できるので forced ではない")
	assert.Equal(t, "cucumber.beat", m["messageCode"])
	assert.Equal(t, "9", m["messageParams"].(map[string]any)["n"])
}

// **合法手が 1 つ = 更新できない、ではありません。**
//
// ちょうど 1 枚だけ更新できる形があるので、その札が実際に超えるかで見分けます。
func TestCucumberWebPresenterDistinguishesForcedFromASingleBeat(t *testing.T) {
	p := new(CucumberWebPresenter)
	c := newCucumberForWeb(t)
	c.SetCurrentPlayerIdxForTest(0)
	c.GiveHandForTest(0,
		domain.NewCard(domain.CardDesignSpade, 3, false),
		domain.NewCard(domain.CardDesignHeart, 5, false))
	c.SetCurrentTrickForTest([]*domain.TrickCard{
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignDiamond, 13, false)},
	})

	m := decodeCucumber(t, p.Output(c, nil))
	assert.True(t, m["forced"].(bool))
	assert.Equal(t, []any{float64(0)}, m["validPlays"], "いちばん低い 3 だけ")
	assert.Equal(t, "cucumber.forced", m["messageCode"])
}

// **失点はラウンドに 1 回だけの出来事。** 配り直す前に載せます。
func TestCucumberWebPresenterReportsTheRoundPenalty(t *testing.T) {
	p := new(CucumberWebPresenter)
	c := newCucumberForWeb(t)
	for i := range c.GetPlayerCnt() {
		c.GiveHandForTest(i, domain.NewCard(domain.CardDesignSpade, 5+i, false))
	}
	c.SetCurrentPlayerIdxForTest(0)
	c.SetCurrentTrickForTest(nil)
	for i := range c.GetPlayerCnt() {
		require.NoError(t, c.PlayForTest(i, 0))
	}
	require.Equal(t, domain.CucumberPhaseRoundEnd, c.GetPhase())

	m := decodeCucumber(t, p.Output(c, nil))
	// 最後の席がいちばん高い札を出したので、その席に失点が付く。
	last := c.GetPlayerCnt() - 1
	assert.Equal(t, float64(last), m["lastTrickWinnerIdx"])
	assert.Equal(t, float64(5+last), m["lastPenalty"])
	assert.Equal(t, "cucumber.round.cpu", m["messageCode"])
	assert.Equal(t, float64(5+last), m["players"].([]any)[last].(map[string]any)["penalty"])
}

func TestCucumberWebPresenterMessages(t *testing.T) {
	p := new(CucumberWebPresenter)

	t.Run("エラーはそのまま返す", func(t *testing.T) {
		m := decodeCucumber(t, p.Output(newCucumberForWeb(t), errors.New("boom")))
		assert.Equal(t, "boom", m["message"])
		assert.Empty(t, m["messageCode"])
	})

	t.Run("リードはそう言う", func(t *testing.T) {
		assert.Equal(t, "cucumber.lead", decodeCucumber(t, p.Output(newCucumberForWeb(t), nil))["messageCode"])
	})

	t.Run("相手の番はそう言う", func(t *testing.T) {
		c := newCucumberForWeb(t)
		c.SetCurrentPlayerIdxForTest(1)
		assert.Equal(t, "cucumber.waiting", decodeCucumber(t, p.Output(c, nil))["messageCode"])
	})

	t.Run("投了すると相手の勝ち", func(t *testing.T) {
		c := newCucumberForWeb(t)
		c.GiveUp()
		assert.Equal(t, "cucumber.result.cpu", decodeCucumber(t, p.Output(c, nil))["messageCode"])
	})

	t.Run("失点がいちばん少なければ勝ち", func(t *testing.T) {
		c := newCucumberForWeb(t)
		for i := 1; i < c.GetPlayerCnt(); i++ {
			c.GetPlayer(i).SetPenalty(c.GetConfig().TargetScore)
		}
		c.GiveHandForTest(0, domain.NewCard(domain.CardDesignSpade, 2, false))
		for i := 1; i < c.GetPlayerCnt(); i++ {
			c.GiveHandForTest(i, domain.NewCard(domain.CardDesignHeart, 3+i, false))
		}
		c.SetCurrentPlayerIdxForTest(0)
		c.SetCurrentTrickForTest(nil)
		for i := range c.GetPlayerCnt() {
			require.NoError(t, c.PlayForTest(i, 0))
		}
		require.True(t, c.GetGameEndFlag())
		m := decodeCucumber(t, p.Output(c, nil))
		assert.Equal(t, "cucumber.result.you", m["messageCode"])
		assert.Equal(t, "0", m["messageParams"].(map[string]any)["n"])
	})
}

func TestCucumberWebPresenterHintOutput(t *testing.T) {
	p := new(CucumberWebPresenter)
	c := newCucumberForWeb(t)
	c.SetCurrentPlayerIdxForTest(0)
	c.SetCurrentTrickForTest(nil)

	hint := decodeCucumber(t, p.HintOutput(c))["hint"].(map[string]any)
	assert.Equal(t, "cucumberLead", hint["reason"])
	assert.NotNil(t, hint["cardIndex"])

	c.GiveHandForTest(0,
		domain.NewCard(domain.CardDesignSpade, 3, false),
		domain.NewCard(domain.CardDesignHeart, 10, false))
	c.SetCurrentTrickForTest([]*domain.TrickCard{
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignDiamond, 9, false)},
	})
	assert.Equal(t, "cucumberBeat", decodeCucumber(t, p.HintOutput(c))["hint"].(map[string]any)["reason"])

	// 受動ヒントは Output() でも埋まる (#4483)。
	assert.NotNil(t, decodeCucumber(t, p.Output(c, nil))["hint"])

	c.GiveUp()
	assert.Nil(t, decodeCucumber(t, p.HintOutput(c))["hint"], "終局後は助言しない")
}

// **棋譜は終局まで伏せる。**
func TestCucumberWebPresenterActionLogOutput(t *testing.T) {
	p := new(CucumberWebPresenter)
	c := newCucumberForWeb(t)
	assert.Empty(t, decodeCucumber(t, p.ActionLogOutput(c))["entries"])

	c.GiveUp()
	assert.NotEmpty(t, decodeCucumber(t, p.ActionLogOutput(c))["entries"])
}
