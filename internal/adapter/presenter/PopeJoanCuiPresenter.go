//go:build !js || !wasm || extra3

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func popeJoanCardListStr(cards []*domain.Card, indexed bool) string {
	if len(cards) == 0 {
		return "-"
	}
	parts := make([]string, 0, len(cards))
	for i, c := range cards {
		if indexed {
			parts = append(parts, "["+strconv.Itoa(i)+"]"+cuiCardStr(c))
			continue
		}
		parts = append(parts, cuiCardStr(c))
	}
	return strings.Join(parts, " ")
}

// popeJoanCompartmentLabel は区画の表示名を返す。
func popeJoanCompartmentLabel(c domain.PopeJoanCompartment) string {
	return i18n.T("popejoan.compartment." + c.String())
}

// PopeJoanCuiPresenter renders the Pope Joan CUI view.
type PopeJoanCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *PopeJoanCuiPresenter) Output(c interfaces.PopeJoanGame, lastErr error) string {
	return buildCuiOutput(i18n.T("popejoan.helpTitle"), func(sb *strings.Builder) {
		turnUp := "-"
		if t := c.GetTurnUp(); t != nil {
			turnUp = cuiCardStr(t)
		}
		sb.WriteString(i18n.Tf("popejoan.header",
			"deal", strconv.Itoa(c.GetDealNumber()+1),
			"target", strconv.Itoa(c.GetConfig().TargetDeals),
			"turnup", turnUp) + "\n")
		// 「区画はトランプでしか払わない」と「♦8 が無いので ♦7 で必ず止まる」の
		// 2 点が最も誤解されやすいので毎回出す。
		sb.WriteString(i18n.T("popejoan.ruleLine") + "\n")

		board := c.GetBoard()
		comps := make([]string, 0, domain.PopeJoanCompartmentCount)
		for i := range domain.PopeJoanCompartmentCount {
			comp := domain.PopeJoanCompartment(i)
			comps = append(comps, popeJoanCompartmentLabel(comp)+":"+strconv.Itoa(board.Get(comp)))
		}
		sb.WriteString(i18n.Tf("popejoan.boardLine", "compartments", strings.Join(comps, " ")) + "\n")

		for _, a := range c.GetAwards() {
			if a == nil {
				continue
			}
			key := "popejoan.awardLine"
			if a.ByTurnUp {
				key = "popejoan.awardTurnUpLine"
			}
			sb.WriteString(i18n.Tf(key,
				"name", cuiPlayerName(c.GetPlayer(a.Player), a.Player),
				"compartment", popeJoanCompartmentLabel(a.Compartment),
				"chips", strconv.Itoa(a.Chips)) + "\n")
		}

		if pile := c.GetPlayedPile(); len(pile) > 0 {
			sb.WriteString(i18n.Tf("popejoan.pileLine", "cards", popeJoanCardListStr(pile, false)) + "\n")
		}

		for i, player := range c.GetPlayers() {
			if player == nil {
				continue
			}
			line := i18n.Tf("popejoan.playerLine",
				"name", cuiPlayerName(player, i),
				"chips", strconv.Itoa(player.GetChips()),
				"cards", strconv.Itoa(player.GetCardsSize()))
			// **Pope を抱えている人はその区画への支払いを免除される**ので、
			// 伏せ手でも公開してよい情報 (Web も CPU に出している)。
			if domain.PopeJoanHoldsPope(player) {
				line += "  " + i18n.T("popejoan.holdsPope")
			}
			sb.WriteString(line + "\n")
			if player.GetIsHuman() && player.GetCardsSize() > 0 {
				hand := make([]*domain.Card, 0, player.GetCardsSize())
				for j := range player.GetCardsSize() {
					hand = append(hand, player.GetCard(j))
				}
				sb.WriteString("  " + popeJoanCardListStr(hand, true) + "\n")
			}
		}

		sb.WriteString("----------\n")
		cuiErrorBlock(sb, lastErr)
		sb.WriteString(p.promptBlock(c))
	})
}

// promptBlock はフェーズごとの案内を返す。
func (p *PopeJoanCuiPresenter) promptBlock(c interfaces.PopeJoanGame) string {
	if c.GetGameEndFlag() {
		key := "popejoan.gameEndLose"
		if c.GetWinnerIdx() == 0 {
			key = "popejoan.gameEndWin"
		}
		return color.Green(i18n.T(key)) + "\n"
	}
	switch c.GetPhase() {
	case domain.PopeJoanPhasePlay:
		// **並びが止まっているかどうかで出せる札がまるで違う。**
		if c.GetRunSuit() < 0 {
			return i18n.T("popejoan.promptLead") + "\n"
		}
		return i18n.T("popejoan.promptFollow") + "\n"
	case domain.PopeJoanPhaseDealEnd:
		var sb strings.Builder
		if w := c.GetDealWinner(); w >= 0 {
			sb.WriteString(i18n.Tf("popejoan.dealResult",
				"name", cuiPlayerName(c.GetPlayer(w), w)) + "\n")
		}
		sb.WriteString(i18n.T("popejoan.promptNext") + "\n")
		return sb.String()
	case domain.PopeJoanPhaseGameEnd:
		return ""
	}
	return ""
}

// HintOutput emits the current Pope Joan hint.
func (p *PopeJoanCuiPresenter) HintOutput(c interfaces.PopeJoanGame) string {
	hint := popeJoanHint(c)
	key := popeJoanHintReasonKeys[hint.Reason]
	if key == "" {
		key = "popejoan.hintNone"
	}
	if hint.CardIndex != nil {
		return color.Yellow(i18n.Tf("popejoan.hintPlay",
			"idx", strconv.Itoa(*hint.CardIndex), "reason", i18n.T(key))) + "\n"
	}
	return color.Yellow(i18n.T(key)) + "\n"
}

// popeJoanHintReasonKeys maps the reason identifiers popeJoanHint returns to
// i18n keys. The Web presenter ships the identifier and the frontend resolves
// it; the CUI must resolve it here or it prints the raw key.
var popeJoanHintReasonKeys = map[string]string{
	"popejoan.hint.game_end":      "popejoan.hintReasonGameEnd",
	"popejoan.hint.deal_end":      "popejoan.hintReasonDealEnd",
	"popejoan.hint.not_your_turn": "popejoan.hintReasonNotYourTurn",
	"popejoan.hint.lead":          "popejoan.hintReasonLead",
	"popejoan.hint.follow":        "popejoan.hintReasonFollow",
	"popejoan.hint.none":          "popejoan.hintNone",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *PopeJoanCuiPresenter) ActionLogOutput(c interfaces.PopeJoanGame) string {
	return actionLogOutputText(c)
}
