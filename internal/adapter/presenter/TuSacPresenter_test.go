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
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// newTuSacForPresenter は本物のドメインを返す。
func newTuSacForPresenter(t *testing.T) *domain.TuSac {
	t.Helper()
	g := domain.NewDefaultTuSac()
	g.Reset()
	return g
}

// tuSacAfterDraw は人間が 1 枚引いた局面を返す。
func tuSacAfterDraw(t *testing.T) *domain.TuSac {
	t.Helper()
	g := newTuSacForPresenter(t)
	require.True(t, g.IsHumanTurn())
	require.NoError(t, g.Draw(false))
	return g
}

// --- CUI ---

func TestTuSacCuiPresenter_RendersJapaneseNotRawKeys(t *testing.T) {
	cp := new(TuSacCuiPresenter)
	out := cp.Output(newTuSacForPresenter(t), nil)

	assert.Contains(t, out, "ラウンド:")
	assert.Contains(t, out, "山:")
	assert.Contains(t, out, "あなたの手札:")
	assert.NotContains(t, out, "tusac.", "生の i18n キーが出力に混ざっている")
}

// **手札には番号を振る。** 同じ色・同じ駒が 4 枚あるので、札の名前では
// 指定できない。
func TestTuSacCuiPresenter_NumbersTheHand(t *testing.T) {
	cp := new(TuSacCuiPresenter)
	g := newTuSacForPresenter(t)
	out := cp.Output(g, nil)

	// 1 始まりで、手札の枚数ぶん番号が出る。
	assert.Contains(t, out, "1:")
	n := len(g.GetPlayers()[g.HumanSeat()].GetCards())
	assert.Contains(t, out, "20:", "手札 %d 枚に番号が振られていない", n)
	// **"0:" だけを見ると "20:" に当たる。** 番号の直前が空白か行頭のものを見る。
	handLine := ""
	for _, l := range splitLines(out) {
		if countOccurrences(l, "あなたの手札:") > 0 {
			handLine = l
		}
	}
	require.NotEmpty(t, handLine, "手札の行が無い")
	assert.Contains(t, handLine, " 1:", "1 始まりで振られていない")
	assert.NotContains(t, handLine, " 0:", "0 始まりで振っている")
}

// **相手の手札は出さない。** 見えるのは枚数と、場に出た組み合わせだけ。
func TestTuSacCuiPresenter_ShowsOnlyTheCountForOtherSeats(t *testing.T) {
	cp := new(TuSacCuiPresenter)
	g := newTuSacForPresenter(t)
	out := cp.Output(g, nil)

	for i, p := range g.GetPlayers() {
		if p.GetIsHuman() {
			continue
		}
		// 枚数は出ている。
		assert.Contains(t, out, p.GetName(), "席 %d が出ていない", i)
	}
	// 自分の手札の行は 1 つだけ。
	assert.Equal(t, 1, countOccurrences(out, "あなたの手札:"))
}

// **色と駒で表す。** 数字の大小という概念が無いデッキ。
func TestTuSacCuiPresenter_RendersColourAndPiece(t *testing.T) {
	cp := new(TuSacCuiPresenter)
	out := cp.Output(newTuSacForPresenter(t), nil)

	colours := 0
	for _, c := range []string{"黄", "赤", "緑", "白"} {
		colours += countOccurrences(out, c)
	}
	assert.Positive(t, colours, "色が出ていない")

	pieces := 0
	for _, p := range []string{"將", "士", "象", "車", "馬", "砲", "卒"} {
		pieces += countOccurrences(out, p)
	}
	assert.Positive(t, pieces, "駒が出ていない")
}

// **場面ごとに求める操作が違う。**
func TestTuSacCuiPresenter_PromptsForTheRightAction(t *testing.T) {
	cp := new(TuSacCuiPresenter)

	draw := cp.Output(newTuSacForPresenter(t), nil)
	assert.Contains(t, draw, "山から引く")
	assert.NotContains(t, draw, "1枚捨てる", "引く場面で捨てる案内が出ている")

	discard := cp.Output(tuSacAfterDraw(t), nil)
	assert.Contains(t, discard, "1枚捨てる")
	assert.NotContains(t, discard, "山から引く", "捨てる場面で引く案内が出ている")
}

