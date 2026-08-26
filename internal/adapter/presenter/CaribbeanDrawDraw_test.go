package presenter

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// caribbeanDrawDrawState is the slice of game state the draw-phase output reads.
type caribbeanDrawDrawState struct {
	phase     int
	ante      int
	drawCost  int
	gameEnd   bool
	rank      int
	hand      []*domain.Card
	dealer    []*domain.Card
	result    domain.GameResult
	playBet   int
	qualified bool
}

// newCaribbeanDrawDrawMock builds a fully-stubbed game mock.
//
// The sibling helpers in this package predate GetDrawCost, so this one registers
// every getter the two presenters touch, including the draw-specific ones.
func newCaribbeanDrawDrawMock(s caribbeanDrawDrawState) *interfaces.MockCaribbeanDrawGame {
	m := new(interfaces.MockCaribbeanDrawGame)
	m.On("GetPhase").Return(s.phase).Maybe()
	m.On("GetAnteBet").Return(s.ante).Maybe()
	m.On("GetDrawCost").Return(s.drawCost).Maybe()
	m.On("GetGameEndFlag").Return(s.gameEnd).Maybe()
	m.On("GetPlayerHandRank").Return(s.rank).Maybe()
	m.On("GetPlayerHand").Return(s.hand).Maybe()
	m.On("GetDealerHand").Return(s.dealer).Maybe()
	m.On("GetResult").Return(s.result).Maybe()
	m.On("GetPlayBet").Return(s.playBet).Maybe()
	m.On("GetDealerQualified").Return(s.qualified).Maybe()
	m.On("GetChips").Return(1000).Maybe()
	m.On("GetJackpotBet").Return(0).Maybe()
	m.On("GetAntePayout").Return(0).Maybe()
	m.On("GetPlayPayout").Return(0).Maybe()
	m.On("GetJackpotPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetDealerHandRank").Return(domain.PokerHandHighCard).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
	return m
}

// requireTranslated fails when a key resolves to itself.
//
// **`i18n.T` は未翻訳ならキーをそのまま返す。** そのため
// `Contains(out, i18n.T(key))` は、翻訳が丸ごと抜けていても両辺が同じ文字列に
// 退化して必ず通る。期待値が実際の文言になっていることを先に確かめる。
func requireTranslated(t *testing.T, key, rendered string) string {
	t.Helper()
	require.NotEqual(t, key, rendered, "%s resolved to its own key -- the assertion would be vacuous", key)
	require.NotContains(t, rendered, "{{", "%s still holds an uninterpolated placeholder", key)
	return rendered
}

// caribbeanDrawStem renders key with every placeholder filled, one of them set
// to a sentinel, and returns the fixed text before it. "line is absent" can then
// be asserted without pinning the interpolated value.
func caribbeanDrawStem(t *testing.T, key string, params ...string) string {
	t.Helper()
	rendered := requireTranslated(t, key, i18n.Tf(key, params...))
	stem, _, found := strings.Cut(rendered, "\x00")
	require.True(t, found, "%s did not interpolate the sentinel", key)
	require.NotEmpty(t, stem, "%s has no fixed text before the sentinel", key)
	return stem
}

func caribbeanDrawFiveCards(values ...int) []*domain.Card {
	designs := []int{
		domain.CardDesignSpade, domain.CardDesignHeart, domain.CardDesignClover,
		domain.CardDesignDiamond, domain.CardDesignSpade,
	}
	cards := make([]*domain.Card, 0, len(values))
	for i, v := range values {
		cards = append(cards, domain.NewCard(designs[i%len(designs)], v, false))
	}
	return cards
}

// --- CUI presenter: draw-phase help --------------------------------------

