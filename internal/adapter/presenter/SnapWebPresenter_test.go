//go:build test

package presenter

import (
	"encoding/json"
	"errors"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newSnapForWeb(t *testing.T) *domain.Snap {
	t.Helper()
	g := domain.NewDefaultSnap()
	g.SetClock(func() time.Time { return time.UnixMilli(0) })
	g.SetRand(rand.New(rand.NewSource(1)))
	g.Reset()
	return g
}

func decodeSnap(t *testing.T, str string) map[string]any {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.Unmarshal([]byte(str), &m))
	return m
}

func TestSnapWebPresenterOutput(t *testing.T) {
	p := new(SnapWebPresenter)
	m := decodeSnap(t, p.Output(newSnapForWeb(t), nil))

	assert.Equal(t, float64(domain.SnapPhasePlay), m["phase"])
	assert.Equal(t, float64(-1), m["winnerIdx"])
	assert.Equal(t, float64(0), m["centerPileSize"], "まだ誰もめくっていない")
	assert.Nil(t, m["topCard"])
	assert.False(t, m["snapAvailable"].(bool))
	assert.Equal(t, float64(domain.SnapDefaultPlayerCnt), m["playerCnt"])

	players := m["players"].([]any)
	require.Len(t, players, domain.SnapDefaultPlayerCnt)
	human := players[0].(map[string]any)
	assert.True(t, human["isHuman"].(bool))
	assert.Equal(t, float64(26), human["stockSize"], "2 人なら 26 枚ずつ")
	// **手札は出さない。** ストックは裏向きで、枚数だけが公開情報。
	assert.NotContains(t, human, "cards")
}

// **上 2 枚が同ランクのときだけ真。** ページはこれを見るだけでよい。
func TestSnapWebPresenterCarriesWhetherSnapIsOn(t *testing.T) {
	p := new(SnapWebPresenter)
	g := newSnapForWeb(t)

	g.SetCenterPileForTest([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 7, false)})
	m := decodeSnap(t, p.Output(g, nil))
	assert.False(t, m["snapAvailable"].(bool), "**1 枚では成立しない**")
	assert.Equal(t, float64(1), m["centerPileSize"])
	require.NotNil(t, m["topCard"])

	g.SetCenterPileForTest([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 7, false),
		domain.NewCard(domain.CardDesignHeart, 7, false),
	})
	m = decodeSnap(t, p.Output(g, nil))
	assert.True(t, m["snapAvailable"].(bool))
	assert.Equal(t, "snap.available", m["messageCode"])
}

// **保留と直近イベントはワイヤに載る。** 反射ゲームは盤面だけでは読めない。
func TestSnapWebPresenterCarriesThePendingAndTheLastEvent(t *testing.T) {
	p := new(SnapWebPresenter)
	g := newSnapForWeb(t)
	g.SetPendingForTest(domain.SnapPending{
		Kind: domain.SnapPendingSnap, PlayerIdx: 1, DeadlineMs: 1234,
	})
	g.StepForTest(0)

	m := decodeSnap(t, p.Output(g, nil))
	assert.Equal(t, float64(domain.SnapEventStep), m["lastEventKind"])
	assert.Equal(t, float64(0), m["lastEventPlayerIdx"])
	assert.GreaterOrEqual(t, m["pendingDeadlineMs"].(float64), float64(0))
}

func TestSnapWebPresenterMessages(t *testing.T) {
	p := new(SnapWebPresenter)

	t.Run("エラーはそのまま返す", func(t *testing.T) {
		m := decodeSnap(t, p.Output(newSnapForWeb(t), errors.New("boom")))
		assert.Equal(t, "boom", m["message"])
		assert.Empty(t, m["messageCode"])
	})

	t.Run("プレイ中は場札の枚数を出す", func(t *testing.T) {
		m := decodeSnap(t, p.Output(newSnapForWeb(t), nil))
		assert.Equal(t, "snap.play", m["messageCode"])
		assert.Equal(t, "0", m["messageParams"].(map[string]any)["n"])
	})

	t.Run("あなたの勝ち", func(t *testing.T) {
		g := newSnapForWeb(t)
		g.GiveStockForTest(1)
		g.SetCenterPileForTest(nil)
		// 席 1 の手番でストックが尽きている → 脱落 → 席 0 が全札
		g.SetCurrentTurnIdxForTest(1)
		g.StepForTest(1)
		m := decodeSnap(t, p.Output(g, nil))
		assert.Equal(t, "snap.result.you", m["messageCode"])
	})

	t.Run("相手の勝ち", func(t *testing.T) {
		g := newSnapForWeb(t)
		g.GiveUp()
		m := decodeSnap(t, p.Output(g, nil))
		assert.Equal(t, "snap.result.cpu", m["messageCode"])
		assert.Equal(t, "1", m["messageParams"].(map[string]any)["idx"])
	})
}

func TestSnapWebPresenterHintOutput(t *testing.T) {
	p := new(SnapWebPresenter)
	g := newSnapForWeb(t)
	g.SetCenterPileForTest([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 7, false),
		domain.NewCard(domain.CardDesignHeart, 7, false),
	})

	hint := decodeSnap(t, p.HintOutput(g))["hint"].(map[string]any)
	assert.True(t, hint["snap"].(bool))
	assert.Equal(t, "snapDeclare", hint["reason"])

	// 受動ヒントは Output() でも埋まる (#4483)。
	assert.NotNil(t, decodeSnap(t, p.Output(g, nil))["hint"])

	g.GiveUp()
	assert.Nil(t, decodeSnap(t, p.HintOutput(g))["hint"], "終局後は助言しない")
}

// **棋譜は終局まで伏せる。**
func TestSnapWebPresenterActionLogOutput(t *testing.T) {
	p := new(SnapWebPresenter)
	g := newSnapForWeb(t)
	assert.Empty(t, decodeSnap(t, p.ActionLogOutput(g))["entries"])

	g.GiveUp()
	assert.NotEmpty(t, decodeSnap(t, p.ActionLogOutput(g))["entries"])
}
