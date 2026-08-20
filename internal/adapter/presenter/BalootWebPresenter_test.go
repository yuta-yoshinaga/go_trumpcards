//go:build test

package presenter

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newBalootForWeb(t *testing.T) *domain.Baloot {
	t.Helper()
	b := domain.NewDefaultBaloot()
	b.Reset()
	return b
}

func decodeBaloot(t *testing.T, s string) map[string]any {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.Unmarshal([]byte(s), &m))
	return m
}

func TestBalootWebPresenterOutput(t *testing.T) {
	p := new(BalootWebPresenter)
	b := newBalootForWeb(t)

	m := decodeBaloot(t, p.Output(b, nil))

	assert.Equal(t, float64(domain.BalootPhaseDeclare), m["phase"], "配り直後は宣言フェーズ")
	assert.Equal(t, float64(domain.BalootModeNone), m["mode"], "モードはまだ決まっていない")
	assert.Equal(t, float64(1), m["roundNumber"])
	assert.Equal(t, float64(-1), m["declarerIdx"], "まだ誰も宣言していない")

	players := m["players"].([]any)
	require.Len(t, players, domain.BalootPlayerCnt)
	human := players[0].(map[string]any)
	assert.True(t, human["isHuman"].(bool))
	// **チーム番号を出す。** 向かい合う席が味方であることは盤面から読めない。
	assert.Equal(t, float64(0), human["team"])
	assert.Equal(t, float64(1), players[1].(map[string]any)["team"])
	assert.Equal(t, float64(0), players[2].(map[string]any)["team"], "0 と 2 が味方")
	assert.Empty(t, players[1].(map[string]any)["cards"], "CPU の手札は伏せる")
	// 宣言前は 5 枚しか配られていない。
	assert.Equal(t, float64(domain.BalootFirstDealSize), human["cardCount"])

	assert.Len(t, m["scores"].([]any), domain.BalootTeamCnt)
	assert.Len(t, m["roundPoints"].([]any), domain.BalootTeamCnt)
}

// **モードと切り札はワイヤに載る。** これが無いと札の強さがクライアントで
// 説明できない（Sun と Hokom で序列そのものが違う）。
func TestBalootWebPresenterModeSurfaces(t *testing.T) {
	p := new(BalootWebPresenter)

	sun := newBalootForWeb(t)
	sun.SetCurrentPlayerIdxForTest(0)
	require.NoError(t, sun.DeclareSun())
	mSun := decodeBaloot(t, p.Output(sun, nil))
	assert.Equal(t, float64(domain.BalootModeSun), mSun["mode"])
	assert.Equal(t, float64(0), mSun["trumpSuit"], "Sun に切り札は無い")
	assert.Equal(t, float64(0), mSun["declarerIdx"])

	hokom := newBalootForWeb(t)
	hokom.SetCurrentPlayerIdxForTest(0)
	require.NoError(t, hokom.DeclareHokom(domain.CardDesignDiamond))
	mHokom := decodeBaloot(t, p.Output(hokom, nil))
	assert.Equal(t, float64(domain.BalootModeHokom), mHokom["mode"])
	assert.Equal(t, float64(domain.CardDesignDiamond), mHokom["trumpSuit"])
	// 宣言が決まると 8 枚に増える。
	assert.Equal(t, float64(domain.BalootHandSize),
		mHokom["players"].([]any)[0].(map[string]any)["cardCount"])
}

// Baloot 役はプレイヤーごとに出る。
func TestBalootWebPresenterBalootSurfaces(t *testing.T) {
	p := new(BalootWebPresenter)
	b := newBalootForWeb(t)
	b.GetPlayer(0).SetHasBaloot(true)
	// **開示済みでなければ保有そのものを送らない** (#5750)。
	b.GetPlayer(0).SetBalootRevealed(true)

	human := decodeBaloot(t, p.Output(b, nil))["players"].([]any)[0].(map[string]any)
	assert.True(t, human["hasBaloot"].(bool))
	assert.True(t, human["balootRevealed"].(bool))
	assert.False(t, decodeBaloot(t, p.Output(b, nil))["players"].([]any)[1].(map[string]any)["hasBaloot"].(bool))
}

// **伏せている席の Baloot はレスポンスにも出さない** (#5750)。フロントで
// 隠すだけだと、通信を見れば分かってしまう。
func TestBalootWebPresenterWithholdsAHiddenBaloot(t *testing.T) {
	p := new(BalootWebPresenter)
	b := newBalootForWeb(t)
	b.GetPlayer(1).SetHasBaloot(true)
	b.GetPlayer(1).SetBalootRevealed(false)

	cpu := decodeBaloot(t, p.Output(b, nil))["players"].([]any)[1].(map[string]any)
	assert.False(t, cpu["hasBaloot"].(bool), "a hidden Baloot must not leak through the API")
	assert.False(t, cpu["balootRevealed"].(bool))

	// 開示されたら送る。
	b.GetPlayer(1).SetBalootRevealed(true)
	revealed := decodeBaloot(t, p.Output(b, nil))["players"].([]any)[1].(map[string]any)
	assert.True(t, revealed["hasBaloot"].(bool))
	assert.True(t, revealed["balootRevealed"].(bool))
}

