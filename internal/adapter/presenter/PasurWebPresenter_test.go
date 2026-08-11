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

func newPasurForWeb(t *testing.T) *domain.Pasur {
	t.Helper()
	p := domain.NewDefaultPasur()
	p.Reset()
	return p
}

func decodePasur(t *testing.T, str string) map[string]any {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.Unmarshal([]byte(str), &m))
	return m
}

func TestPasurWebPresenterOutput(t *testing.T) {
	p := new(PasurWebPresenter)
	m := decodePasur(t, p.Output(newPasurForWeb(t), nil))

	assert.Equal(t, float64(domain.PasurPhasePlay), m["phase"])
	assert.Len(t, m["table"].([]any), domain.PasurInitialTableSize, "場は 4 枚から始まる")
	assert.Equal(t, float64(1), m["packsDealt"])
	assert.Equal(t, float64(-1), m["lastCaptureIdx"])
	assert.Empty(t, m["winners"])
	assert.Equal(t, float64(domain.PasurDefaultPlayerCnt),
		m["config"].(map[string]any)["playerCnt"])

	players := m["players"].([]any)
	require.Len(t, players, domain.PasurDefaultPlayerCnt)
	human := players[0].(map[string]any)
	assert.True(t, human["isHuman"].(bool))
	assert.Equal(t, float64(domain.PasurHandSize), human["cardCount"])
	assert.Empty(t, players[1].(map[string]any)["cards"], "CPU の手札は伏せる")
}

// **11 の部分集合はページ側で作り直さない。** ワイヤに載せる。
//
// 載せた候補が**すべてドメインに受理される**ことを、候補ごとに作り直した同じ盤面で
// 確かめます（送った記録を見るだけでは合法性を見たことになりません）。
func TestPasurWebPresenterCarriesTheCaptureOptions(t *testing.T) {
	p := new(PasurWebPresenter)
	table := []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 7, false),
		domain.NewCard(domain.CardDesignHeart, 4, false),
		domain.NewCard(domain.CardDesignClover, 3, false),
	}
	hand := domain.NewCard(domain.CardDesignDiamond, 4, false)

	build := func() *domain.Pasur {
		g := newPasurForWeb(t)
		g.SetCurrentPlayerIdxForTest(0)
		g.SetTableForTest(append([]*domain.Card{}, table...))
		g.SetHumanHandForTest(hand)
		return g
	}

	options := decodePasur(t, p.Output(build(), nil))["captureOptions"].([]any)
	require.Len(t, options, 1, "手札 1 枚ぶん")
	opts := options[0].([]any)
	// ♦4 は 7 が要る: 単独の ♠7 と、♥4+♣3。
	require.Len(t, opts, 2)

	for _, opt := range opts {
		indices := make([]int, 0)
		for _, v := range opt.([]any) {
			indices = append(indices, int(v.(float64)))
		}
		assert.NoError(t, build().PlayForTest(0, 0, indices), "候補 %v は必ず合法", indices)
	}

	// **負のコントロール: 載っていない組み合わせはドメインが拒否する。**
	assert.Error(t, build().PlayForTest(0, 0, []int{1}), "♥4 だけでは 11 にならない")
}

// **絵札は同ランクだけ。** 候補にも数値の合計は出ない。
func TestPasurWebPresenterFaceCardOptionsAreSameRankOnly(t *testing.T) {
	p := new(PasurWebPresenter)
	g := newPasurForWeb(t)
	g.SetTableForTest([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 12, false),
		domain.NewCard(domain.CardDesignHeart, 5, false),
		domain.NewCard(domain.CardDesignClover, 6, false),
	})
	g.SetHumanHandForTest(domain.NewCard(domain.CardDesignDiamond, 12, false))

	options := decodePasur(t, p.Output(g, nil))["captureOptions"].([]any)
	require.Len(t, options, 1)
	opts := options[0].([]any)
	require.Len(t, opts, 1, "同ランクの 1 通りだけ")
	assert.Equal(t, []any{float64(0)}, opts[0], "♥5+♣6 = 11 は絵札の候補にならない")
}

