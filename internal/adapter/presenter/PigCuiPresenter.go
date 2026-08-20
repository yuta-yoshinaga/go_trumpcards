//go:build !js || !wasm || extra2

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// pigPlayerStr returns the display string for a single seat.
func pigPlayerStr(s interfaces.PigGame, idx int, current bool) string {
	player := s.GetPlayer(idx)
	var b strings.Builder
	marker := " "
	if current {
		marker = ">"
	}
	role := ""
	switch {
	case player.GetEliminated():
		role = i18n.T("pig.roleOut")
	case player.GetNoticedOrder() > 0:
		role = i18n.Tf("pig.roleNoticed", "order", strconv.Itoa(player.GetNoticedOrder()))
	case s.HasChosenPass(idx):
		role = i18n.T("pig.roleChosen")
	}
	letters := player.GetLetterWord()
	if letters == "" {
		letters = "-"
	}
	b.WriteString(marker + i18n.Tf("pig.playerLine",
		"name", cuiPlayerName(player, idx),
		"role", role,
		"cards", strconv.Itoa(player.GetCardsSize()),
		"letters", letters,
		"target", domain.PigLetterTargetWord,
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// PigCuiPresenter renders the Pig CUI view.
type PigCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *PigCuiPresenter) Output(s interfaces.PigGame, lastErr error) string {
	return buildCuiOutput(i18n.T("pig.helpTitle"), func(sb *strings.Builder) {
		sb.WriteString(i18n.Tf("pig.header",
			"round", strconv.Itoa(s.GetRoundNumber()),
			"deck", strconv.Itoa(s.GetDeckSize()),
			"pass", strconv.Itoa(s.GetPassCount())) + "\n")
		// **罰の理由が直感と違う。** 取り合いではなく、気づくのが遅いことが負け。
		sb.WriteString(i18n.T("pig.rule") + "\n")

		for i := 0; i < s.GetPlayerCnt(); i++ {
			sb.WriteString(pigPlayerStr(s, i,
				i == s.GetCurrentPlayerIdx() && s.GetPhase() == domain.PigPhasePass && !s.GetGameEndFlag()))
		}

		sb.WriteString("----------\n")

		cuiErrorBlock(sb, lastErr)

		if s.GetGameEndFlag() {
			var banner string
			if s.GetWinnerIdx() == 0 {
				banner = i18n.T("pig.gameEndYou")
			} else {
				banner = i18n.Tf("pig.gameEndCpu",
					"name", cuiPlayerName(s.GetPlayer(s.GetWinnerIdx()), s.GetWinnerIdx()))
			}
			sb.WriteString(color.Green(banner) + "\n")
			return
		}

		switch s.GetPhase() {
		case domain.PigPhaseSignal:
			// **合図は声に出しません。** 気づいたことをこちらから名乗る必要がある。
			if s.GetPlayer(0).GetHasSignalled() {
				sb.WriteString(i18n.Tf("pig.promptSignalDone",
					"n", strconv.Itoa(s.GetNoticedCnt())) + "\n")
				return
			}
			sb.WriteString(color.Yellow(i18n.T("pig.promptSignal")) + "\n")
			sb.WriteString(i18n.T("pig.promptSignalCmd") + "\n")
		case domain.PigPhaseRoundEnd:
			loser := s.GetRoundLoserIdx()
			sb.WriteString(color.Yellow(i18n.Tf("pig.promptRoundEnd",
				"name", cuiPlayerName(s.GetPlayer(loser), loser),
				"word", s.GetPlayer(loser).GetLetterWord())) + "\n")
			sb.WriteString(i18n.T("pig.promptNext") + "\n")
		default:
			// **人間が脱落しても局は続く。**
			if s.GetPlayer(0).GetEliminated() {
				sb.WriteString(i18n.T("pig.promptEliminated") + "\n")
				return
			}
			if s.HasChosenPass(0) {
				// **同時に渡すので、全員が選ぶまで札は動きません。**
				sb.WriteString(i18n.T("pig.promptWaiting") + "\n")
				return
			}
			sb.WriteString(i18n.T("pig.promptPass") + "\n")
		}
	})
}

// HintOutput emits the current hint.
func (p *PigCuiPresenter) HintOutput(s interfaces.PigGame) string {
	hint := s.GetHint()
	if hint == nil {
		return i18n.T("pig.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, pigHintReasonKeys)
	if hint.CardIndex == nil {
		return color.Yellow(i18n.Tf("pig.hintSignal", "reason", reason)) + "\n"
	}
	card := s.GetPlayer(0).GetCard(*hint.CardIndex)
	return color.Yellow(i18n.Tf("pig.hintCard",
		"idx", strconv.Itoa(*hint.CardIndex),
		"card", cuiCardStr(card),
		"reason", reason)) + "\n"
}

// pigHintReasonKeys maps hint-reason identifiers to their i18n keys.
var pigHintReasonKeys = map[string]string{
	"pigSignal":      "pig.hintReasonSignal",
	"pigDiscardOdd":  "pig.hintReasonDiscardOdd",
	"pigNoSingleton": "pig.hintReasonNoSingleton",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *PigCuiPresenter) ActionLogOutput(s interfaces.PigGame) string {
	return actionLogOutputText(s)
}
