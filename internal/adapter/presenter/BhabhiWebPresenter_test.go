//go:build test

package presenter

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newBhabhiForWeb(t *testing.T) *domain.Bhabhi {
	t.Helper()
	b := domain.NewDefaultBhabhi()
	b.Reset()
	return b
}

func decodeBhabhi(t *testing.T, str string) map[string]any {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.Unmarshal([]byte(str), &m))
	return m
}

func TestBhabhiWebPresenterOutput(t *testing.T) {
	p := new(BhabhiWebPresenter)
	m := decodeBhabhi(t, p.Output(newBhabhiForWeb(t), nil))

	assert.Equal(t, float64(domain.BhabhiPhasePlay), m["phase"])
	assert.Equal(t, float64(0), m["leadSuit"], "配り直後は誰もリードしていない")
	assert.Empty(t, m["pile"])
	assert.Equal(t, float64(-1), m["bhabhiIdx"], "敗者は未確定")
	assert.Equal(t, float64(-1), m["lastPickupIdx"])
	assert.False(t, m["stalemate"].(bool))
	assert.Equal(t, float64(domain.BhabhiStalemateTricks), m["stalemateTricks"])
	assert.Equal(t, float64(domain.BhabhiDefaultPlayers), m["aliveCount"])

	players := m["players"].([]any)
	require.Len(t, players, domain.BhabhiDefaultPlayers)
	human := players[0].(map[string]any)
	assert.True(t, human["isHuman"].(bool))
	assert.Equal(t, float64(-1), human["rank"], "まだ誰も上がっていない")
	assert.Equal(t, float64(0), human["pickups"])
	assert.NotEmpty(t, human["cards"], "人間の手札だけ見える")
	assert.Empty(t, players[1].(map[string]any)["cards"], "CPU の手札は伏せる")
	assert.Equal(t, float64(domain.BhabhiDeckSize/domain.BhabhiDefaultPlayers), human["cardCount"])
}

// **人数は設定で変わる。** ワイヤにもそのまま出る。
func TestBhabhiWebPresenterCarriesThePlayerCount(t *testing.T) {
	p := new(BhabhiWebPresenter)
	for _, n := range []int{domain.BhabhiMinPlayers, 5, domain.BhabhiMaxPlayers} {
		b := domain.NewDefaultBhabhi()
		b.SetConfig(domain.BhabhiConfig{PlayerCnt: n})
		b.Reset()

		m := decodeBhabhi(t, p.Output(b, nil))
		assert.Len(t, m["players"], n)
		assert.Equal(t, float64(n), m["config"].(map[string]any)["playerCnt"])
		assert.Equal(t, float64(n), m["aliveCount"])
	}
}

// **場に何枚積まれているかが、引き取りの重さそのもの。**
func TestBhabhiWebPresenterCarriesThePile(t *testing.T) {
	p := new(BhabhiWebPresenter)
	b := newBhabhiForWeb(t)
	b.SetLeadSuitForTest(domain.CardDesignHeart)
	b.SetPileForTest([]*domain.TrickCard{
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignHeart, 5, false)},
		{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignHeart, 9, false)},
	})

	m := decodeBhabhi(t, p.Output(b, nil))
	assert.Equal(t, float64(domain.CardDesignHeart), m["leadSuit"])
	assert.Len(t, m["pile"], 2)
	assert.Equal(t, "bhabhi.play.pile", m["messageCode"])
	assert.Equal(t, "2", m["messageParams"].(map[string]any)["n"])
}

func TestBhabhiWebPresenterMessageBeforeALead(t *testing.T) {
	p := new(BhabhiWebPresenter)
	assert.Equal(t, "bhabhi.play.lead", decodeBhabhi(t, p.Output(newBhabhiForWeb(t), nil))["messageCode"])
}

// **敗者が誰かで文言が変わる。** 自分が Bhabhi かどうかは一番大事な情報。
func TestBhabhiWebPresenterReportsWhoIsTheBhabhi(t *testing.T) {
	p := new(BhabhiWebPresenter)

	you := newBhabhiForWeb(t)
	you.GiveUp()
	out := decodeBhabhi(t, p.Output(you, nil))
	assert.True(t, out["gameEndFlag"].(bool))
	assert.Equal(t, float64(0), out["bhabhiIdx"])
	assert.Equal(t, "bhabhi.result.you", out["messageCode"])
	assert.False(t, out["stalemate"].(bool))
}

// **膠着で終わったことは盤面から読めない。** 別のコードで言う。
func TestBhabhiWebPresenterReportsAStalemate(t *testing.T) {
	p := new(BhabhiWebPresenter)
	b := newBhabhiForWeb(t)
	b.SetTrickNumberForTest(domain.BhabhiStalemateTricks)
	b.FinishStalemateForTest()

	out := decodeBhabhi(t, p.Output(b, nil))
	assert.True(t, out["stalemate"].(bool))
	assert.Contains(t, out["messageCode"], "stalemate")
	assert.Equal(t, "300", out["messageParams"].(map[string]any)["tricks"])
}

func TestBhabhiWebPresenterSurfacesErrors(t *testing.T) {
	p := new(BhabhiWebPresenter)
	b := newBhabhiForWeb(t)
	err := b.PlayerPlay(999)
	require.Error(t, err)

	out := decodeBhabhi(t, p.Output(b, err))
	assert.Equal(t, err.Error(), out["message"])
	assert.Empty(t, out["messageCode"], "エラー時はコードを立てない")
}

func TestBhabhiWebPresenterHintAndLog(t *testing.T) {
	p := new(BhabhiWebPresenter)
	b := newBhabhiForWeb(t)
	b.SetCurrentIdxForTest(0)

	hint := decodeBhabhi(t, p.HintOutput(b))
	require.NotNil(t, hint["hint"])
	h := hint["hint"].(map[string]any)
	assert.NotEmpty(t, h["reason"])
	idx := int(h["cardIndex"].(float64))
	assert.Contains(t, b.GetValidPlayIndices(0), idx, "勧める札は必ず合法")

	var logOut map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.ActionLogOutput(b)), &logOut))
	assert.Contains(t, logOut, "entries")
}
