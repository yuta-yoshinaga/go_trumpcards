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

// newIronCrossForPresenter は本物のドメインを返す。
func newIronCrossForPresenter(t *testing.T) *domain.IronCross {
	t.Helper()
	g := domain.NewDefaultIronCross()
	g.Reset()
	return g
}

// ironCrossAtChoose は縦横を選ぶ場面まで進めた卓を返す。
func ironCrossAtChoose(t *testing.T) *domain.IronCross {
	t.Helper()
	g := newIronCrossForPresenter(t)
	for steps := 0; g.GetPhase() == domain.IronCrossPhaseBetting; steps++ {
		require.Less(t, steps, 200)
		if err := g.PlayerAction(domain.IronCrossActionCheck, 0); err != nil {
			require.NoError(t, g.PlayerAction(domain.IronCrossActionCall, 0))
		}
	}
	return g
}

// ironCrossSettled はショーダウンまで進めた卓を返す。
func ironCrossSettled(t *testing.T) *domain.IronCross {
	t.Helper()
	g := ironCrossAtChoose(t)
	if g.IsChoosing() {
		require.NoError(t, g.ChooseLine(domain.IronCrossLineVertical))
	}
	return g
}

// --- CUI ---

func TestIronCrossCuiPresenter_RendersJapaneseNotRawKeys(t *testing.T) {
	cp := new(IronCrossCuiPresenter)
	out := cp.Output(newIronCrossForPresenter(t), nil)

	assert.Contains(t, out, "フェーズ:")
	assert.Contains(t, out, "ハンド:")
	assert.Contains(t, out, "十字:")
	assert.NotContains(t, out, "ironcross.", "生の i18n キーが出力に混ざっている")
}

// **十字は十字の形で出る。** 1 行に並べたら縦と横が区別できず、このゲームの
// 唯一の判断が成り立たない。
func TestIronCrossCuiPresenter_DrawsTheCrossAsACross(t *testing.T) {
	cp := new(IronCrossCuiPresenter)
	g := ironCrossAtChoose(t)
	out := cp.Output(g, nil)

	cross := g.GetCross()
	require.Len(t, cross, domain.IronCrossCommunityCards)

	lines := splitLines(out)
	// 中央の札は、その左右に左右の札を伴った 1 行に出る。
	center, left, right := cuiCardStr(cross[domain.IronCrossCenter]),
		cuiCardStr(cross[domain.IronCrossLeft]), cuiCardStr(cross[domain.IronCrossRight])
	top, bottom := cuiCardStr(cross[domain.IronCrossTop]), cuiCardStr(cross[domain.IronCrossBottom])

	middle := -1
	for i, l := range lines {
		if countOccurrences(l, center) > 0 && countOccurrences(l, left) > 0 &&
			countOccurrences(l, right) > 0 {
			middle = i
			break
		}
	}
	require.GreaterOrEqual(t, middle, 1, "左・中央・右が同じ行に並んでいない")
	assert.Positive(t, countOccurrences(lines[middle-1], top), "上の札が中央行の 1 行上に無い")
	assert.Positive(t, countOccurrences(lines[middle+1], bottom), "下の札が中央行の 1 行下に無い")
}

