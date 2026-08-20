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

func nainJauneCardListStr(cards []*domain.Card, indexed bool) string {
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

// nainJauneBoxLabel は区画の表示名を返す。
func nainJauneBoxLabel(b domain.NainJauneBox) string {
	return i18n.T("nainjaune.box." + b.String())
}

// NainJauneCuiPresenter renders the Le Nain Jaune CUI view.
type NainJauneCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *NainJauneCuiPresenter) Output(c interfaces.NainJauneGame, lastErr error) string {
	return buildCuiOutput(i18n.T("nainjaune.helpTitle"), func(sb *strings.Builder) {
		sb.WriteString(i18n.Tf("nainjaune.header",
			"deal", strconv.Itoa(c.GetDealNumber()+1),
			"target", strconv.Itoa(c.GetConfig().TargetDeals),
			"talon", strconv.Itoa(c.GetTalonCount())) + "\n")
		// 「スートは無関係」と「支払いは枚数ではなく点数」が最も誤解されやすい
		// ので毎回出す。
		sb.WriteString(i18n.T("nainjaune.ruleLine") + "\n")

		board := c.GetBoard()
		boxes := make([]string, 0, domain.NainJauneBoxCount)
		for i := range domain.NainJauneBoxCount {
			box := domain.NainJauneBox(i)
			boxes = append(boxes, nainJauneBoxLabel(box)+":"+strconv.Itoa(board.Get(box)))
		}
		sb.WriteString(i18n.Tf("nainjaune.boardLine", "boxes", strings.Join(boxes, " ")) + "\n")

		for _, a := range c.GetAwards() {
			if a == nil {
				continue
			}
			sb.WriteString(i18n.Tf("nainjaune.awardLine",
				"name", cuiPlayerName(c.GetPlayer(a.Player), a.Player),
				"box", nainJauneBoxLabel(a.Box),
				"chips", strconv.Itoa(a.Chips)) + "\n")
		}

		if pile := c.GetPlayedPile(); len(pile) > 0 {
			sb.WriteString(i18n.Tf("nainjaune.pileLine", "cards", nainJauneCardListStr(pile, false)) + "\n")
		}

		for i, player := range c.GetPlayers() {
			if player == nil {
				continue
			}
			points := 0
			for j := range player.GetCardsSize() {
				points += domain.NainJaunePoints(player.GetCard(j))
			}
			sb.WriteString(i18n.Tf("nainjaune.playerLine",
				"name", cuiPlayerName(player, i),
				"chips", strconv.Itoa(player.GetChips()),
				"cards", strconv.Itoa(player.GetCardsSize()),
				"points", strconv.Itoa(points)) + "\n")
			if player.GetIsHuman() && player.GetCardsSize() > 0 {
				hand := make([]*domain.Card, 0, player.GetCardsSize())
				for j := range player.GetCardsSize() {
					hand = append(hand, player.GetCard(j))
				}
				sb.WriteString("  " + nainJauneCardListStr(hand, true) + "\n")
			}
		}

		sb.WriteString("----------\n")
		cuiErrorBlock(sb, lastErr)
		sb.WriteString(p.promptBlock(c))
	})
}

// promptBlock はフェーズごとの案内を返す。
func (p *NainJauneCuiPresenter) promptBlock(c interfaces.NainJauneGame) string {
	if c.GetGameEndFlag() {
		key := "nainjaune.gameEndLose"
		if c.GetWinnerIdx() == 0 {
			key = "nainjaune.gameEndWin"
		}
		return color.Green(i18n.T(key)) + "\n"
	}
	switch c.GetPhase() {
	case domain.NainJaunePhasePlay:
		// **並びが止まっているかどうかで出せる札がまるで違う。**
		if c.GetRunRank() == 0 {
			return i18n.T("nainjaune.promptLead") + "\n"
		}
		return i18n.Tf("nainjaune.promptFollow", "rank", strconv.Itoa(c.GetRunRank()+1)) + "\n"
	case domain.NainJaunePhaseDealEnd:
		var sb strings.Builder
		if w := c.GetDealWinner(); w >= 0 {
			sb.WriteString(i18n.Tf("nainjaune.dealResult",
				"name", cuiPlayerName(c.GetPlayer(w), w)) + "\n")
		}
		sb.WriteString(i18n.T("nainjaune.promptNext") + "\n")
		return sb.String()
	case domain.NainJaunePhaseGameEnd:
		return ""
	}
	return ""
}

// HintOutput emits the current Le Nain Jaune hint.
func (p *NainJauneCuiPresenter) HintOutput(c interfaces.NainJauneGame) string {
	hint := nainJauneHint(c)
	key := nainJauneHintReasonKeys[hint.Reason]
	if key == "" {
		key = "nainjaune.hintNone"
	}
	if hint.CardIndex != nil {
		return color.Yellow(i18n.Tf("nainjaune.hintPlay",
			"idx", strconv.Itoa(*hint.CardIndex), "reason", i18n.T(key))) + "\n"
	}
	return color.Yellow(i18n.T(key)) + "\n"
}

// nainJauneHintReasonKeys maps the reason identifiers nainJauneHint returns to
// i18n keys. The Web presenter ships the identifier and the frontend resolves
// it; the CUI must resolve it here or it prints the raw key.
var nainJauneHintReasonKeys = map[string]string{
	"nainjaune.hint.game_end":      "nainjaune.hintReasonGameEnd",
	"nainjaune.hint.deal_end":      "nainjaune.hintReasonDealEnd",
	"nainjaune.hint.not_your_turn": "nainjaune.hintReasonNotYourTurn",
	"nainjaune.hint.lead":          "nainjaune.hintReasonLead",
	"nainjaune.hint.follow":        "nainjaune.hintReasonFollow",
	"nainjaune.hint.box":           "nainjaune.hintReasonBox",
	"nainjaune.hint.none":          "nainjaune.hintNone",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *NainJauneCuiPresenter) ActionLogOutput(c interfaces.NainJauneGame) string {
	return actionLogOutputTextForSeats[*domain.NainJaunePlayer](c)
}
