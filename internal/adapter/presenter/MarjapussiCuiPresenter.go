//go:build !js || !wasm || extra5

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// marjapussiSuitSymbols maps a suit constant (1-4) to its glyph for CUI display.
var marjapussiSuitSymbols = [...]string{"?", "♠", "♣", "♥", "♦"}

// marjapussiSuitSymbol maps a suit constant (1-4) to its glyph for CUI display.
func marjapussiSuitSymbol(suit int) string {
	if suit < 1 || suit > 4 {
		return "?"
	}
	return marjapussiSuitSymbols[suit]
}

// marjapussiNearWinRatio は「目標に迫っている」とみなす到達率。
const marjapussiNearWinRatio = 0.8

// marjapussiNearWin は score が target の marjapussiNearWinRatio を超えたかを返す。
// target が 0 以下なら常に false (割り算が意味を持たない)。
func marjapussiNearWin(score, target int) bool {
	if target <= 0 {
		return false
	}
	return float64(score)/float64(target) > marjapussiNearWinRatio
}

func marjapussiPlayerStr(g interfaces.MarjapussiGame, idx int) string {
	player := g.GetPlayer(idx)
	if player == nil {
		return ""
	}
	team := idx % domain.MarjapussiTeamCnt
	scores := g.GetTeamScores()
	var b strings.Builder
	line := i18n.Tf("marjapussi.playerLine",
		"name", cuiPlayerName(player, idx),
		"team", strconv.Itoa(team),
		"cards", strconv.Itoa(player.GetCardsSize()),
		"score", strconv.Itoa(scores[team]),
		"tricks", strconv.Itoa(player.GetTrickCount()),
	)
	if marjapussiNearWin(scores[team], g.GetConfig().TargetPoints) {
		line = color.Yellow(line)
	}
	b.WriteString(line)
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// MarjapussiCuiPresenter renders the Marjapussi CUI view.
type MarjapussiCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *MarjapussiCuiPresenter) Output(g interfaces.MarjapussiGame, lastErr error) string {
	return buildCuiOutput(i18n.T("marjapussi.helpTitle"), func(b *strings.Builder) {
		trumpStr := i18n.T("marjapussi.trumpNone")
		if s := g.GetTrumpSuit(); s >= 1 && s <= 4 {
			trumpStr = marjapussiSuitSymbol(s)
		}
		b.WriteString(i18n.Tf("marjapussi.round",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"trick", strconv.Itoa(g.GetTrickNumber()),
			"trump", trumpStr) + "\n")

		scores := g.GetTeamScores()
		b.WriteString(i18n.Tf("marjapussi.headerInfo",
			"target", strconv.Itoa(g.GetConfig().TargetPoints),
			"score0", strconv.Itoa(scores[0]),
			"score1", strconv.Itoa(scores[1])) + "\n")

		marriageStr := i18n.T("marjapussi.marriageNone")
		for _, log := range g.GetActionLog() {
			if log.ActionType == "marriage" {
				marriageStr = log.Detail
			}
		}
		b.WriteString(i18n.Tf("marjapussi.headerMarriage",
			"marriage", marriageStr) + "\n")

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(marjapussiPlayerStr(g, i))
		}

		b.WriteString("----------\n")

		cuiTrickBlock(b, g.GetCurrentTrick(),
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(g.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			winnerTeam := g.GetWinnerTeam()
			banner := i18n.Tf("marjapussi.gameEnd", "team", strconv.Itoa(winnerTeam))
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		switch g.GetPhase() {
		case domain.MarjapussiPhasePlay:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("marjapussi.promptPlay",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("marjapussi.promptPlayHelp") + "\n")
			b.WriteString(i18n.T("marjapussi.promptMarriageHelp") + "\n")
			if player := g.GetPlayer(currentIdx); player != nil && player.GetIsHuman() {
				if opts := g.GetMarriageOptions(currentIdx); len(opts) > 0 {
					suits := make([]string, len(opts))
					for i, opt := range opts {
						suits[i] = cuiSuitName(opt.Suit) + " K-Q (+" + strconv.Itoa(opt.Points) + ")"
					}
					b.WriteString(i18n.Tf("marjapussi.promptMarriageReady",
						"suits", strings.Join(suits, ", ")) + "\n")
				}
			}
		case domain.MarjapussiPhaseTrickEnd:
			b.WriteString(i18n.T("marjapussi.promptTrickEnd") + "\n")
			b.WriteString(i18n.T("marjapussi.promptTrickEndHelp") + "\n")
		case domain.MarjapussiPhaseRoundEnd:
			b.WriteString(i18n.T("marjapussi.promptRoundEnd") + "\n")
			pussiTeam := -1
			pussiPts := 0
			for _, log := range g.GetActionLog() {
				if log.ActionType == "pussi_win" {
					pussiTeam = log.PlayerIdx % domain.MarjapussiTeamCnt
					for _, c := range log.Cards {
						pussiPts += marjapussiCardPoints(c)
					}
				}
			}
			if pussiTeam >= 0 {
				b.WriteString(i18n.Tf("marjapussi.promptRoundEndPussi",
					"team", strconv.Itoa(pussiTeam),
					"pts", strconv.Itoa(pussiPts)) + "\n")
			}
			b.WriteString(i18n.T("marjapussi.promptRoundEndHelp") + "\n")
		}
	})
}

// marjapussiCardPoints カードポイント (A=11, 10=10, K=4, Q=3, J=2, その他=0)。
func marjapussiCardPoints(c *domain.Card) int {
	if c == nil {
		return 0
	}
	switch c.GetValue() {
	case 1:
		return 11
	case 10:
		return 10
	case 13:
		return 4
	case 12:
		return 3
	case 11:
		return 2
	default:
		return 0
	}
}

// HintOutput emits the current Marjapussi hint.
func (p *MarjapussiCuiPresenter) HintOutput(g interfaces.MarjapussiGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("marjapussi.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, marjapussiHintReasonKeys)
	if len(hint.CardIndices) > 0 {
		playerIdx := g.GetCurrentPlayerIdx()
		player := g.GetPlayer(playerIdx)
		cards := make([]string, len(hint.CardIndices))
		for i, idx := range hint.CardIndices {
			if player != nil {
				cards[i] = "[" + strconv.Itoa(idx) + "]" + cuiCardStr(player.GetCard(idx))
			} else {
				cards[i] = strconv.Itoa(idx)
			}
		}
		return color.Yellow(i18n.Tf("marjapussi.hintCard",
			"cards", strings.Join(cards, ", "),
			"reason", reason)) + "\n"
	}
	return color.Yellow(i18n.Tf("marjapussi.hintCard",
		"cards", "-",
		"reason", reason)) + "\n"
}

// marjapussiHintReasonKeys maps Marjapussi-specific hint-reason identifiers to i18n keys.
var marjapussiHintReasonKeys = map[string]string{
	"lead_low":      "marjapussi.hintReasonLeadLow",
	"lead_marriage": "marjapussi.hintReasonLeadMarriage",
	"follow_win":    "marjapussi.hintReasonFollowWin",
	"follow_duck":   "marjapussi.hintReasonFollowDuck",
	"discard_low":   "marjapussi.hintReasonDiscardLow",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *MarjapussiCuiPresenter) ActionLogOutput(g interfaces.MarjapussiGame) string {
	return actionLogOutputTextForSeats[*domain.MarjapussiPlayer](g)
}
