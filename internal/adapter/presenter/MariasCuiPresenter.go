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

// mariasSuitSymbols maps a suit constant (1-4) to its glyph for CUI display.
var mariasSuitSymbols = [...]string{"?", "♠", "♣", "♥", "♦"}

// mariasSuitSymbol maps a suit constant (1-4) to its glyph for CUI display.
func mariasSuitSymbol(suit int) string {
	if suit < 1 || suit > 4 {
		return "?"
	}
	return mariasSuitSymbols[suit]
}

func mariasPlayerStr(g interfaces.MariasGame, idx int) string {
	player := g.GetPlayer(idx)
	if player == nil {
		return ""
	}
	scores := g.GetPlayerScores()
	role := i18n.T("marias.roleDefender")
	if idx == g.GetSoloistIdx() {
		role = i18n.T("marias.roleSoloist")
	}
	var b strings.Builder
	b.WriteString(i18n.Tf("marias.playerLine",
		"name", cuiPlayerName(player, idx),
		"role", role,
		"cards", strconv.Itoa(player.GetCardsSize()),
		"score", strconv.Itoa(scores[idx]),
		"tricks", strconv.Itoa(player.GetTrickCount()),
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// MariasCuiPresenter renders the Mariáš CUI view.
type MariasCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *MariasCuiPresenter) Output(g interfaces.MariasGame, lastErr error) string {
	return buildCuiOutput(i18n.T("marias.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("marias.round",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"trick", strconv.Itoa(g.GetTrickNumber()),
			"trump", mariasSuitSymbol(g.GetTrumpSuit())) + "\n")

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(mariasPlayerStr(g, i))
		}

		b.WriteString("----------\n")

		cuiTrickBlock(b, g.GetCurrentTrick(),
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(g.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			winner := g.GetWinnerPlayer()
			var winnerStr string
			if winner >= 0 {
				winnerStr = cuiPlayerName(g.GetPlayer(winner), winner)
			}
			banner := i18n.Tf("marias.gameEnd", "name", winnerStr)
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		switch g.GetPhase() {
		case domain.MariasPhasePlay:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("marias.promptPlay",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("marias.promptPlayHelp") + "\n")
			b.WriteString(i18n.T("marias.promptMarriageHelp") + "\n")
			// **結婚は配札直後に確定して点が入る** (#5647)。Web は marias-marriage
			// バナーで常時出しているのに、CUI にはそれが成立したという表示が無く、
			// ラウンドが終わるまで自分の点に何が乗っているのか分からなかった。
			if marriage := g.GetRoundMarriage(); marriage[currentIdx] > 0 && g.GetPlayer(currentIdx).GetIsHuman() {
				b.WriteString(i18n.Tf("marias.marriageEarned",
					"points", strconv.Itoa(marriage[currentIdx])) + "\n")
			}
		case domain.MariasPhaseTrickEnd:
			b.WriteString(i18n.T("marias.promptTrickEnd") + "\n")
			b.WriteString(i18n.T("marias.promptTrickEndHelp") + "\n")
		case domain.MariasPhaseRoundEnd:
			pts := g.GetRoundCardPoints()
			// **勝敗はカード点 + 結婚点で決まる** (Marias.go の settle)。カード点
			// だけを合算していたので、結婚で勝ったソロイストが負けたように読めた
			// (#5647)。ドメインと同じ足し方にする。
			marriage := g.GetRoundMarriage()
			soloist := g.GetSoloistIdx()
			// Sum the two defenders' points so the outcome (soloist won or
			// lost the round) is readable from this line alone.
			defenderPts := 0
			for i := 0; i < g.GetPlayerCnt(); i++ {
				if i != soloist {
					defenderPts += pts[i] + marriage[i]
				}
			}
			b.WriteString(i18n.Tf("marias.promptRoundEnd",
				"soloist", cuiPlayerName(g.GetPlayer(soloist), soloist),
				"pts", strconv.Itoa(pts[soloist]+marriage[soloist])) + "\n")
			b.WriteString(i18n.Tf("marias.promptRoundEndDefenders",
				"pts", strconv.Itoa(defenderPts)) + "\n")
			b.WriteString(i18n.T("marias.promptRoundEndHelp") + "\n")
		}
	})
}

// HintOutput emits the current Mariáš hint.
func (p *MariasCuiPresenter) HintOutput(g interfaces.MariasGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("marias.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, mariasHintReasonKeys)
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
		return color.Yellow(i18n.Tf("marias.hintCard",
			"cards", strings.Join(cards, ", "),
			"reason", reason)) + "\n"
	}
	return color.Yellow(i18n.Tf("marias.hintCard",
		"cards", "-",
		"reason", reason)) + "\n"
}

// mariasHintReasonKeys maps Mariáš-specific hint-reason identifiers to i18n keys.
var mariasHintReasonKeys = map[string]string{
	"lead_low":    "marias.hintReasonLeadLow",
	"follow_win":  "marias.hintReasonFollowWin",
	"follow_duck": "marias.hintReasonFollowDuck",
	"discard_low": "marias.hintReasonDiscardLow",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *MariasCuiPresenter) ActionLogOutput(g interfaces.MariasGame) string {
	return actionLogOutputTextForSeats[*domain.MariasPlayer](g)
}
