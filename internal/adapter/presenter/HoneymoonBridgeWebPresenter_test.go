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

func newHoneymoonBridgeForWeb(t *testing.T) *domain.HoneymoonBridge {
	t.Helper()
	h := domain.NewDefaultHoneymoonBridge()
	h.Reset()
	return h
}

func decodeHoneymoonBridge(t *testing.T, str string) map[string]any {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.Unmarshal([]byte(str), &m))
	return m
}

func TestHoneymoonBridgeWebPresenterOutput(t *testing.T) {
	p := new(HoneymoonBridgeWebPresenter)
	m := decodeHoneymoonBridge(t, p.Output(newHoneymoonBridgeForWeb(t), nil))

	assert.Equal(t, float64(domain.HoneymoonBridgePhaseDraw), m["phase"])
	assert.Equal(t, float64(1), m["roundNumber"])
	assert.Equal(t, float64(domain.HoneymoonBridgeStockSize), m["stockSize"], "山札は 26 枚")
	assert.Equal(t, float64(0), m["trumpSuit"], "引き合いは切り札なし")
	assert.Equal(t, float64(-1), m["declarerIdx"])
	assert.Equal(t, float64(0), m["contractLevel"])
	assert.Equal(t, float64(0), m["requiredTricks"], "契約前は 0")
	assert.Equal(t, float64(-1), m["winnerIdx"])
	assert.Equal(t, float64(domain.HoneymoonBridgeDefaultTarget),
		m["config"].(map[string]any)["target"])

	players := m["players"].([]any)
	require.Len(t, players, domain.HoneymoonBridgePlayerCnt)
	human := players[0].(map[string]any)
	assert.True(t, human["isHuman"].(bool))
	assert.Equal(t, float64(domain.HoneymoonBridgeHandSize), human["cardCount"], "13 枚配る")
	assert.Empty(t, players[1].(map[string]any)["cards"], "CPU の手札は伏せる")
}

// **サーバが必ず拒否する値をクライアントに出させない。** 競りのあいだだけ
// 「次に通る最小の宣言」を載せる（[[feedback_page_rederives_a_domain_rule]]）。
func TestHoneymoonBridgeWebPresenterCarriesTheMinimumBid(t *testing.T) {
	p := new(HoneymoonBridgeWebPresenter)

	draw := decodeHoneymoonBridge(t, p.Output(newHoneymoonBridgeForWeb(t), nil))
	assert.Equal(t, float64(0), draw["minBidLevel"], "競り以外では出さない")

	h := newHoneymoonBridgeForWeb(t)
	h.SetPhaseForTest(domain.HoneymoonBridgePhaseBid)
	h.SetCurrentPlayerIdxForTest(0)
	m := decodeHoneymoonBridge(t, p.Output(h, nil))
	assert.Equal(t, float64(1), m["minBidLevel"], "まだ誰も宣言していない")
	assert.Equal(t, float64(domain.CardDesignSpade), m["minBidSuit"])

	require.NoError(t, h.PlayerBid(2, domain.CardDesignHeart))
	h.SetCurrentPlayerIdxForTest(0)
	m = decodeHoneymoonBridge(t, p.Output(h, nil))
	assert.Equal(t, float64(2), m["minBidLevel"], "同レベルで上のスートが最小")
	assert.Equal(t, float64(domain.CardDesignDiamond), m["minBidSuit"])

	// **上限に張り付いたら 0。** 「pass しかない」を出せる。
	h.SetContractForTest(1, domain.HoneymoonBridgeMaxLevel, 0)
	h.SetCurrentPlayerIdxForTest(0)
	m = decodeHoneymoonBridge(t, p.Output(h, nil))
	assert.Equal(t, float64(0), m["minBidLevel"])
}

func TestHoneymoonBridgeWebPresenterCarriesTheContract(t *testing.T) {
	p := new(HoneymoonBridgeWebPresenter)
	h := newHoneymoonBridgeForWeb(t)
	h.SetPhaseForTest(domain.HoneymoonBridgePhasePlay)
	h.SetContractForTest(1, 3, domain.CardDesignHeart)
	h.GiveTricksForTest(1, 4)

	m := decodeHoneymoonBridge(t, p.Output(h, nil))
	assert.Equal(t, float64(1), m["declarerIdx"])
	assert.Equal(t, float64(3), m["contractLevel"])
	assert.Equal(t, float64(domain.CardDesignHeart), m["trumpSuit"])
	assert.Equal(t, float64(domain.HoneymoonBridgeBookTricks+3), m["requiredTricks"])
	assert.Equal(t, float64(4), m["players"].([]any)[1].(map[string]any)["trickCount"])
	assert.Equal(t, "honeymoonbridge.play", m["messageCode"])
	assert.Equal(t, "4", m["messageParams"].(map[string]any)["took"])
}

// **宣言はワイヤに載る。** 相手が何を言ったか読めないと競れない。
func TestHoneymoonBridgeWebPresenterCarriesEachSeatsBid(t *testing.T) {
	p := new(HoneymoonBridgeWebPresenter)
	h := newHoneymoonBridgeForWeb(t)
	h.SetPhaseForTest(domain.HoneymoonBridgePhaseBid)
	h.SetCurrentPlayerIdxForTest(0)
	require.NoError(t, h.PlayerBid(4, domain.CardDesignClover))

	human := decodeHoneymoonBridge(t, p.Output(h, nil))["players"].([]any)[0].(map[string]any)
	assert.Equal(t, float64(4), human["bidLevel"])
	assert.Equal(t, float64(domain.CardDesignClover), human["bidSuit"])
}