// TestCaribbeanDrawCuiPresenter_DrawHelpLine covers the help line the clone
// source has no equivalent of.
//
// **交換できるのはこの一瞬だけで、しかも有料。** 手数料を言わずに勧めると、
// 引いた後に残高が減っていることに気付くことになる。
func TestCaribbeanDrawCuiPresenter_DrawHelpLine(t *testing.T) {
	p := new(CaribbeanDrawCuiPresenter)
	const ante = 100
	expected := requireTranslated(t, "caribbeandraw.drawHelp",
		i18n.Tf("caribbeandraw.drawHelp", "max", "2", "cost", "100"))
	stem := caribbeanDrawStem(t, "caribbeandraw.drawHelp", "max", "2", "cost", "\x00")

	t.Run("appears in the draw phase and names the fee", func(t *testing.T) {
		out := p.Output(newCaribbeanDrawDrawMock(caribbeanDrawDrawState{
			phase: domain.CaribbeanDrawPhaseDraw, ante: ante,
			hand: caribbeanDrawFiveCards(3, 5, 8, 11, 13),
			rank: domain.PokerHandHighCard,
		}), nil)

		assert.Contains(t, out, expected)
		// 手数料はアンテと同額。金額そのものが出ていること。
		assert.Contains(t, out, "100")
		// 交換できる枚数も。
		assert.Contains(t, out, "2")
	})

	// **手数料はアンテ次第。** 定数を焼き込むと、アンテを変えたときに嘘を出す。
	t.Run("quotes the fee for the ante actually posted", func(t *testing.T) {
		out := p.Output(newCaribbeanDrawDrawMock(caribbeanDrawDrawState{
			phase: domain.CaribbeanDrawPhaseDraw, ante: 30,
			hand: caribbeanDrawFiveCards(3, 5, 8, 11, 13),
		}), nil)

		assert.Contains(t, out,
			i18n.Tf("caribbeandraw.drawHelp", "max", "2", "cost", "30"))
		assert.NotContains(t, out, expected, "the 100-chip wording must not leak in")
	})

	for _, phase := range []struct {
		name string
		val  int
	}{
		{"bet", domain.CaribbeanDrawPhaseBet},
		{"action", domain.CaribbeanDrawPhaseAction},
		{"end", domain.CaribbeanDrawPhaseEnd},
	} {
		t.Run("absent in the "+phase.name+" phase", func(t *testing.T) {
			out := p.Output(newCaribbeanDrawDrawMock(caribbeanDrawDrawState{
				phase: phase.val, ante: ante, gameEnd: phase.val == domain.CaribbeanDrawPhaseEnd,
				hand: caribbeanDrawFiveCards(3, 5, 8, 11, 13),
			}), nil)

			assert.NotContains(t, out, stem)
		})
	}
}

// --- CUI presenter: draw fee on the result board --------------------------

// TestCaribbeanDrawCuiPresenter_DrawCostLine covers the fee line on the result
// board. The fee is not a payout, so it never shows up in the payout total --
// without this line the balance arithmetic simply does not add up.
func TestCaribbeanDrawCuiPresenter_DrawCostLine(t *testing.T) {
	p := new(CaribbeanDrawCuiPresenter)
	stem := caribbeanDrawStem(t, "caribbeandraw.drawCostLine", "cost", "\x00")

	endState := func(cost int) caribbeanDrawDrawState {
		return caribbeanDrawDrawState{
			phase: domain.CaribbeanDrawPhaseEnd, ante: 100, drawCost: cost,
			gameEnd: true, playBet: 200, result: domain.GameResultWin,
			qualified: true,
			hand:      caribbeanDrawFiveCards(9, 9, 4, 7, 2),
			dealer:    caribbeanDrawFiveCards(8, 8, 3, 5, 12),
			rank:      domain.PokerHandOnePair,
		}
	}

	t.Run("shown when the player paid for a draw", func(t *testing.T) {
		out := p.Output(newCaribbeanDrawDrawMock(endState(100)), nil)
		assert.Contains(t, out,
			requireTranslated(t, "caribbeandraw.drawCostLine",
				i18n.Tf("caribbeandraw.drawCostLine", "cost", "100")))
	})

	t.Run("quotes the fee actually charged", func(t *testing.T) {
		out := p.Output(newCaribbeanDrawDrawMock(endState(30)), nil)
		assert.Contains(t, out, i18n.Tf("caribbeandraw.drawCostLine", "cost", "30"))
		assert.NotContains(t, out, i18n.Tf("caribbeandraw.drawCostLine", "cost", "100"))
	})

	// **引かなかったラウンドには 0 の行も出さない。** 常に出すと、
	// 手数料を払ったラウンドとの見分けが付かなくなる。
	t.Run("absent when the player stood pat", func(t *testing.T) {
		out := p.Output(newCaribbeanDrawDrawMock(endState(0)), nil)
		assert.NotContains(t, out, stem)
	})

	// 途中の画面は結果ボードではない。
	t.Run("absent before the round ends", func(t *testing.T) {
		out := p.Output(newCaribbeanDrawDrawMock(caribbeanDrawDrawState{
			phase: domain.CaribbeanDrawPhaseAction, ante: 100, drawCost: 100,
			hand:   caribbeanDrawFiveCards(9, 9, 4, 7, 2),
			dealer: caribbeanDrawFiveCards(8, 3, 5, 12, 2),
			rank:   domain.PokerHandOnePair,
		}), nil)

		assert.NotContains(t, out, stem)
	})
}

