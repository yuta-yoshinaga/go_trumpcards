//go:build !js || !wasm || extra

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// goofspielPlayerStr returns the display string for a single seat.
func goofspielPlayerStr(s interfaces.GoofspielGame, idx int) string {
	player := s.GetPlayer(idx)
	var b strings.Builder
	role := ""
	revealed := s.GetRevealedBids()
	switch {
	case s.GetPhase() == domain.GoofspielPhaseReveal && idx < len(revealed) && revealed[idx] != nil:
		role = i18n.Tf("goofspiel.roleRevealed", "card", cuiCardStr(revealed[idx]))
	case s.HasBid(idx):
		// **伏せたことだけを出します。** 中身は公開まで見せません。
		role = i18n.T("goofspiel.roleBid")
	}
	b.WriteString(i18n.Tf("goofspiel.playerLine",
		"name", cuiPlayerName(player, idx),
		"role", role,
		"cards", strconv.Itoa(player.GetCardsSize()),
		"score", strconv.Itoa(player.GetScore()),
	))
	b.WriteString("\n")
	// **残り手札は全員分を出します。** 使った札は場に出るので隠せていません。
	if player.GetCardsSize() > 0 {
		if player.GetIsHuman() {
			b.WriteString(cuiIndexedCardListStr(player) + "\n")
		} else {
			b.WriteString("  " + cuiCardListStr(player) + "\n")
		}
	}
	return b.String()
}

// GoofspielCuiPresenter renders the Goofspiel CUI view.
type GoofspielCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *GoofspielCuiPresenter) Output(s interfaces.GoofspielGame, lastErr error) string {
	return buildCuiOutput(i18n.T("goofspiel.helpTitle"), func(sb *strings.Builder) {
		sb.WriteString(i18n.Tf("goofspiel.header",
			"round", strconv.Itoa(s.GetRoundNumber()),
			"left", strconv.Itoa(s.GetPrizeRemaining())) + "\n")
		// **同時入札であることが規則そのもの。** 毎回書く。
		sb.WriteString(i18n.T("goofspiel.rule") + "\n")

		if prize := s.GetCurrentPrize(); prize != nil {
			sb.WriteString(i18n.Tf("goofspiel.prize",
				"card", cuiCardStr(prize),
				"n", strconv.Itoa(s.PrizeValue())) + "\n")
			// **持ち越しは「今回の賞が増える」こと。** 見えないと計算が合いません。
			if carried := s.GetCarriedPrizes(); len(carried) > 0 {
				sb.WriteString(i18n.Tf("goofspiel.carried", "n", strconv.Itoa(len(carried))) + "\n")
			}
		}

		for i := 0; i < s.GetPlayerCnt(); i++ {
			sb.WriteString(goofspielPlayerStr(s, i))
		}

		sb.WriteString("----------\n")

		cuiErrorBlock(sb, lastErr)

		if s.GetGameEndFlag() {
			winner := s.GetWinnerIdx()
			var banner string
			if winner == 0 {
				banner = i18n.Tf("goofspiel.gameEndYou", "n", strconv.Itoa(s.GetPlayer(winner).GetScore()))
			} else {
				banner = i18n.Tf("goofspiel.gameEndCpu",
					"name", cuiPlayerName(s.GetPlayer(winner), winner),
					"n", strconv.Itoa(s.GetPlayer(winner).GetScore()))
			}
			sb.WriteString(color.Green(banner) + "\n")
			return
		}

		if s.GetPhase() == domain.GoofspielPhaseReveal {
			if s.GetLastWinnerIdx() < 0 {
				sb.WriteString(color.Yellow(i18n.T("goofspiel.promptTie")) + "\n")
			} else {
				sb.WriteString(color.Yellow(i18n.Tf("goofspiel.promptRoundEnd",
					"name", cuiPlayerName(s.GetPlayer(s.GetLastWinnerIdx()), s.GetLastWinnerIdx()),
					"n", strconv.Itoa(s.GetLastGained()))) + "\n")
			}
			sb.WriteString(i18n.T("goofspiel.promptNext") + "\n")
			return
		}

		if s.HasBid(0) {
			sb.WriteString(i18n.T("goofspiel.promptWaiting") + "\n")
			return
		}
		sb.WriteString(i18n.T("goofspiel.promptBid") + "\n")
	})
}

// HintOutput emits the current hint.
func (p *GoofspielCuiPresenter) HintOutput(s interfaces.GoofspielGame) string {
	hint := s.GetHint()
	if hint == nil || hint.CardIndex == nil {
		return i18n.T("goofspiel.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, goofspielHintReasonKeys)
	card := s.GetPlayer(0).GetCard(*hint.CardIndex)
	return color.Yellow(i18n.Tf("goofspiel.hintCard",
		"idx", strconv.Itoa(*hint.CardIndex),
		"card", cuiCardStr(card),
		"reason", reason)) + "\n"
}

// goofspielHintReasonKeys maps hint-reason identifiers to their i18n keys.
var goofspielHintReasonKeys = map[string]string{
	"goofspielMatch":     "goofspiel.hintReasonMatch",
	"goofspielHighPrize": "goofspiel.hintReasonHighPrize",
	"goofspielLowPrize":  "goofspiel.hintReasonLowPrize",
	"goofspielCarried":   "goofspiel.hintReasonCarried",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *GoofspielCuiPresenter) ActionLogOutput(s interfaces.GoofspielGame) string {
	return actionLogOutputTextForSeats[*domain.GoofspielPlayer](s)
}
