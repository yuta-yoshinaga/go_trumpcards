package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// gongZhuPlayerStr returns the display string for a single Gong Zhu player.
func gongZhuPlayerStr(player *domain.GongZhuPlayer, i int) string {
	var b strings.Builder
	b.WriteString(i18n.Tf("gongzhu.playerLine",
		"name", cuiPlayerName(player, i),
		"cum", strconv.Itoa(player.GetCumulativeScore()),
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

// GongZhuCuiPresenter renders the Gong Zhu CUI view.
type GongZhuCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *GongZhuCuiPresenter) Output(g interfaces.GongZhuGame, lastErr error) string {
	return buildCuiOutput(i18n.T("gongzhu.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("gongzhu.round",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"trick", strconv.Itoa(g.GetTrickNumber())) + "\n")

		b.WriteString(i18n.Tf("gongzhu.exposed", "cards", gongZhuExposureStr(g.GetExposure())) + "\n")

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(gongZhuPlayerStr(g.GetPlayer(i), i))
		}

		b.WriteString("----------\n")

		trick := g.GetCurrentTrick()
		cuiTrickBlock(b, trick,
			func(tc *domain.GongZhuTrickCard) int { return tc.PlayerIdx },
			func(tc *domain.GongZhuTrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(g.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			winnerIdx := g.GetWinnerIdx()
			player := g.GetPlayer(winnerIdx)
			banner := i18n.Tf("gongzhu.gameEnd", "name", cuiPlayerName(player, winnerIdx))
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		switch g.GetPhase() {
		case domain.GongZhuPhaseExpose:
			b.WriteString(i18n.T("gongzhu.promptExpose") + "\n")
			b.WriteString(i18n.T("gongzhu.promptExposeHelp") + "\n")
		case domain.GongZhuPhasePlay:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("gongzhu.promptPlay",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("gongzhu.promptPlayHelp") + "\n")
		case domain.GongZhuPhaseTrickEnd:
			b.WriteString(i18n.T("gongzhu.promptTrickEnd") + "\n")
			b.WriteString(i18n.T("gongzhu.promptTrickEndHelp") + "\n")
		case domain.GongZhuPhaseRoundEnd:
			b.WriteString(i18n.T("gongzhu.promptRoundEnd") + "\n")
			b.WriteString(i18n.T("gongzhu.promptRoundEndHelp") + "\n")
		}
	})
}

// gongZhuExposureStr renders the localized exposure summary.
func gongZhuExposureStr(ex domain.GongZhuExposure) string {
	var parts []string
	if ex.Pig {
		parts = append(parts, "♠Q")
	}
	if ex.Sheep {
		parts = append(parts, "♦J")
	}
	if ex.Ace {
		parts = append(parts, "♥A")
	}
	if ex.Doubler {
		parts = append(parts, "♣10")
	}
	if len(parts) == 0 {
		return i18n.T("gongzhu.exposedNone")
	}
	return strings.Join(parts, ", ")
}

// HintOutput emits the current Gong Zhu hint.
func (p *GongZhuCuiPresenter) HintOutput(g interfaces.GongZhuGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("gongzhu.hintNone") + "\n"
	}
	player := g.GetPlayer(0)
	cards := make([]string, len(hint.CardIndices))
	for i, idx := range hint.CardIndices {
		cards[i] = "[" + strconv.Itoa(idx) + "]" + cuiCardStr(player.GetCard(idx))
	}
	cardsStr := strings.Join(cards, ", ")
	if cardsStr == "" {
		cardsStr = i18n.T("gongzhu.exposedNone")
	}
	return color.Yellow(i18n.Tf("gongzhu.hintCard",
		"cards", cardsStr,
		"reason", gongZhuHintReasonStr(hint.Reason))) + "\n"
}

// gongZhuHintReasonKeys maps Gong Zhu-specific hint-reason identifiers to i18n keys.
var gongZhuHintReasonKeys = map[string]string{
	"expose_sheep":   "gongzhu.hintReasonExposeSheep",
	"expose_none":    "gongzhu.hintReasonExposeNone",
	"discard_pig":    "gongzhu.hintReasonDiscardPig",
	"discard_hearts": "gongzhu.hintReasonDiscardHearts",
}

// gongZhuHintReasonStr resolves a reason via the per-game map first, then the shared layer.
func gongZhuHintReasonStr(reason string) string {
	if key, ok := gongZhuHintReasonKeys[reason]; ok {
		return i18n.T(key)
	}
	return lookupHintReason(reason, nil)
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *GongZhuCuiPresenter) ActionLogOutput(g interfaces.GongZhuGame) string {
	return actionLogOutputText(g)
}
