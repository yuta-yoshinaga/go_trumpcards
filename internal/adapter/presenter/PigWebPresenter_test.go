//go:build test

package presenter

import (
	"encoding/json"
	"errors"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newPigForWeb(t *testing.T) *domain.Pig {
	t.Helper()
	g := domain.NewDefaultPig()
	g.SetRand(rand.New(rand.NewSource(1)))
	g.Reset()
	return g
}

func decodePig(t *testing.T, str string) map[string]any {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.Unmarshal([]byte(str), &m))
	return m
}

func TestPigWebPresenterOutput(t *testing.T) {
	p := new(PigWebPresenter)
	g := newPigForWeb(t)
	m := decodePig(t, p.Output(g, nil))

	assert.Equal(t, float64(domain.PigPhasePass), m["phase"])
	assert.Equal(t, float64(-1), m["winnerIdx"])
	assert.Equal(t, float64(-1), m["signallerIdx"], "まだ誰も合図していない")
	assert.Equal(t, float64(-1), m["roundLoserIdx"])
	assert.Equal(t, float64(1), m["roundNumber"])
	assert.Zero(t, m["passCount"])
	// **デッキは人数 × 4 枚。** 4 人なら 16 枚。
	assert.Equal(t, float64(domain.PigDeckSize(domain.PigDefaultPlayerCnt)), m["deckSize"])

	players := m["players"].([]any)
	require.Len(t, players, domain.PigDefaultPlayerCnt)
	human := players[0].(map[string]any)
	assert.True(t, human["isHuman"].(bool))
	assert.Equal(t, float64(domain.PigHandSize), human["cardCount"])
	assert.Zero(t, human["letters"])
	assert.Empty(t, human["letterWord"])
	assert.False(t, human["eliminated"].(bool))
	assert.False(t, human["hasChosenPass"].(bool))
	assert.Empty(t, players[1].(map[string]any)["cards"], "CPU の手札は伏せる")
}

// **同時に渡すので、選び終えた席が並びます。** 盤面には痕跡が残らない。
func TestPigWebPresenterCarriesWhoHasChosen(t *testing.T) {
	p := new(PigWebPresenter)
	g := newPigForWeb(t)
	require.NoError(t, g.ChoosePassForTest(0, 0))

	m := decodePig(t, p.Output(g, nil))
	assert.True(t, m["players"].([]any)[0].(map[string]any)["hasChosenPass"].(bool))
	assert.False(t, m["players"].([]any)[1].(map[string]any)["hasChosenPass"].(bool))
	assert.Equal(t, "pig.waiting", m["messageCode"], "全員が選ぶまで待つ")
	assert.Empty(t, m["validPlays"], "選んだあとは渡せる札が無い")
}

// **合図と気づいた順はワイヤに載る。** どちらも盤面に痕跡が残らない。
func TestPigWebPresenterCarriesTheSignal(t *testing.T) {
	p := new(PigWebPresenter)
	g := newPigForWeb(t)
	g.OpenSignalForTest(2)

	m := decodePig(t, p.Output(g, nil))
	assert.Equal(t, float64(domain.PigPhaseSignal), m["phase"])
	assert.Equal(t, float64(2), m["signallerIdx"])
	assert.Equal(t, float64(1), m["noticedCnt"])
	assert.Equal(t, "pig.signal", m["messageCode"])

	signaller := m["players"].([]any)[2].(map[string]any)
	assert.True(t, signaller["hasSignalled"].(bool))
	assert.Equal(t, float64(1), signaller["noticedOrder"])

	// 人間が名乗ると文言が変わる。
	require.NoError(t, g.PlayerSignal())
	m = decodePig(t, p.Output(g, nil))
	assert.Equal(t, "pig.signalDone", m["messageCode"])
}