// **親は見送れないので案内が変わる。** 選べない選択肢を出さない。両側を踏む。
func TestBalootWebPresenterDeclareMessages(t *testing.T) {
	p := new(BalootWebPresenter)

	t.Run("the human is the dealer", func(t *testing.T) {
		b := newBalootForWeb(t)
		b.SetDealerIdxForTest(0)
		b.SetCurrentPlayerIdxForTest(0)
		assert.Equal(t, "baloot.declare.dealerStuck", decodeBaloot(t, p.Output(b, nil))["messageCode"])
	})

	t.Run("someone else is the dealer", func(t *testing.T) {
		b := newBalootForWeb(t)
		b.SetDealerIdxForTest(2)
		b.SetCurrentPlayerIdxForTest(0)
		assert.Equal(t, "baloot.declare.choose", decodeBaloot(t, p.Output(b, nil))["messageCode"])
	})
}

func TestBalootWebPresenterRoundEndMessage(t *testing.T) {
	p := new(BalootWebPresenter)
	b := newBalootForWeb(t)
	b.SetPhaseForTest(domain.BalootPhaseRoundEnd)
	b.SetRoundPointsForTest(0, 92)
	b.SetRoundPointsForTest(1, 70)

	m := decodeBaloot(t, p.Output(b, nil))
	assert.Equal(t, "baloot.roundEnd", m["messageCode"])
	params := m["messageParams"].(map[string]any)
	assert.Equal(t, "92", params["t0"])
	assert.Equal(t, "70", params["t1"])
}

func TestBalootWebPresenterResultMessage(t *testing.T) {
	p := new(BalootWebPresenter)

	for _, tc := range []struct {
		name     string
		t0, t1   int
		wantCode string
	}{
		{"your team wins", 160, 100, "baloot.result.team0"},
		{"they win", 100, 160, "baloot.result.team1"},
		{"a tie", 152, 152, "baloot.result.tie"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			b := newBalootForWeb(t)
			b.SetScoreForTestUse(0, tc.t0)
			b.SetScoreForTestUse(1, tc.t1)
			b.FinishGameForTest()

			m := decodeBaloot(t, p.Output(b, nil))
			assert.Equal(t, tc.wantCode, m["messageCode"])
		})
	}
}

func TestBalootWebPresenterError(t *testing.T) {
	p := new(BalootWebPresenter)
	b := newBalootForWeb(t)

	m := decodeBaloot(t, p.Output(b, assert.AnError))
	assert.Equal(t, assert.AnError.Error(), m["message"])
	assert.Empty(t, m["messageCode"])
}

func TestBalootWebPresenterValidPlaysNeverNull(t *testing.T) {
	p := new(BalootWebPresenter)
	b := newBalootForWeb(t)
	b.GiveUp()

	m := decodeBaloot(t, p.Output(b, nil))
	require.NotNil(t, m["validPlays"])
	assert.IsType(t, []any{}, m["validPlays"])
}

// 宣言フェーズのヒントは札を指さず、Hokom ならスートを添える。
func TestBalootWebPresenterHintInDeclarePhase(t *testing.T) {
	p := new(BalootWebPresenter)
	b := newBalootForWeb(t)
	b.SetCurrentPlayerIdxForTest(0)

	hint, ok := decodeBaloot(t, p.HintOutput(b))["hint"].(map[string]any)
	require.True(t, ok)
	assert.Nil(t, hint["cardIndex"], "宣言では札を指さない")
	assert.Contains(t,
		[]any{"balootDeclareSun", "balootDeclareHokom", "balootPassDeclare"},
		hint["reason"])
}

// **Hokom のヒントは切り札スートまで載せる。** 「1 スートが厚い」だけでは
// どのスートを宣言すればよいか決められない。
func TestBalootWebPresenterHintCarriesTheHokomSuit(t *testing.T) {
	p := new(BalootWebPresenter)
	b := newBalootForWeb(t)
	b.SetCurrentPlayerIdxForTest(0)
	hand := b.GetPlayer(0)
	hand.Reset()
	for _, c := range []*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 11, false),
		domain.NewCard(domain.CardDesignHeart, 9, false),
		domain.NewCard(domain.CardDesignHeart, 1, false),
		domain.NewCard(domain.CardDesignSpade, 7, false),
		domain.NewCard(domain.CardDesignClover, 8, false),
	} {
		hand.AddCard(c)
	}

	hint := decodeBaloot(t, p.HintOutput(b))["hint"].(map[string]any)
	assert.Equal(t, "balootDeclareHokom", hint["reason"])
	assert.Equal(t, float64(domain.CardDesignHeart), hint["suit"])
}

func TestBalootWebPresenterHintOmittedAfterGameEnd(t *testing.T) {
	p := new(BalootWebPresenter)
	b := newBalootForWeb(t)
	b.GiveUp()
	assert.Nil(t, decodeBaloot(t, p.HintOutput(b))["hint"])
}

func TestBalootWebPresenterConfigSurfaces(t *testing.T) {
	p := new(BalootWebPresenter)
	b := newBalootForWeb(t)
	b.SetConfig(domain.BalootConfig{Target: 300})

	assert.Equal(t, float64(300), decodeBaloot(t, p.Output(b, nil))["config"].(map[string]any)["target"])
}

func TestBalootWebPresenterActionLogOutput(t *testing.T) {
	p := new(BalootWebPresenter)
	b := newBalootForWeb(t)

	var during map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.ActionLogOutput(b)), &during))
	assert.Empty(t, during["entries"], "進行中は空")

	b.GiveUp()
	var after map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.ActionLogOutput(b)), &after))
	assert.NotEmpty(t, after["entries"], "終局後は残る")
}
