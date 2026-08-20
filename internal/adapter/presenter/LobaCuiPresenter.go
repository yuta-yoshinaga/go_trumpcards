//go:build !js || !wasm || extra2

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func lobaCardListStr(cards []*domain.Card, indexed bool) string {
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

// lobaMeldKindName はメルド種別の名前を返す。
func lobaMeldKindName(k domain.LobaMeldKind) string {
	if k == domain.LobaMeldPierna {
		return i18n.T("loba.pierna")
	}
	return i18n.T("loba.escalera")
}

// LobaCuiPresenter renders the Loba CUI view.
type LobaCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *LobaCuiPresenter) Output(c interfaces.LobaGame, lastErr error) string {
	return buildCuiOutput(i18n.T("loba.helpTitle"), func(sb *strings.Builder) {
		sb.WriteString(i18n.Tf("loba.header",
			"round", strconv.Itoa(c.GetRoundNumber()+1),
			"stock", strconv.Itoa(c.GetStockCount()),
			"knockout", strconv.Itoa(domain.LobaKnockOut)) + "\n")
		// ピエルナの「異なる 3 スート」とジョーカーの制限が、この game で最も
		// 間違えやすい 2 点なので毎回出す。
		sb.WriteString(i18n.T("loba.ruleLine") + "\n")

		top := c.GetDiscardTop()
		topStr := "-"
		if top != nil {
			topStr = cuiCardStr(top)
		}
		sb.WriteString(i18n.Tf("loba.discardLine", "card", topStr) + "\n")

		for i, m := range c.GetMelds() {
			if m == nil {
				continue
			}
			sb.WriteString(i18n.Tf("loba.meldLine",
				"idx", strconv.Itoa(i),
				"kind", lobaMeldKindName(m.Kind),
				"owner", strconv.Itoa(m.Owner),
				"cards", lobaCardListStr(m.Cards, false)) + "\n")
		}

		for i, player := range c.GetPlayers() {
			line := i18n.Tf("loba.playerLine",
				"name", cuiPlayerName(player, i),
				"score", strconv.Itoa(c.GetScore(i)),
				"cards", strconv.Itoa(player.GetCardsSize()))
			if c.IsEliminated(i) {
				line += " " + i18n.T("loba.eliminatedMark")
			}
			sb.WriteString(line + "\n")
			if player.GetIsHuman() && player.GetCardsSize() > 0 {
				hand := make([]*domain.Card, 0, player.GetCardsSize())
				for j := range player.GetCardsSize() {
					hand = append(hand, player.GetCard(j))
				}
				sb.WriteString("  " + lobaCardListStr(hand, true) + "\n")
			}
		}

		sb.WriteString("----------\n")
		cuiErrorBlock(sb, lastErr)
		sb.WriteString(p.promptBlock(c))
	})
}

// promptBlock はフェーズごとの案内を返す。
func (p *LobaCuiPresenter) promptBlock(c interfaces.LobaGame) string {
	if c.GetGameEndFlag() {
		key := "loba.gameEndLose"
		if c.GetWinnerIdx() == 0 {
			key = "loba.gameEndWin"
		}
		return color.Green(i18n.T(key)) + "\n"
	}
	switch c.GetPhase() {
	case domain.LobaPhaseDraw:
		return i18n.T("loba.promptDraw") + "\n"
	case domain.LobaPhaseAct:
		return i18n.T("loba.promptAct") + "\n"
	case domain.LobaPhaseRoundEnd:
		var sb strings.Builder
		if w := c.GetRoundWinner(); w >= 0 {
			key := "loba.roundWinner"
			if c.IsRoundClean() {
				key = "loba.roundWinnerClean"
			}
			sb.WriteString(i18n.Tf(key, "name", cuiPlayerName(c.GetPlayer(w), w)) + "\n")
		}
		sb.WriteString(i18n.T("loba.promptNext") + "\n")
		return sb.String()
	case domain.LobaPhaseGameEnd:
		return ""
	}
	return ""
}

// HintOutput emits the current Loba hint.
func (p *LobaCuiPresenter) HintOutput(c interfaces.LobaGame) string {
	hint := lobaHint(c)
	key := lobaHintReasonKeys[hint.Reason]
	if key == "" {
		key = "loba.hintNone"
	}
	switch {
	case hint.DrawStock:
		return color.Yellow(i18n.Tf("loba.hintDraw", "reason", i18n.T(key))) + "\n"
	case len(hint.CardIndices) > 0:
		parts := make([]string, 0, len(hint.CardIndices))
		for _, i := range hint.CardIndices {
			parts = append(parts, strconv.Itoa(i))
		}
		return color.Yellow(i18n.Tf("loba.hintMeld",
			"idxs", strings.Join(parts, ","), "reason", i18n.T(key))) + "\n"
	case hint.CardIndex != nil:
		return color.Yellow(i18n.Tf("loba.hintDiscard",
			"idx", strconv.Itoa(*hint.CardIndex), "reason", i18n.T(key))) + "\n"
	default:
		return color.Yellow(i18n.T(key)) + "\n"
	}
}

// lobaHintReasonKeys maps the reason identifiers lobaHint returns to i18n keys.
// The Web presenter ships the identifier and the frontend resolves it; the CUI
// must resolve it here or it prints the raw key.
var lobaHintReasonKeys = map[string]string{
	"loba.hint.game_end":      "loba.hintReasonGameEnd",
	"loba.hint.not_your_turn": "loba.hintReasonNotYourTurn",
	"loba.hint.draw":          "loba.hintReasonDraw",
	"loba.hint.meld":          "loba.hintReasonMeld",
	"loba.hint.discard":       "loba.hintReasonDiscard",
	"loba.hint.none":          "loba.hintNone",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *LobaCuiPresenter) ActionLogOutput(c interfaces.LobaGame) string {
	return actionLogOutputTextForSeats[*domain.LobaPlayer](c)
}