func splitLines(s string) []string {
	out, start := []string{}, 0
	for i := range len(s) {
		if s[i] == '\n' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	return append(out, s[start:])
}

// **CPU の手札は伏せたまま。**
func TestIronCrossCuiPresenter_HidesCpuHandsUntilShowdown(t *testing.T) {
	cp := new(IronCrossCuiPresenter)
	g := newIronCrossForPresenter(t)
	out := cp.Output(g, nil)

	assert.Equal(t, len(g.GetPlayers())-1, countOccurrences(out, "(伏せ)"),
		"伏せている席数が合わない")

	after := cp.Output(ironCrossSettled(t), nil)
	assert.NotContains(t, after, "(伏せ)", "ショーダウンでも伏せたままになっている")
}

// **公開枚数を出す。** 降りどころの判断に要る。
func TestIronCrossCuiPresenter_ShowsHowManyCardsRemain(t *testing.T) {
	cp := new(IronCrossCuiPresenter)
	assert.Contains(t, cp.Output(newIronCrossForPresenter(t), nil), "0 / 5 枚公開")
	assert.Contains(t, cp.Output(ironCrossSettled(t), nil), "5 / 5 枚公開")
}

// **選ぶ場面ではそう言う。** ベットの案内を出したままだと、何を入力すれば
// よいのか分からない盤面で止まる。
func TestIronCrossCuiPresenter_AsksForTheLineWhenChoosing(t *testing.T) {
	cp := new(IronCrossCuiPresenter)
	g := ironCrossAtChoose(t)
	require.True(t, g.IsChoosing(), "選ぶ場面まで進んでいない")

	out := cp.Output(g, nil)
	assert.Contains(t, out, "縦(vertical)か横(horizontal)を選んでください")
	assert.NotContains(t, out, "チェックできます", "選ぶ場面でベットの案内が残っている")
}

func TestIronCrossCuiPresenter_ShowsBetGuidanceAndResult(t *testing.T) {
	cp := new(IronCrossCuiPresenter)
	out := cp.Output(newIronCrossForPresenter(t), nil)
	assert.True(t,
		countOccurrences(out, "チェックできます") > 0 || countOccurrences(out, "コールに") > 0,
		"賭けの案内が出ていない")

	settled := cp.Output(ironCrossSettled(t), nil)
	assert.Contains(t, settled, "獲得", "決着の獲得額が出ていない")
	assert.NotContains(t, settled, "ironcross.")
}

func TestIronCrossCuiPresenter_ErrorsAndHint(t *testing.T) {
	cp := new(IronCrossCuiPresenter)
	g := newIronCrossForPresenter(t)
	assert.Contains(t, cp.Output(g, errors.New("賭け金が範囲外です")), "賭け金が範囲外です")
	assert.NotEmpty(t, cp.ActionLogOutput(g))

	hint := cp.HintOutput(g)
	assert.NotEmpty(t, hint)
	assert.NotContains(t, hint, "ironcross.", "助言のキーが訳されていない")

	// **選ぶ場面の助言は列を名指しする。** 「列を選ぶ」だけでは助言にならない。
	choosing := cp.HintOutput(ironCrossAtChoose(t))
	assert.True(t,
		countOccurrences(choosing, "縦を選ぶ") > 0 || countOccurrences(choosing, "横を選ぶ") > 0,
		"助言がどちらの列かを言っていない: %s", choosing)

	assert.Equal(t, "いまは助言できる場面ではありません", cp.HintOutput(ironCrossSettled(t)))
}

// --- Web ---

func TestIronCrossWebPresenter_ArraysAreNeverNull(t *testing.T) {
	cp := new(IronCrossWebPresenter)
	var out map[string]json.RawMessage
	require.NoError(t, json.Unmarshal([]byte(cp.Output(newIronCrossForPresenter(t), nil)), &out))
	for _, key := range []string{"seats", "cross", "verticalIndexes", "horizontalIndexes"} {
		assert.NotEqual(t, "null", string(out[key]), "%s が null で返っている", key)
	}
}

// **十字は添字を保ったまま送る。** 伏せている位置を詰めると、3 枚目が中央なのか
// 左なのか画面には分からず、縦横の選択が成り立たなくなる。
func TestIronCrossWebPresenter_KeepsTheCrossPositions(t *testing.T) {
	cp := new(IronCrossWebPresenter)
	g := newIronCrossForPresenter(t)

	var got struct {
		Cross             []json.RawMessage `json:"cross"`
		RevealedCount     int               `json:"revealedCount"`
		CrossTotal        int               `json:"crossTotal"`
		VerticalIndexes   []int             `json:"verticalIndexes"`
		HorizontalIndexes []int             `json:"horizontalIndexes"`
	}
	require.NoError(t, json.Unmarshal([]byte(cp.Output(g, nil)), &got))

	// 配る前でも 5 つの枠がある。
	assert.Len(t, got.Cross, domain.IronCrossCommunityCards)
	for i, c := range got.Cross {
		assert.Equal(t, "null", string(c), "位置 %d が伏せられていない", i)
	}
	assert.Zero(t, got.RevealedCount)
	assert.Equal(t, domain.IronCrossCommunityCards, got.CrossTotal)
	assert.Equal(t, domain.IronCrossLineIndexes(domain.IronCrossLineVertical), got.VerticalIndexes)
	assert.Equal(t, domain.IronCrossLineIndexes(domain.IronCrossLineHorizontal), got.HorizontalIndexes)
	// 中央は両方の列に入る唯一の位置。
	assert.Contains(t, got.VerticalIndexes, domain.IronCrossCenter)
	assert.Contains(t, got.HorizontalIndexes, domain.IronCrossCenter)

	// 開いた札はその位置に入る。
	after := ironCrossAtChoose(t)
	require.NoError(t, json.Unmarshal([]byte(cp.Output(after, nil)), &got))
	require.Len(t, got.Cross, domain.IronCrossCommunityCards)
	for i, c := range got.Cross {
		assert.NotEqual(t, "null", string(c), "位置 %d が開いていない", i)
	}
}

// **CPU の手札と選んだ列をワイヤに乗せない。**
func TestIronCrossWebPresenter_DoesNotShipCpuHands(t *testing.T) {
	cp := new(IronCrossWebPresenter)
	g := newIronCrossForPresenter(t)

	var got struct {
		Seats []struct {
			IsHuman bool              `json:"isHuman"`
			Cards   []json.RawMessage `json:"cards"`
			Line    int               `json:"line"`
		} `json:"seats"`
	}
	require.NoError(t, json.Unmarshal([]byte(cp.Output(g, nil)), &got))
	require.Len(t, got.Seats, len(g.GetPlayers()))

	humans := 0
	for i, s := range got.Seats {
		if s.IsHuman {
			humans++
			assert.Len(t, s.Cards, domain.IronCrossHoleCards, "人間の手札が届いていない")
			continue
		}
		assert.Empty(t, s.Cards, "席 %d (CPU) の手札がワイヤに乗っている", i)
		assert.Equal(t, int(domain.IronCrossLineNone), s.Line,
			"席 %d (CPU) の選んだ列がワイヤに乗っている", i)
	}
	require.Equal(t, 1, humans)
}

// **ショーダウンでは全員の手札・役・使った列が載る。**
func TestIronCrossWebPresenter_ShowsEveryHandAtShowdown(t *testing.T) {
	cp := new(IronCrossWebPresenter)
	g := ironCrossSettled(t)

	var got struct {
		Seats []struct {
			Cards    []json.RawMessage `json:"cards"`
			BestHand []json.RawMessage `json:"bestHand"`
			Line     int               `json:"line"`
			Folded   bool              `json:"folded"`
		} `json:"seats"`
		RevealedCount int  `json:"revealedCount"`
		Pot           int  `json:"pot"`
		WinnerSeat    int  `json:"winnerSeat"`
		IsChoosing    bool `json:"isChoosing"`
	}
	require.NoError(t, json.Unmarshal([]byte(cp.Output(g, nil)), &got))

	assert.Equal(t, domain.IronCrossCommunityCards, got.RevealedCount)
	assert.Zero(t, got.Pot, "決着後にポットが残っている")
	assert.False(t, got.IsChoosing)
	for i, s := range got.Seats {
		assert.Len(t, s.Cards, domain.IronCrossHoleCards, "席 %d の手札が開いていない", i)
		if !s.Folded {
			assert.Len(t, s.BestHand, domain.IronCrossHandSize, "席 %d の最良 5 枚が載っていない", i)
			assert.NotEqual(t, int(domain.IronCrossLineNone), s.Line,
				"席 %d が使った列が載っていない", i)
		}
	}
	assert.Equal(t, g.WinnerSeat(), got.WinnerSeat)
}

// **選ぶ場面であることはサーバが載せる。** ページにフェーズ番号から
// 割り出させない。
func TestIronCrossWebPresenter_FlagsTheChoosePhase(t *testing.T) {
	cp := new(IronCrossWebPresenter)

	var got struct {
		IsChoosing bool `json:"isChoosing"`
		Phase      int  `json:"phase"`
	}
	require.NoError(t, json.Unmarshal([]byte(cp.Output(ironCrossAtChoose(t), nil)), &got))
	assert.True(t, got.IsChoosing, "選ぶ場面が isChoosing に出ていない")
	assert.Equal(t, int(domain.IronCrossPhaseChoose), got.Phase)

	require.NoError(t, json.Unmarshal([]byte(cp.Output(newIronCrossForPresenter(t), nil)), &got))
	assert.False(t, got.IsChoosing, "配った直後に選ぶ場面になっている")
}

// **賭けの状態はサーバが載せる。** ページに計算し直させない。
func TestIronCrossWebPresenter_BettingStateIsOnTheWire(t *testing.T) {
	cp := new(IronCrossWebPresenter)
	g := newIronCrossForPresenter(t)

	var got struct {
		Pot         int  `json:"pot"`
		CurrentBet  int  `json:"currentBet"`
		ToCall      int  `json:"toCall"`
		RaiseCount  int  `json:"raiseCount"`
		CanRaise    bool `json:"canRaise"`
		TurnSeat    int  `json:"turnSeat"`
		HumanSeat   int  `json:"humanSeat"`
		IsHumanTurn bool `json:"isHumanTurn"`
	}
	require.NoError(t, json.Unmarshal([]byte(cp.Output(g, nil)), &got))
	assert.Equal(t, g.GetPot(), got.Pot)
	assert.Equal(t, g.GetToCall(), got.ToCall)
	assert.Equal(t, g.CanRaise(), got.CanRaise)
	assert.Equal(t, g.GetTurnSeat(), got.TurnSeat)
	assert.Equal(t, g.HumanSeat(), got.HumanSeat)
	assert.Equal(t, g.IsHumanTurn(), got.IsHumanTurn)
	assert.Positive(t, got.Pot, "アンティがポットに入っていない")
}

func TestIronCrossWebPresenter_ErrorAndHint(t *testing.T) {
	cp := new(IronCrossWebPresenter)
	g := newIronCrossForPresenter(t)

	var withErr struct {
		Message string `json:"message"`
	}
	require.NoError(t, json.Unmarshal([]byte(cp.Output(g, errors.New("boom"))), &withErr))
	assert.Equal(t, "boom", withErr.Message)
	assert.NotEmpty(t, cp.ActionLogOutput(g))

	var hint struct {
		Action string `json:"action"`
		Line   int    `json:"line"`
		Reason string `json:"reason"`
	}
	require.NoError(t, json.Unmarshal([]byte(cp.HintOutput(g)), &hint))
	assert.NotEmpty(t, hint.Action)
	assert.NotEmpty(t, hint.Reason)

	// **選ぶ場面ではどちらの列かまで送る。** action だけでは画面が押す先を
	// 決められない。
	require.NoError(t, json.Unmarshal([]byte(cp.HintOutput(ironCrossAtChoose(t))), &hint))
	assert.Equal(t, "line", hint.Action)
	assert.NotEqual(t, int(domain.IronCrossLineNone), hint.Line, "助言が列を名指ししていない")

	assert.JSONEq(t, `{"hint":null}`, cp.HintOutput(ironCrossSettled(t)))
}