// --- CUI presenter: phase label -------------------------------------------

// TestCaribbeanDrawCuiPresenter_PhaseStr_Draw pins the new phase's label.
//
// Phases renumbered when the draw phase was inserted (Bet=1, Draw=2, Action=3,
// End=4); a phaseStr that never learned about Draw renders "unknown" on a
// perfectly normal turn.
func TestCaribbeanDrawCuiPresenter_PhaseStr_Draw(t *testing.T) {
	p := new(CaribbeanDrawCuiPresenter)

	drawLabel := requireTranslated(t, "caribbeandraw.phaseDraw", i18n.T("caribbeandraw.phaseDraw"))
	unknownLabel := requireTranslated(t, "caribbeandraw.phaseUnknown", i18n.T("caribbeandraw.phaseUnknown"))
	require.NotEqual(t, drawLabel, unknownLabel)

	out := p.Output(newCaribbeanDrawDrawMock(caribbeanDrawDrawState{
		phase: domain.CaribbeanDrawPhaseDraw, ante: 100,
		hand: caribbeanDrawFiveCards(3, 5, 8, 11, 13),
	}), nil)

	assert.Contains(t, out, i18n.Tf("caribbeandraw.phaseLine", "phase", drawLabel))
	assert.NotContains(t, out, i18n.Tf("caribbeandraw.phaseLine", "phase", unknownLabel))

	// 負のコントロール: 本当に未知のフェーズなら unknown が出る。
	unknownOut := p.Output(newCaribbeanDrawDrawMock(caribbeanDrawDrawState{phase: 99}), nil)
	assert.Contains(t, unknownOut, i18n.Tf("caribbeandraw.phaseLine", "phase", unknownLabel))

	// 他のフェーズのラベルも Draw と混ざっていないこと。
	for _, phase := range []int{
		domain.CaribbeanDrawPhaseBet, domain.CaribbeanDrawPhaseAction, domain.CaribbeanDrawPhaseEnd,
	} {
		other := p.Output(newCaribbeanDrawDrawMock(caribbeanDrawDrawState{
			phase: phase, ante: 100, gameEnd: phase == domain.CaribbeanDrawPhaseEnd,
		}), nil)
		assert.NotContains(t, other, i18n.Tf("caribbeandraw.phaseLine", "phase", drawLabel))
	}
}

// --- CUI presenter: draw-phase hint ---------------------------------------

// TestCaribbeanDrawCuiPresenter_HintOutput_DrawPhase covers the advice branch
// that only exists because of the draw phase.
func TestCaribbeanDrawCuiPresenter_HintOutput_DrawPhase(t *testing.T) {
	p := new(CaribbeanDrawCuiPresenter)

	standPat := requireTranslated(t, "caribbeandraw.hintStandPat", i18n.T("caribbeandraw.hintStandPat"))
	drawWeak := requireTranslated(t, "caribbeandraw.hintDrawWeak", i18n.T("caribbeandraw.hintDrawWeak"))
	require.NotEqual(t, standPat, drawWeak, "the two draw hints must be distinguishable")

	pair := caribbeanDrawFiveCards(9, 9, 4, 7, 2)
	junk := caribbeanDrawFiveCards(9, 7, 5, 4, 2)

	t.Run("stands pat with a made hand", func(t *testing.T) {
		out := p.HintOutput(newCaribbeanDrawDrawMock(caribbeanDrawDrawState{
			phase: domain.CaribbeanDrawPhaseDraw, rank: domain.PokerHandOnePair, hand: pair,
		}))
		assert.Contains(t, out, standPat)
		assert.NotContains(t, out, drawWeak)
	})

	t.Run("stands pat with a big made hand", func(t *testing.T) {
		out := p.HintOutput(newCaribbeanDrawDrawMock(caribbeanDrawDrawState{
			phase: domain.CaribbeanDrawPhaseDraw, rank: domain.PokerHandFlush, hand: pair,
		}))
		assert.Contains(t, out, standPat)
	})

	t.Run("suggests drawing without a made hand", func(t *testing.T) {
		out := p.HintOutput(newCaribbeanDrawDrawMock(caribbeanDrawDrawState{
			phase: domain.CaribbeanDrawPhaseDraw, rank: domain.PokerHandHighCard, hand: junk,
		}))
		assert.Contains(t, out, drawWeak)
		assert.NotContains(t, out, standPat)
	})

	// **ドローの助言はドローフェーズだけ。** アクションフェーズに漏れると、
	// もう引けない場面で「交換を検討」と出る。
	t.Run("gives play/fold advice in the action phase, not draw advice", func(t *testing.T) {
		playOut := p.HintOutput(newCaribbeanDrawDrawMock(caribbeanDrawDrawState{
			phase: domain.CaribbeanDrawPhaseAction, rank: domain.PokerHandOnePair, hand: pair,
		}))
		assert.Contains(t, playOut, i18n.T("caribbeandraw.hintPlay"))
		assert.Contains(t, playOut, i18n.T("caribbeandraw.hintPairOrBetter"))
		assert.NotContains(t, playOut, standPat)
		assert.NotContains(t, playOut, drawWeak)

		foldOut := p.HintOutput(newCaribbeanDrawDrawMock(caribbeanDrawDrawState{
			phase: domain.CaribbeanDrawPhaseAction, rank: domain.PokerHandHighCard, hand: junk,
		}))
		assert.Contains(t, foldOut, i18n.T("caribbeandraw.hintFold"))
		assert.NotContains(t, foldOut, drawWeak)
	})

	// 札が配られる前は、フェーズ番号だけでは助言できない。
	t.Run("gives no hint before any card is dealt", func(t *testing.T) {
		out := p.HintOutput(newCaribbeanDrawDrawMock(caribbeanDrawDrawState{
			phase: domain.CaribbeanDrawPhaseDraw, rank: domain.PokerHandHighCard,
		}))
		assert.Contains(t, out, i18n.T("caribbeandraw.hintNone"))
		assert.NotContains(t, out, drawWeak)
	})
}

