package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// gongZhuIsPointCard reports whether a card scores in Gong Zhu: any heart, the
// pig (♠Q), the sheep (♦J), or the doubler (♣10).
func gongZhuIsPointCard(c *domain.Card) bool {
	switch {
	case c.GetDesign() == domain.CardDesignHeart:
		return true
	case c.GetDesign() == domain.CardDesignSpade && c.GetValue() == 12: // pig
		return true
	case c.GetDesign() == domain.CardDesignDiamond && c.GetValue() == 11: // sheep
		return true
	case c.GetDesign() == domain.CardDesignClover && c.GetValue() == 10: // doubler
		return true
	default:
		return false
	}
}

// gongZhuCapturedPoints returns the scoring cards a player has captured in
// their tricks, in capture order.
func gongZhuCapturedPoints(player *domain.GongZhuPlayer) []*domain.Card {
	var pts []*domain.Card
	for _, trick := range player.GetTricksTaken() {
		for _, c := range trick {
			if gongZhuIsPointCard(c) {
				pts = append(pts, c)
			}
		}
	}
	return pts
}

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
	// Show which scoring cards this player has taken (who holds the pig etc.);
	// players with none get no line, keeping the output compact.
	if pts := gongZhuCapturedPoints(player); len(pts) > 0 {
		b.WriteString(i18n.Tf("gongzhu.capturedLine",
			"cards", formatCardSlice(pts, cuiCardStr, " ")) + "\n")
	}
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
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return cuiCardStr(tc.Card) },
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
			// **なぜその点なのかを出す。**猪/羊の公開、全ハート、猪抜きの倍率が
			// 重なるので、最終値だけでは検算できない (#5630)。
			gongZhuWriteBreakdown(b, g)
			b.WriteString(i18n.T("gongzhu.promptRoundEnd") + "\n")
			b.WriteString(i18n.T("gongzhu.promptRoundEndHelp") + "\n")
		}
	})
}

// gongZhuWriteBreakdown writes each player's round-score breakdown.
//
// 起きていない項目は書かない ── 取っていない猪の行が出ると、何が起きたのか
// 読み取れなくなる。
func gongZhuWriteBreakdown(b *strings.Builder, g interfaces.GongZhuGame) {
	b.WriteString(i18n.T("gongzhu.breakdownTitle") + "\n")
	for i := 0; i < g.GetPlayerCnt(); i++ {
		d := g.ScoreBreakdownFor(i)
		b.WriteString(cuiPlayerName(g.GetPlayer(i), i) + "\n")
		if d.HeartCount > 0 {
			b.WriteString(i18n.Tf("gongzhu.breakdownHearts",
				"count", strconv.Itoa(d.HeartCount), "sum", strconv.Itoa(d.HeartsSum)) + "\n")
		}
		if d.AllHearts {
			b.WriteString(i18n.T("gongzhu.breakdownAllHearts") + "\n")
		}
		if d.AceExposed && d.HeartCount > 0 {
			b.WriteString(i18n.T("gongzhu.breakdownAceExposed") + "\n")
		}
		if d.HasPig {
			key := "gongzhu.breakdownPig"
			if d.PigExposed {
				key = "gongzhu.breakdownPigExposed"
			}
			b.WriteString(i18n.T(key) + "\n")
		}
		if d.HasSheep {
			key := "gongzhu.breakdownSheep"
			if d.SheepExposed {
				key = "gongzhu.breakdownSheepExposed"
			}
			b.WriteString(i18n.T(key) + "\n")
		}
		if d.DoublerMultiplier > 0 {
			b.WriteString(i18n.Tf("gongzhu.breakdownSubtotal", "subtotal", strconv.Itoa(d.Subtotal)) + "\n")
			b.WriteString(i18n.Tf("gongzhu.breakdownDoubler", "mult", strconv.Itoa(d.DoublerMultiplier)) + "\n")
		}
		if d.DoublerStandalone != 0 {
			b.WriteString(i18n.Tf("gongzhu.breakdownDoublerStandalone",
				"points", strconv.Itoa(d.DoublerStandalone)) + "\n")
		}
		b.WriteString(i18n.Tf("gongzhu.breakdownTotal", "total", strconv.Itoa(d.Total)) + "\n")
	}
}

// gongZhuExposureStr renders the localized exposure summary.
func gongZhuExposureStr(ex domain.GongZhuExposure) string {
	var parts []string
	if ex.Pig {
		parts = append(parts, i18n.T("gongzhu.card.spadeQueen"))
	}
	if ex.Sheep {
		parts = append(parts, i18n.T("gongzhu.card.diamondJack"))
	}
	if ex.Ace {
		parts = append(parts, i18n.T("gongzhu.card.heartAce"))
	}
	if ex.Doubler {
		parts = append(parts, i18n.T("gongzhu.card.clubTen"))
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
		"reason", hintReasonStr(hint.Reason, gongZhuHintReasonKeys))) + "\n"
}

// gongZhuHintReasonKeys maps Gong Zhu-specific hint-reason identifiers to i18n keys.
var gongZhuHintReasonKeys = map[string]string{
	"expose_sheep":   "gongzhu.hintReasonExposeSheep",
	"expose_none":    "gongzhu.hintReasonExposeNone",
	"discard_pig":    "gongzhu.hintReasonDiscardPig",
	"discard_hearts": "gongzhu.hintReasonDiscardHearts",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *GongZhuCuiPresenter) ActionLogOutput(g interfaces.GongZhuGame) string {
	return actionLogOutputText(g)
}
