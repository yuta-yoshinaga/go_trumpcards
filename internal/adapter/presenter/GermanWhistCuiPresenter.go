//go:build !js || !wasm || classic

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// germanWhistPlayerStr returns the display string for a single German Whist player.
func germanWhistPlayerStr(player *domain.GermanWhistPlayer, idx int) string {
	var b strings.Builder
	b.WriteString(i18n.Tf("germanwhist.playerLine",
		"name", cuiPlayerName(player, idx),
		"tricks", strconv.Itoa(player.GetTrickCount()),
		"scoring", strconv.Itoa(player.GetScoringTricks()),
		"cards", strconv.Itoa(player.GetCardsSize()),
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// GermanWhistCuiPresenter renders the German Whist CUI view.
type GermanWhistCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *GermanWhistCuiPresenter) Output(g interfaces.GermanWhistGame, lastErr error) string {
	return buildCuiOutput(i18n.T("germanwhist.helpTitle"), func(sb *strings.Builder) {
		phaseKey := "germanwhist.phaseFirst"
		if g.GetPhase() != domain.GermanWhistPhaseDraw {
			phaseKey = "germanwhist.phaseSecond"
		}
		sb.WriteString(i18n.Tf("germanwhist.header",
			"trick", strconv.Itoa(g.GetTrickNumber()+1),
			"stock", strconv.Itoa(g.GetStockCount()),
			"phase", i18n.T(phaseKey)) + "\n")

		sb.WriteString(i18n.Tf("germanwhist.trumpLine",
			"suit", germanWhistSuitName(g.GetTrumpSuit())) + "\n")

		// 前半のあいだだけ、奪い合う 1 枚が表向きで場に出ている。
		if up := g.GetUpCard(); up != nil {
			sb.WriteString(i18n.Tf("germanwhist.upCardLine", "card", cuiCardStr(up)) + "\n")
		} else {
			sb.WriteString(i18n.T("germanwhist.upCardNone") + "\n")
		}

		for i := 0; i < g.GetPlayerCnt(); i++ {
			sb.WriteString(germanWhistPlayerStr(g.GetPlayer(i), i))
		}

		sb.WriteString("----------\n")

		cuiTrickBlock(sb, g.GetCurrentTrick(),
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(g.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(sb, lastErr)

		if g.GetGameEndFlag() {
			s0 := g.GetPlayer(0).GetScoringTricks()
			s1 := g.GetPlayer(1).GetScoringTricks()
			var banner string
			switch g.GetWinnerIdx() {
			case 0:
				banner = i18n.Tf("germanwhist.gameEndP0", "p0", strconv.Itoa(s0), "p1", strconv.Itoa(s1))
			case 1:
				banner = i18n.Tf("germanwhist.gameEndP1", "p0", strconv.Itoa(s0), "p1", strconv.Itoa(s1))
			default:
				banner = i18n.Tf("germanwhist.gameEndTie", "p0", strconv.Itoa(s0), "p1", strconv.Itoa(s1))
			}
			sb.WriteString(color.Green(banner) + "\n")
			return
		}

		currentIdx := g.GetCurrentPlayerIdx()
		sb.WriteString(i18n.Tf("germanwhist.promptCurrentPlayer",
			"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
		sb.WriteString(i18n.T("germanwhist.promptPlay") + "\n")
	})
}

// germanWhistSuitName スート番号を i18n のスート名に変換する。
func germanWhistSuitName(suit int) string {
	switch suit {
	case domain.CardDesignSpade:
		return i18n.T("germanwhist.suitSpade")
	case domain.CardDesignClover:
		return i18n.T("germanwhist.suitClover")
	case domain.CardDesignHeart:
		return i18n.T("germanwhist.suitHeart")
	case domain.CardDesignDiamond:
		return i18n.T("germanwhist.suitDiamond")
	default:
		return "?"
	}
}

// HintOutput emits the current German Whist hint.
func (p *GermanWhistCuiPresenter) HintOutput(g interfaces.GermanWhistGame) string {
	hint := g.GetHint()
	if hint == nil || hint.CardIndex == nil {
		return i18n.T("germanwhist.hintNone") + "\n"
	}
	card := g.GetPlayer(0).GetCard(*hint.CardIndex)
	return color.Yellow(i18n.Tf("germanwhist.hintCard",
		"idx", strconv.Itoa(*hint.CardIndex),
		"card", cuiCardStr(card),
		"reason", hintReasonStr(hint.Reason, germanWhistHintReasonKeys))) + "\n"
}

// germanWhistHintReasonKeys maps German Whist hint-reason identifiers to their
// i18n keys. Reasons not listed here fall through to cui_common via
// hintReasonStr.
var germanWhistHintReasonKeys = map[string]string{
	"germanWhistTakeUpCard": "germanwhist.hintReasonTakeUpCard",
	"germanWhistDuck":       "germanwhist.hintReasonDuck",
	"germanWhistWinTrick":   "germanwhist.hintReasonWinTrick",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *GermanWhistCuiPresenter) ActionLogOutput(g interfaces.GermanWhistGame) string {
	return actionLogOutputText(g)
}
