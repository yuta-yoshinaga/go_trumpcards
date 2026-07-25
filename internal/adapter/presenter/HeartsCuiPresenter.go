package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// heartsPlayerStr returns the display string for a single Hearts player. The
// cumulative score is highlighted once a player passes 80% of the point limit,
// since reaching it ends the game (and loses).
func heartsPlayerStr(player *domain.HeartsPlayer, i, pointLimit int) string {
	var b strings.Builder
	cum := strconv.Itoa(player.GetCumulativeScore())
	if pointLimit > 0 && player.GetCumulativeScore()*100 >= pointLimit*80 {
		cum = color.Yellow(cum)
	}
	b.WriteString(i18n.Tf("hearts.playerLine",
		"name", cuiPlayerName(player, i),
		"cum", cum,
		"round", strconv.Itoa(player.GetRoundScore()),
		"cards", strconv.Itoa(player.GetCardsSize()),
		"tricks", strconv.Itoa(player.GetTrickCount()),
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// HeartsCuiPresenter renders the Hearts CUI view.
type HeartsCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (p *HeartsCuiPresenter) Output(h interfaces.HeartsGame, lastErr error) string {
	return buildCuiOutput(i18n.T("hearts.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("hearts.round",
			"round", strconv.Itoa(h.GetRoundNumber()),
			"trick", strconv.Itoa(h.GetTrickNumber())) + "\n")

		if h.GetHeartsBroken() {
			b.WriteString(i18n.T("hearts.heartsBrokenYes") + "\n")
		} else {
			b.WriteString(i18n.T("hearts.heartsBrokenNo") + "\n")
		}

		// Point-limit progress: the highest cumulative score is closest to ending
		// the game (whoever reaches the limit loses), so surface it and the limit.
		pointLimit := h.GetConfig().PointLimit
		leaderIdx, maxScore := 0, -1
		for i := 0; i < h.GetPlayerCnt(); i++ {
			if s := h.GetPlayer(i).GetCumulativeScore(); s > maxScore {
				maxScore, leaderIdx = s, i
			}
		}
		b.WriteString(i18n.Tf("hearts.limitProgress",
			"limit", strconv.Itoa(pointLimit),
			"name", cuiPlayerName(h.GetPlayer(leaderIdx), leaderIdx),
			"score", strconv.Itoa(maxScore)) + "\n")

		for i := 0; i < h.GetPlayerCnt(); i++ {
			b.WriteString(heartsPlayerStr(h.GetPlayer(i), i, pointLimit))
		}

		b.WriteString("----------\n")

		// Current trick
		trick := h.GetCurrentTrick()
		cuiTrickBlock(b, trick,
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(h.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(b, lastErr)

		// Game state
		if h.GetGameEndFlag() {
			winnerIdx := h.GetWinnerIdx()
			player := h.GetPlayer(winnerIdx)
			banner := i18n.Tf("hearts.gameEnd", "name", cuiPlayerName(player, winnerIdx))
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		switch h.GetPhase() {
		case domain.HeartsPhasePass:
			b.WriteString(i18n.Tf("hearts.promptPass",
				"direction", cuiPassDirectionStr(h.GetPassDirection())) + "\n")
			b.WriteString(i18n.T("hearts.promptPassHelp") + "\n")
		case domain.HeartsPhasePlay:
			currentIdx := h.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("hearts.promptPlay",
				"name", cuiPlayerName(h.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("hearts.promptPlayHelp") + "\n")
		case domain.HeartsPhaseTrickEnd:
			b.WriteString(i18n.T("hearts.promptTrickEnd") + "\n")
			b.WriteString(i18n.T("hearts.promptTrickEndHelp") + "\n")
		case domain.HeartsPhaseRoundEnd:
			b.WriteString(i18n.T("hearts.promptRoundEnd") + "\n")
			b.WriteString(i18n.T("hearts.promptRoundEndHelp") + "\n")
		}
	})
}

// HintOutput emits the current Hearts hint.
func (p *HeartsCuiPresenter) HintOutput(h interfaces.HeartsGame) string {
	hint := h.GetHint()
	if hint == nil {
		return i18n.T("hearts.hintNone") + "\n"
	}
	player := h.GetPlayer(0)
	cards := make([]string, len(hint.CardIndices))
	for i, idx := range hint.CardIndices {
		cards[i] = "[" + strconv.Itoa(idx) + "]" + cuiCardStr(player.GetCard(idx))
	}
	return color.Yellow(i18n.Tf("hearts.hintCard",
		"cards", strings.Join(cards, ", "),
		"reason", hintReasonStr(hint.Reason, heartsHintReasonKeys))) + "\n"
}

// heartsHintReasonKeys maps Hearts-specific hint-reason identifiers to
// their i18n keys. Reasons not in this map fall through to
// hintReasonStr → cui_common.
var heartsHintReasonKeys = map[string]string{
	"pass_high_risk_cards": "hearts.hintReasonPassHighRisk",
	"discard_queen_spades": "hearts.hintReasonDiscardQueenSpades",
	"discard_hearts":       "hearts.hintReasonDiscardHearts",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *HeartsCuiPresenter) ActionLogOutput(h interfaces.HeartsGame) string {
	return actionLogOutputText(h)
}

// cuiPassDirectionStr returns the localized label for a Hearts pass direction.
func cuiPassDirectionStr(dir domain.HeartsPassDirection) string {
	switch dir {
	case domain.HeartsPassLeft:
		return i18n.T("hearts.passLeft")
	case domain.HeartsPassRight:
		return i18n.T("hearts.passRight")
	case domain.HeartsPassAcross:
		return i18n.T("hearts.passAcross")
	case domain.HeartsPassNone:
		return i18n.T("hearts.passNone")
	default:
		return i18n.T("hearts.passUnknown")
	}
}
