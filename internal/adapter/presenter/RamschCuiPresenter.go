//go:build !js || !wasm || extra3

package presenter

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// ramschPlayerStr returns the display string for a single Ramsch player.
//
// **点は罰点なので「多いほど不利」と読ませる必要がある。** 集めた点をただ
// 並べると、多い人が勝っていると読まれる。
func ramschPlayerStr(player *domain.RamschPlayer, i int) string {
	var b strings.Builder
	name := cuiPlayerName(player, i)
	fmt.Fprintf(&b, "%s: tricks=%d cardPts=%d total=%d round=%d hand=%d\n",
		name,
		player.GetTrickCount(),
		player.GetCardPoints(),
		player.GetCumulativeScore(),
		player.GetRoundScore(),
		player.GetCardsSize(),
	)
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player))
		b.WriteString("\n")
	}
	return b.String()
}

// RamschCuiPresenter Ramsch CUI presenter.
type RamschCuiPresenter struct{}

// Output renders the game state as a CUI string.
func (p *RamschCuiPresenter) Output(s interfaces.RamschGame, lastErr error) string {
	return buildCuiOutput(i18n.T("ramsch.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("ramsch.statusLine",
			"round", strconv.Itoa(s.GetRoundNumber()),
			"trick", strconv.Itoa(s.GetTrickNumber()),
			"dealer", strconv.Itoa(s.GetDealerIdx()),
			"fore", strconv.Itoa(s.GetForehandIdx()),
			"mid", strconv.Itoa(s.GetMiddlehandIdx()),
			"rear", strconv.Itoa(s.GetRearhandIdx())) + "\n")

		// **切り札は常にジャック 4 枚。** 入札で決まる余地が無いことを毎回書く ——
		// スカート系のつもりで来た人は「何が切り札か」をまず探す。
		b.WriteString(i18n.T("ramsch.trumpFixed") + "\n")
		// **点は罰点。** これを言わないと、点を集めている人が勝っていると読まれる。
		b.WriteString(i18n.T("ramsch.scoringNote") + "\n")

		for i := 0; i < s.GetPlayerCnt(); i++ {
			b.WriteString(ramschPlayerStr(s.GetPlayer(i), i))
		}

		b.WriteString("----------\n")

		trick := s.GetCurrentTrick()
		cuiTrickBlock(b, trick,
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(s.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(b, lastErr)

		if s.GetGameEndFlag() {
			b.WriteString(color.Green(i18n.T("ramsch.gameOver")) + "\n")
			return
		}

		switch s.GetPhase() {
		case domain.RamschPhasePlay:
			currentIdx := s.GetCurrentPlayerIdx()
			player := s.GetPlayer(currentIdx)
			b.WriteString(i18n.Tf("ramsch.turnLabel", "name", cuiPlayerName(player, currentIdx)) + "\n")
			b.WriteString(i18n.T("ramsch.promptPlay") + "\n")
		case domain.RamschPhaseTrickEnd:
			b.WriteString(i18n.T("ramsch.trickComplete") + "\n")
			b.WriteString(i18n.T("ramsch.promptNext") + "\n")
		case domain.RamschPhaseRoundEnd:
			// **誰がいくつ取ったのかを全員分出す。** 罰点なので「最多が誰か」が
			// 結果そのものだが、差が 1 点のこともある。合計だけでは読めない。
			for i := 0; i < s.GetPlayerCnt(); i++ {
				b.WriteString(i18n.Tf("ramsch.roundPointsLine",
					"name", cuiPlayerName(s.GetPlayer(i), i),
					"points", strconv.Itoa(s.GetCardPoints(i))) + "\n")
			}
			if s.IsDurchmarsch() {
				idx := s.GetDurchmarschIdx()
				b.WriteString(color.Green(i18n.Tf("ramsch.durchmarsch",
					"name", cuiPlayerName(s.GetPlayer(idx), idx))) + "\n")
			} else if loser := s.GetLoserIdx(); loser >= 0 {
				b.WriteString(color.Red(i18n.Tf("ramsch.roundLoser",
					"name", cuiPlayerName(s.GetPlayer(loser), loser),
					"points", strconv.Itoa(s.GetCardPoints(loser)))) + "\n")
			} else {
				// 同点は全員が負う。誰か 1 人を選ぶ根拠が無い。
				b.WriteString(color.Red(i18n.T("ramsch.roundTied")) + "\n")
			}
			b.WriteString(i18n.T("ramsch.promptNextRound") + "\n")
		}
	})
}

// HintOutput renders the hint output.
func (p *RamschCuiPresenter) HintOutput(s interfaces.RamschGame) string {
	hint := s.GetHint()
	if hint == nil {
		return i18n.T("ramsch.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, ramschHintReasonKeys)
	if hint.CardIndex != nil {
		card := s.GetPlayer(0).GetCard(*hint.CardIndex)
		return color.Yellow(i18n.Tf("ramsch.hintPlay",
			"idx", strconv.Itoa(*hint.CardIndex), "card", cuiCardStr(card), "reason", reason)) + "\n"
	}
	return i18n.T("ramsch.hintNone") + "\n"
}

// ActionLogOutput returns the round's action log as text.
func (p *RamschCuiPresenter) ActionLogOutput(s interfaces.RamschGame) string {
	return actionLogOutputTextForSeats[*domain.RamschPlayer](s)
}

// ramschHintReasonKeys maps Ramsch-specific hint-reason identifiers to i18n keys.
var ramschHintReasonKeys = map[string]string{
	"avoid_points":   "ramsch.hintReasonAvoidPoints",
	"lead_low":       "ramsch.hintReasonLeadLow",
	"forced_discard": "ramsch.hintReasonForcedDiscard",
}
