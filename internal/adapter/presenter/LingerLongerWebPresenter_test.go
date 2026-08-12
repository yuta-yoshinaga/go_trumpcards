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

func newLingerLongerForWeb(t *testing.T) *domain.LingerLonger {
	t.Helper()
	l := domain.NewDefaultLingerLonger()
	l.Reset()
	return l
}

func decodeLingerLonger(t *testing.T, str string) map[string]any {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.Unmarshal([]byte(str), &m))
	return m
}

func TestLingerLongerWebPresenterOutput(t *testing.T) {
	p := new(LingerLongerWebPresenter)
	l := newLingerLongerForWeb(t)
	m := decodeLingerLonger(t, p.Output(l, nil))

	assert.Equal(t, float64(domain.LingerLongerPhasePlay), m["phase"])
	assert.Equal(t, float64(-1), m["winnerIdx"])
	assert.Equal(t, float64(-1), m["lastDrawIdx"], "まだ誰も補充していない")
	assert.Zero(t, m["discarded"])
	assert.Zero(t, m["eliminatedCnt"])

	players := m["players"].([]any)
	n := domain.DefaultLingerLongerConfig().PlayerCnt
	require.Len(t, players, n)
	human := players[0].(map[string]any)
	assert.True(t, human["isHuman"].(bool))
	// **配る枚数は人数と同じ。** 山札は残り全部。
	assert.Equal(t, float64(n), human["cardCount"])
	assert.Equal(t, float64(domain.LingerLongerDeckSize-n*n), m["stockSize"])
	assert.Zero(t, human["tricksWon"])
	assert.Zero(t, human["eliminatedAt"])
	assert.Empty(t, players[1].(map[string]any)["cards"], "CPU の手札は伏せる")
}

// **補充した席はワイヤに載る。** 盤面には痕跡が残らないので、
// これが無いと「なぜ相手の手札が減らないのか」が画面から分かりません。
func TestLingerLongerWebPresenterCarriesTheDraw(t *testing.T) {
	p := new(LingerLongerWebPresenter)
	l := newLingerLongerForWeb(t)
	// 席 0 が確実に取る形にして 1 トリック回す。
	l.SetLeadPlayerIdxForTest(0)
	l.SetCurrentPlayerIdxForTest(0)
	for i := range l.GetPlayerCnt() {
		l.GiveHandForTest(i, domain.NewCard(domain.CardDesignSpade, 13-i, false),
			domain.NewCard(domain.CardDesignHeart, 2, false))
	}
	for i := range l.GetPlayerCnt() {
		require.NoError(t, l.PlayForTest(i, 0))
	}

	m := decodeLingerLonger(t, p.Output(l, nil))
	assert.Equal(t, float64(0), m["lastDrawIdx"])
	assert.Equal(t, float64(1), m["players"].([]any)[0].(map[string]any)["tricksWon"])
	// **取った席だけが補充する。** 手札は 1 枚出して 1 枚引いたので元の枚数に戻る。
	assert.Equal(t, float64(2), m["players"].([]any)[0].(map[string]any)["cardCount"])
	assert.Equal(t, float64(1), m["players"].([]any)[1].(map[string]any)["cardCount"])
	// 出し切ったトリックは場から抜ける。
	assert.Equal(t, float64(l.GetPlayerCnt()), m["discarded"])
}

