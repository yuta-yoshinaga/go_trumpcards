//go:build !js || !wasm || extra3

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// stealingBundlesPlayerStr returns the display string for a single seat.
func stealingBundlesPlayerStr(s interfaces.StealingBundlesGame, idx int, current bool) string {
	player := s.GetPlayer(idx)
	var b strings.Builder
	marker := " "
	if current {
		marker = ">"
	}
	// **一番上だけが見えます。** そこが狙われる場所なので必ず出します。
	top := i18n.T("stealingbundles.bundleEmpty")
	if c := player.GetBundleTop(); c != nil {
		top = cuiCardStr(c)
	}
	b.WriteString(marker + i18n.Tf("stealingbundles.playerLine",
		"name", cuiPlayerName(player, idx),
		"cards", strconv.Itoa(player.GetCardsSize()),
		"bundle", strconv.Itoa(player.GetBundleSize()),
		"top", top,
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// stealingBundlesTableStr renders the table, or a placeholder when it is empty.
//
// **空の場は「取れる札が無い」という情報。** 何も書かないと行が消えてしまい、
// 場を見落としたのか空なのか区別が付きません。
func stealingBundlesTableStr(s interfaces.StealingBundlesGame) string {
	cards := s.GetTableCards()
	if len(cards) == 0 {
		return i18n.T("stealingbundles.tableEmpty")
	}
	return cuiCardSliceStr(cards)
}

// StealingBundlesCuiPresenter renders the Stealing Bundles CUI view.
type StealingBundlesCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *StealingBundlesCuiPresenter) Output(s interfaces.StealingBundlesGame, lastErr error) string {
	return buildCuiOutput(i18n.T("stealingbundles.helpTitle"), func(sb *strings.Builder) {
		sb.WriteString(i18n.Tf("stealingbundles.header",
			"turn", strconv.Itoa(s.GetTurnNumber()+1),
			"deck", strconv.Itoa(s.GetDeckRemaining())) + "\n")
		// **束の一番上が弱点、というのが規則そのもの。** 毎回書く。
		sb.WriteString(i18n.T("stealingbundles.rule") + "\n")

		sb.WriteString(i18n.Tf("stealingbundles.table",
			"cards", stealingBundlesTableStr(s)) + "\n")

		for i := 0; i < s.GetPlayerCnt(); i++ {
			sb.WriteString(stealingBundlesPlayerStr(s, i,
				i == s.GetCurrentPlayerIdx() && !s.GetGameEndFlag()))
		}

		sb.WriteString("----------\n")

		cuiErrorBlock(sb, lastErr)

		if s.GetGameEndFlag() {
			winner := s.GetWinnerIdx()
			var banner string
			if winner == 0 {
				banner = i18n.Tf("stealingbundles.gameEndYou",
					"n", strconv.Itoa(s.GetPlayer(winner).GetBundleSize()))
			} else {
				banner = i18n.Tf("stealingbundles.gameEndCpu",
					"name", cuiPlayerName(s.GetPlayer(winner), winner),
					"n", strconv.Itoa(s.GetPlayer(winner).GetBundleSize()))
			}
			sb.WriteString(color.Green(banner) + "\n")
			return
		}

		if !s.IsHumanTurn() {
			sb.WriteString(i18n.Tf("stealingbundles.promptCurrentPlayer",
				"name", cuiPlayerName(s.GetPlayer(s.GetCurrentPlayerIdx()), s.GetCurrentPlayerIdx())) + "\n")
			return
		}

		// **取れるときは置けません。** 黙っていると trail が弾かれる理由が分かりません。
		if s.CanCapture(0) {
			sb.WriteString(color.Yellow(i18n.T("stealingbundles.promptMustCapture")) + "\n")
			sb.WriteString(i18n.T("stealingbundles.promptTake") + "\n")
			sb.WriteString(i18n.T("stealingbundles.promptSteal") + "\n")
			return
		}
		sb.WriteString(i18n.T("stealingbundles.promptNoCapture") + "\n")
		sb.WriteString(i18n.T("stealingbundles.promptTrail") + "\n")
	})
}

// HintOutput emits the current hint.
func (p *StealingBundlesCuiPresenter) HintOutput(s interfaces.StealingBundlesGame) string {
	hint := s.GetHint()
	if hint == nil || hint.CardIndex == nil {
		return i18n.T("stealingbundles.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, stealingBundlesHintReasonKeys)
	card := s.GetPlayer(0).GetCard(*hint.CardIndex)
	if hint.VictimIdx >= 0 {
		return color.Yellow(i18n.Tf("stealingbundles.hintSteal",
			"idx", strconv.Itoa(*hint.CardIndex),
			"card", cuiCardStr(card),
			"victim", cuiPlayerName(s.GetPlayer(hint.VictimIdx), hint.VictimIdx),
			"reason", reason)) + "\n"
	}
	return color.Yellow(i18n.Tf("stealingbundles.hintCard",
		"idx", strconv.Itoa(*hint.CardIndex),
		"card", cuiCardStr(card),
		"reason", reason)) + "\n"
}

// stealingBundlesHintReasonKeys maps hint-reason identifiers to their i18n keys.
var stealingBundlesHintReasonKeys = map[string]string{
	"stealingbundlesSteal": "stealingbundles.hintReasonSteal",
	"stealingbundlesTake":  "stealingbundles.hintReasonTake",
	"stealingbundlesTrail": "stealingbundles.hintReasonTrail",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *StealingBundlesCuiPresenter) ActionLogOutput(s interfaces.StealingBundlesGame) string {
	return actionLogOutputText(s)
}
