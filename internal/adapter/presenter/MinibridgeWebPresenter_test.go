//go:build test

package presenter

import (
	"encoding/json"
	"errors"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newMinibridgeForWeb(t *testing.T) *domain.Minibridge {
	t.Helper()
	m := domain.NewDefaultMinibridge()
	m.Reset()
	return m
}

func decodeMinibridge(t *testing.T, str string) map[string]any {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.Unmarshal([]byte(str), &m))
	return m
}

func TestMinibridgeWebPresenterOutput(t *testing.T) {
	p := new(MinibridgeWebPresenter)
	m := decodeMinibridge(t, p.Output(newMinibridgeForWeb(t), nil))

	assert.Equal(t, float64(domain.MinibridgePhaseContract), m["phase"])
	assert.Equal(t, float64(1), m["roundNumber"])
	assert.Equal(t, float64(0), m["contractLevel"], "まだ選ばれていない")
	assert.Equal(t, float64(0), m["requiredTricks"])
	assert.Equal(t, float64(-1), m["winnerTeam"])
	assert.Empty(t, m["dummyHand"], "契約が決まるまでダミーは伏せる")
	assert.Equal(t, []any{float64(0), float64(0)}, m["teamScores"])

	players := m["players"].([]any)
	require.Len(t, players, domain.MinibridgePlayerCnt)
	human := players[0].(map[string]any)
	assert.True(t, human["isHuman"].(bool))
	assert.Equal(t, float64(domain.MinibridgeHandSize), human["cardCount"], "13 枚配る")
}

// **HCP は 4 席ぶんワイヤに載り、合計は必ず 40。** これが唯一の公開情報。
func TestMinibridgeWebPresenterCarriesEverySeatsHcp(t *testing.T) {
	p := new(MinibridgeWebPresenter)
	m := decodeMinibridge(t, p.Output(newMinibridgeForWeb(t), nil))

	total := 0.0
	teams := map[float64]int{}
	for _, pl := range m["players"].([]any) {
		seat := pl.(map[string]any)
		total += seat["hcp"].(float64)
		teams[seat["team"].(float64)]++
	}
	assert.Equal(t, float64(domain.MinibridgeTotalHcp), total)
	assert.Equal(t, map[float64]int{0: 2, 1: 2}, teams, "2 対 2 に分かれる")
}

// **落札者は HCP の多いペアの席。** ワイヤの HCP と一致していること。
func TestMinibridgeWebPresenterDeclarerMatchesTheAnnouncedHcp(t *testing.T) {
	p := new(MinibridgeWebPresenter)
	for range 20 {
		g := domain.NewDefaultMinibridge()
		g.Reset()
		m := decodeMinibridge(t, p.Output(g, nil))
		players := m["players"].([]any)
		decl := int(m["declarerIdx"].(float64))
		dummy := int(m["dummyIdx"].(float64))
		require.GreaterOrEqual(t, decl, 0)
		assert.Equal(t, (decl+2)%domain.MinibridgePlayerCnt, dummy, "ダミーは正面")

		declSeat := players[decl].(map[string]any)
		dummySeat := players[dummy].(map[string]any)
		assert.GreaterOrEqual(t, declSeat["hcp"].(float64), dummySeat["hcp"].(float64),
			"落札者はペアのうち HCP が多い側")
	}
}

// **契約が決まるとダミーが公開される。**
func TestMinibridgeWebPresenterRevealsTheDummyAfterTheContract(t *testing.T) {
	p := new(MinibridgeWebPresenter)
	g := newMinibridgeForWeb(t)
	decl := g.GetDeclarerIdx()
	require.NoError(t, g.SelectContractForTest(decl, 3, domain.CardDesignHeart))

	m := decodeMinibridge(t, p.Output(g, nil))
	assert.Len(t, m["dummyHand"].([]any), domain.MinibridgeHandSize)
	assert.Equal(t, float64(3), m["contractLevel"])
	assert.Equal(t, float64(domain.CardDesignHeart), m["contractSuit"])
	assert.Equal(t, float64(domain.MinibridgeBookTricks+3), m["requiredTricks"])

	dummy := m["players"].([]any)[g.GetDummyIdx()].(map[string]any)
	assert.Len(t, dummy["cards"].([]any), domain.MinibridgeHandSize, "席の手札としても公開する")
}

// **ダミーの手番では、ダミーの合法手を返す。** 席 0 固定だと何も出せない。
func TestMinibridgeWebPresenterValidPlaysFollowTheControlledSeat(t *testing.T) {
	p := new(MinibridgeWebPresenter)
	g := newMinibridgeForWeb(t)
	g.SetContractForTest(0, 3, domain.CardDesignHeart) // 席 0 = 人間が落札者
	g.SetPhaseForTest(domain.MinibridgePhasePlay)
	g.SetCurrentPlayerIdxForTest(2) // ダミーの手番

	m := decodeMinibridge(t, p.Output(g, nil))
	got := m["validPlays"].([]any)
	assert.Len(t, got, domain.MinibridgeHandSize, "ダミーの 13 枚すべてが出せる")
}

