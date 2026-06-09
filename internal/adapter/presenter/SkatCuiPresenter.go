//go:build !js || !wasm || casino

package presenter

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// skatPlayerStr returns the display string for a single Skat player.
func skatPlayerStr(player *domain.SkatPlayer, i int) string {
	var b strings.Builder
	name := cuiPlayerName(player, i)
	role := ""
	if player.GetIsDeclarer() {
		role = " [Declarer]"
	}
	bidStr := "-"
	if player.GetBid() == 0 {
		bidStr = "pass"
	} else if player.GetBid() > 0 {
		bidStr = fmt.Sprintf("%d", player.GetBid())
	}
	fmt.Fprintf(&b, "%s%s: bid=%s tricks=%d cardPts=%d total=%d round=%d hand=%d\n",
		name, role, bidStr,
		player.GetTrickCount(),
		player.GetCardPoints(),
		player.GetCumulativeScore(),
		player.GetRoundScore(),
		player.GetCardsSize(),
	)
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player))
		b.WriteString("\n")
	}
	return b.String()
}

// SkatCuiPresenter Skat CUI presenter.
type SkatCuiPresenter struct{}

// Output renders the game state as a CUI string.
func (p *SkatCuiPresenter) Output(s interfaces.SkatGame, lastErr error) string {
	return buildCuiOutput(i18n.T("skat.helpTitle"), func(b *strings.Builder) {
		fmt.Fprintf(b, "Round: %d  Trick: %d  Dealer: %d (Fore=%d / Mid=%d / Rear=%d)\n",
			s.GetRoundNumber(), s.GetTrickNumber(), s.GetDealerIdx(),
			s.GetForehandIdx(), s.GetMiddlehandIdx(), s.GetRearhandIdx())

		if s.GetGameType() != domain.SkatGameNone {
			fmt.Fprintf(b, "Game: %s", skatGameTypeLabel(s.GetGameType()))
			if s.GetGameType() == domain.SkatGameSuit {
				fmt.Fprintf(b, " (trump=%s)", skatSuitSymbol(s.GetTrumpSuit()))
			}
			b.WriteString("\n")
		}
		if s.GetCurrentBid() > 0 {
			fmt.Fprintf(b, "Current bid: %d\n", s.GetCurrentBid())
		}

		for i := 0; i < s.GetPlayerCnt(); i++ {
			b.WriteString(skatPlayerStr(s.GetPlayer(i), i))
		}

		b.WriteString("----------\n")

		trick := s.GetCurrentTrick()
		cuiTrickBlock(b, trick,
			func(tc *domain.SkatTrickCard) int { return tc.PlayerIdx },
			func(tc *domain.SkatTrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(s.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(b, lastErr)

		if s.GetGameEndFlag() {
			b.WriteString(color.Green("Game over!") + "\n")
			return
		}

		switch s.GetPhase() {
		case domain.SkatPhaseBid:
			actor := s.GetActiveBidActorIdx()
			if actor >= 0 {
				name := cuiPlayerName(s.GetPlayer(actor), actor)
				fmt.Fprintf(b, "Bidding: %s's turn\n", name)
			}
			b.WriteString("b 0/1 - pass (0) or accept the active bid step (1)\n")
		case domain.SkatPhaseSkatPickup:
			b.WriteString("Skat pickup: declarer decides\n")
			b.WriteString("ps 0/1 - decline (0) or pick up the skat (1)\n")
		case domain.SkatPhaseDiscard:
			b.WriteString("Discard 2 cards into the skat\n")
			b.WriteString("d <i> <j>\n")
		case domain.SkatPhaseGameDeclaration:
			b.WriteString("Game declaration\n")
			b.WriteString("g <1=Suit|2=Grand|3=Null> [trumpSuit 1-4]\n")
		case domain.SkatPhasePlay:
			currentIdx := s.GetCurrentPlayerIdx()
			player := s.GetPlayer(currentIdx)
			fmt.Fprintf(b, "Turn: %s\n", cuiPlayerName(player, currentIdx))
			b.WriteString("p <idx>\n")
		case domain.SkatPhaseTrickEnd:
			b.WriteString("Trick complete\n")
			b.WriteString("n / next\n")
		case domain.SkatPhaseRoundEnd:
			fmt.Fprintf(b, "Round end. Declarer points: %d / Defenders: %d / Game value: %d\n",
				s.GetDeclarerCardPoints(), s.GetDefendersCardPoints(), s.GetGameValue())
			b.WriteString("nr / nextround\n")
		}
	})
}

// HintOutput renders the hint output.
func (p *SkatCuiPresenter) HintOutput(s interfaces.SkatGame) string {
	hint := s.GetHint()
	if hint == nil {
		return i18n.T("skat.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, skatHintReasonKeys)
	if hint.Bid != nil {
		choice := i18n.T("skat.choiceBidPass")
		if *hint.Bid == 1 {
			choice = i18n.T("skat.choiceBidAccept")
		}
		return color.Yellow(i18n.Tf("skat.hintBid", "choice", choice, "reason", reason)) + "\n"
	}
	if hint.PickSkat != nil {
		choice := i18n.T("skat.choiceSkatDecline")
		if *hint.PickSkat {
			choice = i18n.T("skat.choiceSkatPickUp")
		}
		return color.Yellow(i18n.Tf("skat.hintPickSkat", "choice", choice, "reason", reason)) + "\n"
	}
	if hint.DiscardIndex != nil {
		card := s.GetPlayer(0).GetCard(*hint.DiscardIndex)
		return color.Yellow(i18n.Tf("skat.hintDiscard",
			"idx", strconv.Itoa(*hint.DiscardIndex), "card", cuiCardStr(card), "reason", reason)) + "\n"
	}
	if hint.GameType != nil {
		gt := domain.SkatGameType(*hint.GameType)
		s2 := skatGameTypeLabel(gt)
		if gt == domain.SkatGameSuit && hint.TrumpSuit != nil {
			s2 = fmt.Sprintf("%s %s", s2, skatSuitSymbol(*hint.TrumpSuit))
		}
		return color.Yellow(i18n.Tf("skat.hintDeclare", "game", s2, "reason", reason)) + "\n"
	}
	if hint.CardIndex != nil {
		card := s.GetPlayer(0).GetCard(*hint.CardIndex)
		return color.Yellow(i18n.Tf("skat.hintPlay",
			"idx", strconv.Itoa(*hint.CardIndex), "card", cuiCardStr(card), "reason", reason)) + "\n"
	}
	return i18n.T("skat.hintNone") + "\n"
}

// ActionLogOutput returns the round's action log as text.
func (p *SkatCuiPresenter) ActionLogOutput(s interfaces.SkatGame) string {
	return actionLogOutputText(s)
}

// skatGameTypeLabel returns the human-readable label for a Skat game type.
func skatGameTypeLabel(gt domain.SkatGameType) string {
	switch gt {
	case domain.SkatGameSuit:
		return "Suit"
	case domain.SkatGameGrand:
		return "Grand"
	case domain.SkatGameNull:
		return "Null"
	}
	return "None"
}

// skatSuitSymbol returns the suit symbol.
func skatSuitSymbol(suit int) string {
	switch suit {
	case domain.CardDesignSpade:
		return "♠"
	case domain.CardDesignClover:
		return "♣"
	case domain.CardDesignHeart:
		return "♥"
	case domain.CardDesignDiamond:
		return "♦"
	}
	return "?"
}

// skatHintReasonKeys maps Skat-specific hint-reason identifiers to i18n keys.
var skatHintReasonKeys = map[string]string{
	"strategic_bid": "skat.hintReasonStrategicBid",
	"skat_pickup":   "skat.hintReasonSkatPickup",
	"discard_low":   "skat.hintReasonDiscardLow",
	"game_choice":   "skat.hintReasonGameChoice",
	"best_play":     "skat.hintReasonBestPlay",
}
