//go:build !js || !wasm || casino

package presenter

import (
	"sort"
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// VideoPokerCuiPresenter ビデオポーカーCUIプレゼンタークラス
type VideoPokerCuiPresenter struct {
}

// Output ゲーム状態を出力
func (vpp *VideoPokerCuiPresenter) Output(vp interfaces.VideoPokerGame, lastErr error) string {
	var sb strings.Builder

	sb.WriteString("----------\n")
	sb.WriteString(i18n.Tf("videopoker.chipsLine", "chips", strconv.Itoa(vp.GetChips())) + "\n")
	sb.WriteString(i18n.Tf("videopoker.phaseLine", "phase", vpp.phaseStr(vp.GetPhase())) + "\n")

	hand := vp.GetHand()
	if len(hand) > 0 {
		sb.WriteString(i18n.T("videopoker.handHeader") + "\n")
		held := vp.GetHeldIndices()
		holdLabel := i18n.T("videopoker.holdLabel")
		parts := make([]string, len(hand))
		for i, card := range hand {
			s := vpp.cardStr(vp, card)
			if held[i] {
				s += " " + holdLabel
			}
			parts[i] = s
		}
		sb.WriteString(strings.Join(parts, ", "))
		sb.WriteString("\n")
	}

	// ベットフェーズではベット額決定の判断材料として配当表を表示する。
	if vp.GetPhase() == domain.VideoPokerPhaseBet {
		sb.WriteString(vpp.paytableStr(vp.GetVariantName()))
	}

	sb.WriteString("----------\n")

	if lastErr != nil {
		sb.WriteString(color.Red(lastErr.Error()) + "\n")
	}

	if vp.GetGameEndFlag() {
		sb.WriteString(i18n.Tf("videopoker.betLine", "bet", strconv.Itoa(vp.GetBetAmount())) + "\n")
		if vp.GetResult() == domain.GameResultWin {
			sb.WriteString(color.Green(i18n.Tf("videopoker.winLine", "handName", vpp.handNameForWin(vp))) + "\n")
		} else {
			sb.WriteString(color.Red(i18n.T("videopoker.noWin")) + "\n")
		}
		sb.WriteString(i18n.Tf("videopoker.payoutLine", "payout", strconv.Itoa(vp.GetPayout())) + "\n")
		sb.WriteString("----------\n")
	}

	return sb.String()
}

// ActionLogOutput 棋譜をテキスト出力
func (vpp *VideoPokerCuiPresenter) ActionLogOutput(vp interfaces.VideoPokerGame) string {
	return actionLogOutputText(vp)
}

// cardStr ワイルドカードを強調した手札カード文字列（ジョーカーは太字黄、Deuces Wildの2は黄）
func (vpp *VideoPokerCuiPresenter) cardStr(vp interfaces.VideoPokerGame, card *domain.Card) string {
	if card == nil {
		return cuiCardStr(card)
	}
	if card.GetDesign() == domain.CardDesignJoker {
		return color.BoldYellow("JOKER")
	}
	if card.GetValue() == 2 && vp.GetVariantName() == "deuceswild" {
		// 赤スートの通常色（赤）を上書きしないよう、素のスート名から組み立てる
		return color.Yellow(cuiSuitName(card.GetDesign()) + " 2")
	}
	return cuiCardStr(card)
}

// paytableStr はバリアント固有の配当表（役名と 1 コインあたりの倍率）を組み立てる。
// 配当値は domain.VideoPokerPaytable を単一情報源として参照する。
func (vpp *VideoPokerCuiPresenter) paytableStr(variantName string) string {
	var sb strings.Builder
	sb.WriteString(i18n.T("videopoker.payoutTitle") + "\n")
	for _, row := range domain.VideoPokerPaytable(variantName) {
		line := i18n.T("videopoker."+row.HandKey) + " x" + strconv.Itoa(row.Multiplier)
		if row.RoyalJackpot {
			line += " " + i18n.T("videopoker.payoutMaxBetNote")
		}
		sb.WriteString(line + "\n")
	}
	return sb.String()
}

// handNameForWin は勝利行に表示する役名を返す。Deuces Wild は安定キー
// (GetHandKey) 経由で "deuceswild.hand.<key>" を翻訳し、ja/en それぞれの
// ロケールで役名を表示する。他バリアント (videopoker / jokerpoker) は従来通り
// 英語の GetHandName をそのまま返し、後方互換を保つ。
func (vpp *VideoPokerCuiPresenter) handNameForWin(vp interfaces.VideoPokerGame) string {
	if vp.GetVariantName() == "deuceswild" {
		if key := vp.GetHandKey(); key != "" {
			return i18n.T("deuceswild.hand." + key)
		}
	}
	return vp.GetHandName()
}

// phaseStr フェーズ文字列
func (vpp *VideoPokerCuiPresenter) phaseStr(phase int) string {
	switch phase {
	case domain.VideoPokerPhaseBet:
		return i18n.T("videopoker.phaseBet")
	case domain.VideoPokerPhaseDraw:
		return i18n.T("videopoker.phaseDraw")
	case domain.VideoPokerPhaseResult:
		return i18n.T("videopoker.phaseResult")
	default:
		return i18n.T("videopoker.phaseUnknown")
	}
}

// videoPokerHighCardThreshold is the lowest rank (Jack) treated as a high card.
const videoPokerHighCardThreshold = 11

// videoPokerDrawCount is the minimum matching count that makes a flush or
// straight draw worth holding (4 of the 5 cards).
const videoPokerDrawCount = 4

// videoPokerIsWild reports whether a card is wild for the given variant. Deuces
// Wild treats every 2 as wild; Joker Poker treats the joker as wild; other
// variants (Jacks or Better) have no wild cards.
func videoPokerIsWild(variant string, c *domain.Card) bool {
	if c == nil {
		return false
	}
	switch variant {
	case "deuceswild":
		return c.GetValue() == 2
	case "jokerpoker":
		return c.GetDesign() == domain.CardDesignJoker
	default:
		return false
	}
}

// videoPokerHold computes the recommended hold mask and a reason i18n key
// (without the "videopoker." prefix) for a draw-phase hand. It is a faithful
// port of the frontend getVideoPokerBaseHint heuristic
// (frontend/src/utils/hints/videoPokerBaseHint.ts): hold wilds plus any made
// pair, otherwise keep the best made group, a four-card flush/straight draw, or
// the high cards. An empty mask (reason "") means redraw all five.
func videoPokerHold(hand []*domain.Card, variant string) ([]bool, string) {
	hold := make([]bool, len(hand))

	// Wild-aware: always hold wilds, then any made pair-or-better among the rest.
	var wildIdx []int
	for i, c := range hand {
		if videoPokerIsWild(variant, c) {
			wildIdx = append(wildIdx, i)
		}
	}
	if len(wildIdx) > 0 {
		for _, i := range wildIdx {
			hold[i] = true
		}
		pairIdx := videoPokerGroupIndices(hand, variant)
		if len(pairIdx) >= 2 {
			for _, i := range pairIdx {
				hold[i] = true
			}
			return hold, "holdWildAndPair"
		}
		return hold, "holdWild"
	}

	// Made pair/trips/quads: hold every rank group of two or more.
	groups := map[int][]int{}
	for i, c := range hand {
		groups[c.GetValue()] = append(groups[c.GetValue()], i)
	}
	var groupIdx []int
	maxCount := 0
	for _, idxs := range groups {
		if len(idxs) >= 2 {
			groupIdx = append(groupIdx, idxs...)
			if len(idxs) > maxCount {
				maxCount = len(idxs)
			}
		}
	}
	if len(groupIdx) > 0 {
		for _, i := range groupIdx {
			hold[i] = true
		}
		switch {
		case maxCount >= 4:
			return hold, "holdQuads"
		case maxCount >= 3:
			return hold, "holdTrips"
		default:
			return hold, "holdPair"
		}
	}

	// Four-card flush draw.
	if fd := videoPokerFlushDraw(hand); fd != nil {
		for _, i := range fd {
			hold[i] = true
		}
		return hold, "holdFlushDraw"
	}

	// Four-card straight draw.
	if sd := videoPokerStraightDraw(hand); sd != nil {
		for _, i := range sd {
			hold[i] = true
		}
		return hold, "holdStraightDraw"
	}

	// High cards (Jack or higher, plus the ace).
	var highIdx []int
	for i, c := range hand {
		if v := c.GetValue(); v == 1 || v >= videoPokerHighCardThreshold {
			highIdx = append(highIdx, i)
		}
	}
	if len(highIdx) > 0 {
		for _, i := range highIdx {
			hold[i] = true
		}
		return hold, "holdHighCards"
	}

	return hold, ""
}

// videoPokerGroupIndices returns the indices of all non-wild cards belonging to
// a rank group of two or more.
func videoPokerGroupIndices(hand []*domain.Card, variant string) []int {
	groups := map[int][]int{}
	for i, c := range hand {
		if videoPokerIsWild(variant, c) {
			continue
		}
		groups[c.GetValue()] = append(groups[c.GetValue()], i)
	}
	var idx []int
	for _, idxs := range groups {
		if len(idxs) >= 2 {
			idx = append(idx, idxs...)
		}
	}
	return idx
}

// videoPokerFlushDraw returns the indices of four or more same-suit cards, or
// nil if none.
func videoPokerFlushDraw(hand []*domain.Card) []int {
	suits := map[int][]int{}
	for i, c := range hand {
		suits[c.GetDesign()] = append(suits[c.GetDesign()], i)
	}
	for _, idxs := range suits {
		if len(idxs) >= videoPokerDrawCount {
			return idxs
		}
	}
	return nil
}

// videoPokerStraightDraw returns the indices of the longest run of four or more
// consecutive ranks (aces count both low and high), or nil if none.
func videoPokerStraightDraw(hand []*domain.Card) []int {
	type entry struct{ value, index int }
	entries := make([]entry, 0, len(hand)+1)
	for i, c := range hand {
		v := c.GetValue()
		entries = append(entries, entry{v, i})
		if v == 1 { // ace also plays high for 10-J-Q-K-A
			entries = append(entries, entry{14, i})
		}
	}
	sort.Slice(entries, func(a, b int) bool { return entries[a].value < entries[b].value })

	var best []entry
	for start := range entries {
		seq := []entry{entries[start]}
		for j := start + 1; j < len(entries); j++ {
			last := seq[len(seq)-1].value
			if entries[j].value == last+1 {
				seq = append(seq, entries[j])
			} else if entries[j].value != last {
				break
			}
		}
		if len(seq) > len(best) {
			best = seq
		}
	}
	if len(best) < videoPokerDrawCount {
		return nil
	}
	seen := map[int]bool{}
	var idx []int
	for _, e := range best {
		if !seen[e.index] {
			seen[e.index] = true
			idx = append(idx, e.index)
		}
	}
	return idx
}

// HintOutput recommends which cards to hold during the draw phase for any Video
// Poker variant (Jacks or Better, Deuces Wild, Joker Poker), based on the shared
// getVideoPokerBaseHint heuristic. Outside the draw phase it returns no hint.
func (p *VideoPokerCuiPresenter) HintOutput(g interfaces.VideoPokerGame) string {
	if g.GetPhase() != domain.VideoPokerPhaseDraw {
		return i18n.T("videopoker.hintNone") + "\n"
	}
	hand := g.GetHand()
	if len(hand) == 0 {
		return i18n.T("videopoker.hintNone") + "\n"
	}
	hold, reasonKey := videoPokerHold(hand, g.GetVariantName())
	var parts []string
	for i, h := range hold {
		if h {
			parts = append(parts, "["+strconv.Itoa(i)+"]"+cuiCardStr(hand[i]))
		}
	}
	if len(parts) == 0 {
		return color.Yellow(i18n.T("videopoker.hintHoldNone")) + "\n"
	}
	return color.Yellow(i18n.Tf("videopoker.hintHold",
		"cards", strings.Join(parts, " "),
		"reason", i18n.T("videopoker."+reasonKey),
	)) + "\n"
}
