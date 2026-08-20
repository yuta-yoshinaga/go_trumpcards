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

// tysiacSuitSymbols maps a suit constant (1-4) to its glyph for CUI display.
var tysiacSuitSymbols = [...]string{"?", "♠", "♣", "♥", "♦"}

// tysiacSuitSymbol maps a suit constant (1-4) to its glyph for CUI display.
func tysiacSuitSymbol(suit int) string {
	if suit < 1 || suit > 4 {
		return "?"
	}
	return tysiacSuitSymbols[suit]
}

func tysiacPlayerStr(g interfaces.TysiacGame, idx int) string {
	player := g.GetPlayer(idx)
	if player == nil {
		return ""
	}
	scores := g.GetPlayerScores()
	role := i18n.T("tysiac.rolePlayer")
	if idx == g.GetDeclarerIdx() {
		role = i18n.T("tysiac.roleDeclarer")
	}
	var b strings.Builder
	b.WriteString(i18n.Tf("tysiac.playerLine",
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

// TysiacCuiPresenter renders the Tysiąc CUI view.
type TysiacCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *TysiacCuiPresenter) Output(g interfaces.TysiacGame, lastErr error) string {
	return buildCuiOutput(i18n.T("tysiac.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("tysiac.round",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"trick", strconv.Itoa(g.GetTrickNumber()),
			"trump", tysiacSuitSymbol(g.GetTrumpSuit())) + "\n")

		// Surface the declarer's contract obligation and the match target during
		// play (not only at round end), plus the live bid while bidding.
		contractStr := "-"
		if c := g.GetContract(); c > 0 {
			contractStr = strconv.Itoa(c)
		}
		b.WriteString(i18n.Tf("tysiac.headerInfo",
			"contract", contractStr,
			"target", strconv.Itoa(g.GetConfig().TargetPoints)) + "\n")
		if g.GetPhase() == domain.TysiacPhaseBid {
			b.WriteString(i18n.Tf("tysiac.headerBid",
				"bid", strconv.Itoa(g.GetCurrentBid())) + "\n")
		}

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(tysiacPlayerStr(g, i))
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
			banner := i18n.Tf("tysiac.gameEnd", "name", winnerStr)
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		switch g.GetPhase() {
		case domain.TysiacPhaseBid:
			bidderIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("tysiac.promptBid",
				"bid", strconv.Itoa(g.GetCurrentBid()),
				"name", cuiPlayerName(g.GetPlayer(bidderIdx), bidderIdx)) + "\n")
			b.WriteString(i18n.T("tysiac.promptBidHelp") + "\n")
		case domain.TysiacPhaseTalon:
			declarerIdx := g.GetDeclarerIdx()
			b.WriteString(i18n.Tf("tysiac.promptTalon",
				"name", cuiPlayerName(g.GetPlayer(declarerIdx), declarerIdx)) + "\n")
			b.WriteString(i18n.T("tysiac.promptTalonHelp") + "\n")
		case domain.TysiacPhasePlay:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("tysiac.promptPlay",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("tysiac.promptPlayHelp") + "\n")
			b.WriteString(i18n.T("tysiac.promptMarriageHelp") + "\n")
			// 宣言できるスートは手札を数えないと分からないので、Web のバナーと同じく
			// 具体的に挙げる。CPU 番では出さない (手札が読めてしまう)。
			if player := g.GetPlayer(currentIdx); player != nil && player.GetIsHuman() {
				if opts := g.GetMarriageOptions(currentIdx); len(opts) > 0 {
					suits := make([]string, len(opts))
					for i, opt := range opts {
						suits[i] = cuiSuitName(opt.Suit) + " K-Q (+" + strconv.Itoa(opt.Points) + ")"
					}
					b.WriteString(i18n.Tf("tysiac.promptMarriageReady",
						"suits", strings.Join(suits, ", ")) + "\n")
				}
			}
		case domain.TysiacPhaseTrickEnd:
			b.WriteString(i18n.T("tysiac.promptTrickEnd") + "\n")
			b.WriteString(i18n.T("tysiac.promptTrickEndHelp") + "\n")
		case domain.TysiacPhaseRoundEnd:
			pts := g.GetRoundCardPoints()
			declarer := g.GetDeclarerIdx()
			declarerPts := 0
			if declarer >= 0 && declarer < len(pts) {
				declarerPts = pts[declarer]
			}
			b.WriteString(i18n.Tf("tysiac.promptRoundEnd",
				"declarer", cuiPlayerName(g.GetPlayer(declarer), declarer),
				"contract", strconv.Itoa(g.GetContract()),
				"pts", strconv.Itoa(declarerPts)) + "\n")
			b.WriteString(i18n.T("tysiac.promptRoundEndHelp") + "\n")
		}
	})
}

// HintOutput emits the current Tysiąc hint.
func (p *TysiacCuiPresenter) HintOutput(g interfaces.TysiacGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("tysiac.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, tysiacHintReasonKeys)
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
		return color.Yellow(i18n.Tf("tysiac.hintCard",
			"cards", strings.Join(cards, ", "),
			"reason", reason)) + "\n"
	}
	return color.Yellow(i18n.Tf("tysiac.hintCard",
		"cards", "-",
		"reason", reason)) + "\n"
}

// tysiacHintReasonKeys maps Tysiąc-specific hint-reason identifiers to i18n keys.
var tysiacHintReasonKeys = map[string]string{
	"lead_low":      "tysiac.hintReasonLeadLow",
	"lead_marriage": "tysiac.hintReasonLeadMarriage",
	"follow_win":    "tysiac.hintReasonFollowWin",
	"follow_duck":   "tysiac.hintReasonFollowDuck",
	"discard_low":   "tysiac.hintReasonDiscardLow",
	"bid_raise":     "tysiac.hintReasonBidRaise",
	"bid_pass":      "tysiac.hintReasonBidPass",
	"talon_discard": "tysiac.hintReasonTalonDiscard",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *TysiacCuiPresenter) ActionLogOutput(g interfaces.TysiacGame) string {
	return actionLogOutputTextForSeats[*domain.TysiacPlayer](g)
}