func TestLingerLongerWebPresenterMessages(t *testing.T) {
	p := new(LingerLongerWebPresenter)

	t.Run("エラーはそのまま返す", func(t *testing.T) {
		m := decodeLingerLonger(t, p.Output(newLingerLongerForWeb(t), errors.New("boom")))
		assert.Equal(t, "boom", m["message"])
		assert.Empty(t, m["messageCode"])
	})

	t.Run("プレイ中は山札と手札の枚数を出す", func(t *testing.T) {
		l := newLingerLongerForWeb(t)
		m := decodeLingerLonger(t, p.Output(l, nil))
		assert.Equal(t, "lingerlonger.play", m["messageCode"])
		params := m["messageParams"].(map[string]any)
		assert.Equal(t, "4", params["n"])
		assert.Equal(t, "36", params["stock"])
	})

	// **山札が尽きたら誰も補充できない。** そこから一気に脱落が進みます。
	t.Run("山札が尽きたらそう言う", func(t *testing.T) {
		l := newLingerLongerForWeb(t)
		l.DrainStockForTest()
		m := decodeLingerLonger(t, p.Output(l, nil))
		assert.Equal(t, "lingerlonger.noStock", m["messageCode"])
	})

	// **人間が脱落しても局は続く。** 打てない理由を名乗らないと操作不能に見える。
	t.Run("脱落した人間にはそう伝える", func(t *testing.T) {
		l := newLingerLongerForWeb(t)
		l.GetPlayer(0).SetEliminatedAt(1)
		l.GiveHandForTest(0)
		m := decodeLingerLonger(t, p.Output(l, nil))
		assert.Equal(t, "lingerlonger.eliminated", m["messageCode"])
	})

	t.Run("投了すると相手の勝ち", func(t *testing.T) {
		l := newLingerLongerForWeb(t)
		l.GiveUp()
		m := decodeLingerLonger(t, p.Output(l, nil))
		assert.Equal(t, "lingerlonger.result.cpu", m["messageCode"])
		assert.NotEqual(t, "0", m["messageParams"].(map[string]any)["idx"])
	})

	t.Run("最後まで持ち続けたら勝ち", func(t *testing.T) {
		l := newLingerLongerForWeb(t)
		l.DrainStockForTest()
		l.SetLeadPlayerIdxForTest(0)
		l.SetCurrentPlayerIdxForTest(0)
		// 席 0 だけ 2 枚。ほかは 1 枚で、このトリックで出し切って脱落する。
		l.GiveHandForTest(0, domain.NewCard(domain.CardDesignSpade, 13, false),
			domain.NewCard(domain.CardDesignSpade, 12, false))
		for i := 1; i < l.GetPlayerCnt(); i++ {
			l.GiveHandForTest(i, domain.NewCard(domain.CardDesignSpade, 2+i, false))
		}
		for i := range l.GetPlayerCnt() {
			require.NoError(t, l.PlayForTest(i, 0))
		}
		require.True(t, l.GetGameEndFlag())
		m := decodeLingerLonger(t, p.Output(l, nil))
		assert.Equal(t, "lingerlonger.result.you", m["messageCode"])
	})
}

func TestLingerLongerWebPresenterHintOutput(t *testing.T) {
	p := new(LingerLongerWebPresenter)
	l := newLingerLongerForWeb(t)
	l.SetCurrentPlayerIdxForTest(0)

	hint := decodeLingerLonger(t, p.HintOutput(l))["hint"].(map[string]any)
	assert.NotNil(t, hint["cardIndex"])
	assert.Contains(t, []any{"lingerlongerWinTrick", "lingerlongerDuck"}, hint["reason"])

	// **山札が空なら取っても補充は無い。** 助言の理由が変わる。
	won := newLingerLongerForWeb(t)
	won.SetLeadPlayerIdxForTest(0)
	won.SetCurrentPlayerIdxForTest(0)
	won.GiveHandForTest(0, domain.NewCard(domain.CardDesignSpade, 14, false))
	assert.Equal(t, "lingerlongerWinTrick",
		decodeLingerLonger(t, p.HintOutput(won))["hint"].(map[string]any)["reason"])
	won.DrainStockForTest()
	assert.Equal(t, "lingerlongerNoStock",
		decodeLingerLonger(t, p.HintOutput(won))["hint"].(map[string]any)["reason"])

	// 受動ヒントは Output() でも埋まる (#4483)。
	assert.NotNil(t, decodeLingerLonger(t, p.Output(won, nil))["hint"])

	l.GiveUp()
	assert.Nil(t, decodeLingerLonger(t, p.HintOutput(l))["hint"], "終局後は助言しない")
}

// **棋譜は終局まで伏せる。**
func TestLingerLongerWebPresenterActionLogOutput(t *testing.T) {
	p := new(LingerLongerWebPresenter)
	l := newLingerLongerForWeb(t)
	assert.Empty(t, decodeLingerLonger(t, p.ActionLogOutput(l))["entries"])

	l.GiveUp()
	assert.NotEmpty(t, decodeLingerLonger(t, p.ActionLogOutput(l))["entries"])
}