// --- Web presenter: drawCost ----------------------------------------------

// TestCaribbeanDrawWebPresenter_DrawCost pins the field the web GUI needs to
// show the exchange fee. A missing key unmarshals as 0, so the test looks at the
// raw JSON rather than the decoded struct.
func TestCaribbeanDrawWebPresenter_DrawCost(t *testing.T) {
	p := new(CaribbeanDrawWebPresenter)

	rawField := func(t *testing.T, jsonStr, field string) json.RawMessage {
		t.Helper()
		var raw map[string]json.RawMessage
		require.NoError(t, json.Unmarshal([]byte(jsonStr), &raw))
		v, ok := raw[field]
		require.True(t, ok, "%q is missing from the response: %s", field, jsonStr)
		return v
	}

	t.Run("carries the fee that was charged", func(t *testing.T) {
		out := p.Output(newCaribbeanDrawDrawMock(caribbeanDrawDrawState{
			phase: domain.CaribbeanDrawPhaseEnd, ante: 100, drawCost: 100,
			gameEnd: true, playBet: 200, result: domain.GameResultWin, qualified: true,
			hand:   caribbeanDrawFiveCards(9, 9, 4, 7, 2),
			dealer: caribbeanDrawFiveCards(8, 8, 3, 5, 12),
			rank:   domain.PokerHandOnePair,
		}), nil)

		assert.JSONEq(t, `100`, string(rawField(t, out, "drawCost")))
	})

	// **引かなかったラウンドでもキーは出す。** 省略されるとフロントが
	// undefined を読み、前のラウンドの手数料が残って見える。
	t.Run("present as 0 when the player stood pat", func(t *testing.T) {
		out := p.Output(newCaribbeanDrawDrawMock(caribbeanDrawDrawState{
			phase: domain.CaribbeanDrawPhaseEnd, ante: 100, drawCost: 0,
			gameEnd: true, playBet: 200, result: domain.GameResultWin, qualified: true,
			hand:   caribbeanDrawFiveCards(9, 9, 4, 7, 2),
			dealer: caribbeanDrawFiveCards(8, 8, 3, 5, 12),
			rank:   domain.PokerHandOnePair,
		}), nil)

		assert.JSONEq(t, `0`, string(rawField(t, out, "drawCost")))
	})

	t.Run("present during the draw phase itself", func(t *testing.T) {
		out := p.Output(newCaribbeanDrawDrawMock(caribbeanDrawDrawState{
			phase: domain.CaribbeanDrawPhaseDraw, ante: 50, drawCost: 50,
			hand:   caribbeanDrawFiveCards(9, 9, 4, 7, 2),
			dealer: caribbeanDrawFiveCards(8, 8, 3, 5, 12),
			rank:   domain.PokerHandOnePair,
		}), nil)

		assert.JSONEq(t, `50`, string(rawField(t, out, "drawCost")))
		// フェーズ番号も 2 (Draw) のまま届くこと。
		assert.JSONEq(t, `2`, string(rawField(t, out, "phase")))
	})
}