func TestMinibridgeWebPresenterMessages(t *testing.T) {
	p := new(MinibridgeWebPresenter)

	t.Run("エラーはそのまま返す", func(t *testing.T) {
		m := decodeMinibridge(t, p.Output(newMinibridgeForWeb(t), errors.New("boom")))
		assert.Equal(t, "boom", m["message"])
		assert.Empty(t, m["messageCode"])
	})

	for _, tc := range []struct {
		name string
		decl int
		want string
	}{
		{"人間が選ぶ", 0, "minibridge.contract.choose"},
		{"CPU が選ぶ", 1, "minibridge.contract.wait"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newMinibridgeForWeb(t)
			g.SetContractForTest(tc.decl, 0, 0)
			g.SetPhaseForTest(domain.MinibridgePhaseContract)
			m := decodeMinibridge(t, p.Output(g, nil))
			assert.Equal(t, tc.want, m["messageCode"])
			// **ペアの合計 HCP を出す。** 契約の大きさを決める材料。
			assert.NotEmpty(t, m["messageParams"].(map[string]any)["hcp"])
		})
	}

	t.Run("プレイ中は落札側の取得数を出す", func(t *testing.T) {
		g := newMinibridgeForWeb(t)
		g.SetContractForTest(0, 2, domain.CardDesignHeart)
		g.SetPhaseForTest(domain.MinibridgePhasePlay)
		g.GiveTricksForTest(0, 3)
		g.GiveTricksForTest(2, 2) // ダミーぶんも足す
		m := decodeMinibridge(t, p.Output(g, nil))
		assert.Equal(t, "minibridge.play", m["messageCode"])
		assert.Equal(t, "5", m["messageParams"].(map[string]any)["took"], "ペアの 2 席ぶん")
		assert.Equal(t, "8", m["messageParams"].(map[string]any)["need"])
	})

	for _, tc := range []struct {
		name  string
		level int
		took  int
		want  string
	}{
		{"成立", 2, 8, "minibridge.roundEnd.made"},
		{"失敗", 3, 5, "minibridge.roundEnd.down"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newMinibridgeForWeb(t)
			g.SetContractForTest(0, tc.level, domain.CardDesignHeart)
			g.SetPhaseForTest(domain.MinibridgePhasePlay)
			g.GiveTricksForTest(0, tc.took)
			g.GiveTricksForTest(1, domain.MinibridgeTotalTricks-tc.took)
			g.FinishRoundForTest()
			m := decodeMinibridge(t, p.Output(g, nil))
			assert.Equal(t, tc.want, m["messageCode"])
			assert.Equal(t, strconv.Itoa(domain.MinibridgeBookTricks+tc.level),
				m["messageParams"].(map[string]any)["need"])
		})
	}

	for _, tc := range []struct {
		name   string
		winner int
		want   string
	}{
		{"あなたのペア", 0, "minibridge.result.you"},
		{"相手ペア", 1, "minibridge.result.cpu"},
		{"同点", -1, "minibridge.result.tie"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newMinibridgeForWeb(t)
			if tc.winner >= 0 {
				g.SetTeamScore(tc.winner, 100)
			}
			g.FinishGameForTest()
			m := decodeMinibridge(t, p.Output(g, nil))
			assert.Equal(t, tc.want, m["messageCode"])
			assert.Equal(t, strconv.Itoa(tc.winner), m["messageParams"].(map[string]any)["idx"])
		})
	}
}

func TestMinibridgeWebPresenterHintOutput(t *testing.T) {
	p := new(MinibridgeWebPresenter)

	g := newMinibridgeForWeb(t)
	g.SetContractForTest(0, 0, 0)
	g.SetPhaseForTest(domain.MinibridgePhaseContract)
	hint := decodeMinibridge(t, p.HintOutput(g))["hint"].(map[string]any)
	assert.Equal(t, "minibridgeContract", hint["reason"])
	assert.Nil(t, hint["cardIndex"])
	assert.GreaterOrEqual(t, hint["level"].(float64), float64(1))

	g.SetContractForTest(0, 2, domain.CardDesignHeart)
	g.SetPhaseForTest(domain.MinibridgePhasePlay)
	g.SetCurrentPlayerIdxForTest(0)
	hint = decodeMinibridge(t, p.HintOutput(g))["hint"].(map[string]any)
	assert.Equal(t, "minibridgeWinTrick", hint["reason"])
	assert.NotNil(t, hint["cardIndex"])

	// 受動ヒントは Output() でも埋まる (#4483)。
	assert.NotNil(t, decodeMinibridge(t, p.Output(g, nil))["hint"])

	g.FinishGameForTest()
	assert.Nil(t, decodeMinibridge(t, p.HintOutput(g))["hint"], "終局後は助言しない")
}

// **棋譜は終局まで伏せる。**
func TestMinibridgeWebPresenterActionLogOutput(t *testing.T) {
	p := new(MinibridgeWebPresenter)
	g := newMinibridgeForWeb(t)
	assert.Empty(t, decodeMinibridge(t, p.ActionLogOutput(g))["entries"])

	g.GiveUp()
	assert.NotEmpty(t, decodeMinibridge(t, p.ActionLogOutput(g))["entries"])
}