func TestHoneymoonBridgeWebPresenterMessages(t *testing.T) {
	p := new(HoneymoonBridgeWebPresenter)

	t.Run("エラーはそのまま返す", func(t *testing.T) {
		m := decodeHoneymoonBridge(t, p.Output(newHoneymoonBridgeForWeb(t), errors.New("boom")))
		assert.Equal(t, "boom", m["message"])
		assert.Empty(t, m["messageCode"])
	})

	t.Run("引き合いは山札の残りを出す", func(t *testing.T) {
		m := decodeHoneymoonBridge(t, p.Output(newHoneymoonBridgeForWeb(t), nil))
		assert.Equal(t, "honeymoonbridge.draw", m["messageCode"])
		assert.Equal(t, "26", m["messageParams"].(map[string]any)["stock"])
	})

	for _, tc := range []struct {
		name    string
		current int
		want    string
	}{
		{"人間の番", 0, "honeymoonbridge.bid.choose"},
		{"相手の番", 1, "honeymoonbridge.bid.wait"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := newHoneymoonBridgeForWeb(t)
			h.SetPhaseForTest(domain.HoneymoonBridgePhaseBid)
			h.SetCurrentPlayerIdxForTest(tc.current)
			assert.Equal(t, tc.want, decodeHoneymoonBridge(t, p.Output(h, nil))["messageCode"])
		})
	}

	for _, tc := range []struct {
		name  string
		level int
		took  int
		want  string
	}{
		{"成立", 2, 8, "honeymoonbridge.roundEnd.made"},
		{"失敗", 3, 5, "honeymoonbridge.roundEnd.down"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := newHoneymoonBridgeForWeb(t)
			h.SetContractForTest(0, tc.level, domain.CardDesignHeart)
			h.SetPhaseForTest(domain.HoneymoonBridgePhasePlay)
			h.GiveTricksForTest(0, tc.took)
			h.GiveTricksForTest(1, domain.HoneymoonBridgeTricksPerPhase-tc.took)
			h.FinishRoundForTest()
			m := decodeHoneymoonBridge(t, p.Output(h, nil))
			assert.Equal(t, tc.want, m["messageCode"])
			assert.Equal(t, strconv.Itoa(domain.HoneymoonBridgeBookTricks+tc.level),
				m["messageParams"].(map[string]any)["need"], "必要トリックは 6+レベル")
		})
	}

	// **両者パスならディールは流れる。** 契約が無いので「成立/失敗」は言えない。
	t.Run("流局", func(t *testing.T) {
		h := newHoneymoonBridgeForWeb(t)
		h.SetPhaseForTest(domain.HoneymoonBridgePhaseRoundEnd)
		assert.Equal(t, "honeymoonbridge.roundEnd.passedOut",
			decodeHoneymoonBridge(t, p.Output(h, nil))["messageCode"])
	})

	for _, tc := range []struct {
		name   string
		winner int
		want   string
	}{
		{"あなたの勝ち", 0, "honeymoonbridge.result.you"},
		{"相手の勝ち", 1, "honeymoonbridge.result.cpu"},
		{"同点", -1, "honeymoonbridge.result.tie"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := newHoneymoonBridgeForWeb(t)
			if tc.winner >= 0 {
				h.GetPlayer(tc.winner).SetScore(1)
			}
			h.FinishGameForTest()
			m := decodeHoneymoonBridge(t, p.Output(h, nil))
			assert.Equal(t, tc.want, m["messageCode"])
			assert.Equal(t, strconv.Itoa(tc.winner), m["messageParams"].(map[string]any)["idx"])
		})
	}
}

func TestHoneymoonBridgeWebPresenterHintOutput(t *testing.T) {
	p := new(HoneymoonBridgeWebPresenter)

	h := newHoneymoonBridgeForWeb(t)
	h.SetCurrentPlayerIdxForTest(0)
	m := decodeHoneymoonBridge(t, p.HintOutput(h))
	hint := m["hint"].(map[string]any)
	assert.Equal(t, "honeymoonbridgeDraw", hint["reason"])
	assert.NotNil(t, hint["cardIndex"])

	// **競りの助言は札ではなく契約を指す。**
	h.SetPhaseForTest(domain.HoneymoonBridgePhaseBid)
	hint = decodeHoneymoonBridge(t, p.HintOutput(h))["hint"].(map[string]any)
	assert.Nil(t, hint["cardIndex"])
	assert.Contains(t, []any{"honeymoonbridgeBid", "honeymoonbridgePass"}, hint["reason"])

	// 受動ヒントは Output() でも埋まる (#4483)。
	assert.NotNil(t, decodeHoneymoonBridge(t, p.Output(h, nil))["hint"])

	h.FinishGameForTest()
	assert.Nil(t, decodeHoneymoonBridge(t, p.HintOutput(h))["hint"], "終局後は助言しない")
}

// **棋譜は終局まで伏せる。** 進行中に出すと相手の手が読める。
func TestHoneymoonBridgeWebPresenterActionLogOutput(t *testing.T) {
	p := new(HoneymoonBridgeWebPresenter)
	h := newHoneymoonBridgeForWeb(t)
	assert.Empty(t, decodeHoneymoonBridge(t, p.ActionLogOutput(h))["entries"])

	h.GiveUp()
	assert.NotEmpty(t, decodeHoneymoonBridge(t, p.ActionLogOutput(h))["entries"])
}