func TestTuSacCuiPresenter_ErrorsAndHint(t *testing.T) {
	cp := new(TuSacCuiPresenter)
	g := tuSacAfterDraw(t)
	assert.Contains(t, cp.Output(g, errors.New("組み合わせになっていません")), "組み合わせになっていません")
	assert.NotEmpty(t, cp.ActionLogOutput(g))

	hint := cp.HintOutput(g)
	assert.NotEmpty(t, hint)
	assert.NotContains(t, hint, "tusac.", "助言のキーが訳されていない")
	// **薦める札も 1 始まり。** 画面の番号と揃っていないと押し間違える。
	assert.NotContains(t, hint, "（0）", "0 始まりで薦めている")
}

// --- Web ---

func TestTuSacWebPresenter_ArraysAreNeverNull(t *testing.T) {
	cp := new(TuSacWebPresenter)
	var out map[string]json.RawMessage
	require.NoError(t, json.Unmarshal([]byte(cp.Output(newTuSacForPresenter(t), nil)), &out))
	for _, key := range []string{"seats", "meldPointsByKind"} {
		assert.NotEqual(t, "null", string(out[key]), "%s が null で返っている", key)
	}
}

// **相手の手札はワイヤに乗せない。** 枚数だけを送る。
func TestTuSacWebPresenter_DoesNotShipOtherHands(t *testing.T) {
	cp := new(TuSacWebPresenter)
	g := newTuSacForPresenter(t)

	var got struct {
		Seats []struct {
			IsHuman   bool              `json:"isHuman"`
			Cards     []json.RawMessage `json:"cards"`
			HandCount int               `json:"handCount"`
			Melds     []json.RawMessage `json:"melds"`
		} `json:"seats"`
	}
	require.NoError(t, json.Unmarshal([]byte(cp.Output(g, nil)), &got))
	require.Len(t, got.Seats, len(g.GetPlayers()))

	humans := 0
	for i, s := range got.Seats {
		if s.IsHuman {
			humans++
			assert.Len(t, s.Cards, domain.TuSacHandSize, "自分の手札が届いていない")
			continue
		}
		assert.Empty(t, s.Cards, "席 %d の手札がワイヤに乗っている", i)
		// 枚数だけは分かる。
		assert.Positive(t, s.HandCount, "席 %d の枚数が届いていない", i)
	}
	require.Equal(t, 1, humans)
}

// **場に出た組み合わせは全員ぶん見える。** そこから読むのがこのゲームの筋。
func TestTuSacWebPresenter_ShipsEveryMeld(t *testing.T) {
	cp := new(TuSacWebPresenter)
	g := tuSacAfterDraw(t)
	h := g.GetHint()
	if h == nil || h.Action != "meld" {
		t.Skip("この配りでは出せる組み合わせが無い")
	}
	require.NoError(t, g.Meld(h.Indexes))

	var got struct {
		Seats []struct {
			IsHuman bool `json:"isHuman"`
			Melds   []struct {
				Kind   int               `json:"kind"`
				Points int               `json:"points"`
				Cards  []json.RawMessage `json:"cards"`
			} `json:"melds"`
			MeldPoints int `json:"meldPoints"`
		} `json:"seats"`
	}
	require.NoError(t, json.Unmarshal([]byte(cp.Output(g, nil)), &got))

	seat := got.Seats[g.HumanSeat()]
	require.NotEmpty(t, seat.Melds, "出した組み合わせが届いていない")
	assert.Positive(t, seat.Melds[0].Kind)
	assert.Positive(t, seat.Melds[0].Points, "組み合わせの得点が載っていない")
	assert.NotEmpty(t, seat.Melds[0].Cards, "組み合わせの札が載っていない")
	assert.Positive(t, seat.MeldPoints)
}

// **札の総数と得点表はサーバが載せる。**
func TestTuSacWebPresenter_ShipsTheRules(t *testing.T) {
	cp := new(TuSacWebPresenter)
	g := newTuSacForPresenter(t)

	var got struct {
		HandSize         int   `json:"handSize"`
		DeckSize         int   `json:"deckSize"`
		MeldPointsByKind []int `json:"meldPointsByKind"`
		StockCount       int   `json:"stockCount"`
		DiscardCount     int   `json:"discardCount"`
		RoundNumber      int   `json:"roundNumber"`
		Rounds           int   `json:"rounds"`
		WentOutSeat      int   `json:"wentOutSeat"`
		IsHumanTurn      bool  `json:"isHumanTurn"`
	}
	require.NoError(t, json.Unmarshal([]byte(cp.Output(g, nil)), &got))
	assert.Equal(t, domain.TuSacHandSize, got.HandSize)
	assert.Equal(t, 112, got.DeckSize, "四色牌は 112 枚")
	assert.Len(t, got.MeldPointsByKind, int(domain.TuSacMeldKindMax)+1)
	assert.Greater(t, got.MeldPointsByKind[domain.TuSacMeldSoldierSet],
		got.MeldPointsByKind[domain.TuSacMeldSameColorSet])
	assert.Equal(t, g.GetStockCount(), got.StockCount)
	assert.Positive(t, got.DiscardCount, "捨て札が空で始まっている")
	assert.Equal(t, 1, got.RoundNumber)
	assert.Equal(t, g.GetConfig().Rounds, got.Rounds)
	assert.Equal(t, -1, got.WentOutSeat)
	assert.True(t, got.IsHumanTurn)
}