// **スールと通常の捕獲は別に数える。**
func TestPasurWebPresenterCarriesCapturesAndSoors(t *testing.T) {
	p := new(PasurWebPresenter)
	g := newPasurForWeb(t)
	g.SetCurrentPlayerIdxForTest(0)
	g.SetTableForTest([]*domain.Card{domain.NewCard(domain.CardDesignClover, 2, false)})
	g.SetHumanHandForTest(domain.NewCard(domain.CardDesignSpade, 9, false))
	require.NoError(t, g.PlayForTest(0, 0, []int{0}))

	human := decodePasur(t, p.Output(g, nil))["players"].([]any)[0].(map[string]any)
	assert.Equal(t, float64(2), human["capturedCount"])
	assert.Equal(t, float64(1), human["soors"])
	assert.Equal(t, float64(4), human["score"], "2♣ の 2 点がスールで倍")
}

func TestPasurWebPresenterMessages(t *testing.T) {
	p := new(PasurWebPresenter)

	t.Run("エラーはそのまま返す", func(t *testing.T) {
		m := decodePasur(t, p.Output(newPasurForWeb(t), errors.New("boom")))
		assert.Equal(t, "boom", m["message"])
		assert.Empty(t, m["messageCode"])
	})

	t.Run("プレイ中は場と山札の枚数を出す", func(t *testing.T) {
		m := decodePasur(t, p.Output(newPasurForWeb(t), nil))
		assert.Equal(t, "pasur.play", m["messageCode"])
		assert.Equal(t, "4", m["messageParams"].(map[string]any)["table"])
	})

	t.Run("勝敗", func(t *testing.T) {
		g := newPasurForWeb(t)
		g.EmptyHandsForTest()
		g.DrainDeckForTest()
		g.SetTableForTest(nil)
		g.FinishGameForTest()
		m := decodePasur(t, p.Output(g, nil))
		// 誰も取っていないので全員 0 点＝同点。
		assert.Equal(t, "pasur.result.tie", m["messageCode"])
		assert.Equal(t, "4", m["messageParams"].(map[string]any)["n"])
	})

	t.Run("単独勝利", func(t *testing.T) {
		g := newPasurForWeb(t)
		g.SetCurrentPlayerIdxForTest(0)
		g.SetTableForTest([]*domain.Card{domain.NewCard(domain.CardDesignClover, 2, false)})
		g.SetHumanHandForTest(domain.NewCard(domain.CardDesignSpade, 9, false))
		require.NoError(t, g.PlayForTest(0, 0, []int{0}))
		g.EmptyHandsForTest()
		g.DrainDeckForTest()
		g.FinishGameForTest()
		m := decodePasur(t, p.Output(g, nil))
		assert.Equal(t, "pasur.result.you", m["messageCode"])
	})
}

func TestPasurWebPresenterHintOutput(t *testing.T) {
	p := new(PasurWebPresenter)
	g := newPasurForWeb(t)
	g.SetCurrentPlayerIdxForTest(0)
	g.SetTableForTest([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 7, false)})
	g.SetHumanHandForTest(domain.NewCard(domain.CardDesignDiamond, 4, false))

	hint := decodePasur(t, p.HintOutput(g))["hint"].(map[string]any)
	assert.Equal(t, "pasurSoor", hint["reason"])
	assert.NotNil(t, hint["cardIndex"])
	assert.Equal(t, []any{float64(0)}, hint["table"])

	// 受動ヒントは Output() でも埋まる (#4483)。
	assert.NotNil(t, decodePasur(t, p.Output(g, nil))["hint"])

	g.GiveUp()
	assert.Nil(t, decodePasur(t, p.HintOutput(g))["hint"], "終局後は助言しない")
}

// **棋譜は終局まで伏せる。**
func TestPasurWebPresenterActionLogOutput(t *testing.T) {
	p := new(PasurWebPresenter)
	g := newPasurForWeb(t)
	assert.Empty(t, decodePasur(t, p.ActionLogOutput(g))["entries"])

	g.GiveUp()
	assert.NotEmpty(t, decodePasur(t, p.ActionLogOutput(g))["entries"])
}