// **ラウンドの罰は 1 回きりの出来事。** 消える前に載せる。
func TestPigWebPresenterReportsTheRoundLoser(t *testing.T) {
	p := new(PigWebPresenter)
	g := newPigForWeb(t)
	g.OpenSignalForTest(1)
	g.NoticeForTest(2)
	g.NoticeForTest(3)
	require.Equal(t, domain.PigPhaseRoundEnd, g.GetPhase())

	m := decodePig(t, p.Output(g, nil))
	assert.Equal(t, float64(0), m["roundLoserIdx"], "気づかなかったのは人間")
	assert.Equal(t, "pig.round.you", m["messageCode"])
	assert.Equal(t, "P", m["messageParams"].(map[string]any)["word"])
	assert.Equal(t, "P", m["players"].([]any)[0].(map[string]any)["letterWord"])
}

func TestPigWebPresenterMessages(t *testing.T) {
	p := new(PigWebPresenter)

	t.Run("エラーはそのまま返す", func(t *testing.T) {
		m := decodePig(t, p.Output(newPigForWeb(t), errors.New("boom")))
		assert.Equal(t, "boom", m["message"])
		assert.Empty(t, m["messageCode"])
	})

	t.Run("渡す場面はそう言う", func(t *testing.T) {
		assert.Equal(t, "pig.pass", decodePig(t, p.Output(newPigForWeb(t), nil))["messageCode"])
	})

	// **人間が脱落しても局は続く。**
	t.Run("脱落した人間にはそう伝える", func(t *testing.T) {
		g := newPigForWeb(t)
		g.GetPlayer(0).SetEliminated(true)
		assert.Equal(t, "pig.eliminated", decodePig(t, p.Output(g, nil))["messageCode"])
	})

	t.Run("投了すると相手の勝ち", func(t *testing.T) {
		g := newPigForWeb(t)
		g.GiveUp()
		m := decodePig(t, p.Output(g, nil))
		assert.Equal(t, "pig.result.cpu", m["messageCode"])
	})

	// **最後の 1 人になるのは、最後の相手が 3 文字目をもらった瞬間。**
	//
	// 生存者を 1 人だけ残した盤面を直に作ってはいけません——進行中の局に生存者が
	// 1 人しか居ない状態は codec 自身が拒む、到達しない状態です。
	t.Run("最後まで残ったら勝ち", func(t *testing.T) {
		g := newPigForWeb(t)
		for i := 1; i < g.GetPlayerCnt()-1; i++ {
			g.GetPlayer(i).SetLetters(domain.PigMaxLetters)
			g.GetPlayer(i).SetEliminated(true)
		}
		last := g.GetPlayerCnt() - 1
		g.GetPlayer(last).SetLetters(domain.PigMaxLetters - 1)

		// 人間が気づき、最後の相手が取り残されて 3 文字目を受け取る。
		g.OpenSignalForTest(0)
		require.True(t, g.GetGameEndFlag())
		assert.Equal(t, 0, g.GetWinnerIdx())
		assert.Equal(t, "pig.result.you", decodePig(t, p.Output(g, nil))["messageCode"])
	})
}

func TestPigWebPresenterHintOutput(t *testing.T) {
	p := new(PigWebPresenter)
	g := newPigForWeb(t)

	hint := decodePig(t, p.HintOutput(g))["hint"].(map[string]any)
	assert.NotNil(t, hint["cardIndex"], "渡す場面では札を指す")

	// **合図の場面では札を指さない。**
	g.OpenSignalForTest(1)
	hint = decodePig(t, p.HintOutput(g))["hint"].(map[string]any)
	assert.Nil(t, hint["cardIndex"])
	assert.Equal(t, "pigSignal", hint["reason"])

	// 受動ヒントは Output() でも埋まる (#4483)。
	assert.NotNil(t, decodePig(t, p.Output(g, nil))["hint"])

	g.GiveUp()
	assert.Nil(t, decodePig(t, p.HintOutput(g))["hint"], "終局後は助言しない")
}

// **棋譜は終局まで伏せる。**
func TestPigWebPresenterActionLogOutput(t *testing.T) {
	p := new(PigWebPresenter)
	g := newPigForWeb(t)
	assert.Empty(t, decodePig(t, p.ActionLogOutput(g))["entries"])

	g.GiveUp()
	assert.NotEmpty(t, decodePig(t, p.ActionLogOutput(g))["entries"])
}