func TestTuSacWebPresenter_ErrorAndHint(t *testing.T) {
	cp := new(TuSacWebPresenter)
	g := tuSacAfterDraw(t)

	var withErr struct {
		Message string `json:"message"`
	}
	require.NoError(t, json.Unmarshal([]byte(cp.Output(g, errors.New("boom"))), &withErr))
	assert.Equal(t, "boom", withErr.Message)
	assert.NotEmpty(t, cp.ActionLogOutput(g))

	var hint struct {
		Action  string `json:"action"`
		Indexes []int  `json:"indexes"`
		Reason  string `json:"reason"`
	}
	require.NoError(t, json.Unmarshal([]byte(cp.HintOutput(g)), &hint))
	assert.Contains(t, []string{"meld", "discard"}, hint.Action)
	assert.NotEmpty(t, hint.Reason)
	assert.NotNil(t, hint.Indexes, "薦める札が null で返っている")
}

// **四色牌の札は手続き描画で送る。** 共有のカード絵はトランプのスートを
// 描くので、四色牌では何も意味しない ── 色と駒の漢字を札そのものに載せる
// (ADR-0033)。
func TestTuSacWebPresenter_ShipsProceduralCardFaces(t *testing.T) {
	cp := new(TuSacWebPresenter)
	g := newTuSacForPresenter(t)

	var got struct {
		Seats []struct {
			IsHuman bool `json:"isHuman"`
			Cards   []struct {
				Glyph string `json:"glyph"`
				Label string `json:"label"`
				Color string `json:"color"`
				Deck  string `json:"deck"`
			} `json:"cards"`
		} `json:"seats"`
		DiscardTop struct {
			Glyph string `json:"glyph"`
			Deck  string `json:"deck"`
		} `json:"discardTop"`
	}
	require.NoError(t, json.Unmarshal([]byte(cp.Output(g, nil)), &got))

	hand := got.Seats[g.HumanSeat()].Cards
	require.NotEmpty(t, hand)
	glyphs := map[string]bool{}
	colors := map[string]bool{}
	for i, c := range hand {
		assert.Equal(t, "tusac", c.Deck, "%d 枚目が手続き描画に乗っていない", i)
		assert.NotEmpty(t, c.Glyph, "%d 枚目に駒の字が無い", i)
		assert.NotEqual(t, "?", c.Glyph, "%d 枚目の駒が不明", i)
		assert.NotEmpty(t, c.Color, "%d 枚目に色が無い", i)
		glyphs[c.Glyph] = true
		colors[c.Color] = true
	}
	// 20 枚あれば駒も色も複数種類出る。
	assert.Greater(t, len(glyphs), 1, "駒の字が 1 種類しか出ていない")
	assert.Greater(t, len(colors), 1, "色が 1 種類しか出ていない")

	assert.Equal(t, "tusac", got.DiscardTop.Deck, "捨て札が手続き描画に乗っていない")
	assert.NotEmpty(t, got.DiscardTop.Glyph)
}

// **5 枚の卒を揃える価値は、狙う前に知りたい** (#5784)。点数は写した表では
// なくドメインから作る。
func TestTuSacCuiPresenter_ShowsTheMeldPointsTable(t *testing.T) {
	cp := new(TuSacCuiPresenter)
	out := cp.Output(newTuSacForPresenter(t), nil)

	assert.Contains(t, out, i18n.Tf("tusac.meldPointsLine", "table", tuSacMeldPointsTableStr()))
	for k := domain.TuSacMeldNone + 1; k <= domain.TuSacMeldKindMax; k++ {
		assert.Contains(t, out, i18n.T("tusac.meld."+domain.TuSacMeldKindName(k))+" "+
			i18n.Tf("tusac.meld.points", "points", strconv.Itoa(domain.TuSacMeldPoints(k))),
			"種別 %d の点が出ていない", k)
	}
	assert.NotContains(t, out, "tusac.")
}
