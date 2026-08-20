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

func desmocheCardListStr(cards []*domain.Card, indexed bool) string {
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

// desmocheMeldKindName はメルド種別の名前を返す。
func desmocheMeldKindName(k domain.DesmocheMeldKind) string {
	if k == domain.DesmocheMeldSet {
		return i18n.T("desmoche.set")
	}
	return i18n.T("desmoche.run")
}

// DesmocheCuiPresenter renders the Desmoche CUI view.
type DesmocheCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *DesmocheCuiPresenter) Output(c interfaces.DesmocheGame, lastErr error) string {
	return buildCuiOutput(i18n.T("desmoche.helpTitle"), func(sb *strings.Builder) {
		sb.WriteString(i18n.Tf("desmoche.header",
			"round", strconv.Itoa(c.GetRoundNumber()+1),
			"stock", strconv.Itoa(c.GetStockCount()),
			"pot", strconv.Itoa(c.GetPot())) + "\n")
		// 上がりが 9 枚ではなく 10 枚であること、ポーカーの役は使わないことが
		// この game で最も誤解されやすいので毎回出す。
		sb.WriteString(i18n.T("desmoche.ruleLine") + "\n")

		top := c.GetDiscardTop()
		topStr := "-"
		if top != nil {
			topStr = cuiCardStr(top)
		}
		sb.WriteString(i18n.Tf("desmoche.discardLine", "card", topStr) + "\n")

		for i, m := range c.GetMelds() {
			if m == nil {
				continue
			}
			sb.WriteString(i18n.Tf("desmoche.meldLine",
				"idx", strconv.Itoa(i),
				"kind", desmocheMeldKindName(m.Kind),
				"owner", strconv.Itoa(m.Owner),
				"cards", desmocheCardListStr(m.Cards, false)) + "\n")
		}

		for i, player := range c.GetPlayers() {
			if player == nil {
				continue
			}
			sb.WriteString(i18n.Tf("desmoche.playerLine",
				"name", cuiPlayerName(player, i),
				"score", strconv.Itoa(c.GetScore(i)),
				"melded", strconv.Itoa(c.MeldedCount(i)),
				"goout", strconv.Itoa(domain.DesmocheGoOutSize),
				"cards", strconv.Itoa(player.GetCardsSize())) + "\n")
			if player.GetIsHuman() && player.GetCardsSize() > 0 {
				hand := make([]*domain.Card, 0, player.GetCardsSize())
				for j := range player.GetCardsSize() {
					hand = append(hand, player.GetCard(j))
				}
				sb.WriteString("  " + desmocheCardListStr(hand, true) + "\n")
			}
		}

		sb.WriteString("----------\n")
		cuiErrorBlock(sb, lastErr)
		sb.WriteString(p.promptBlock(c))
	})
}

// promptBlock はフェーズごとの案内を返す。
func (p *DesmocheCuiPresenter) promptBlock(c interfaces.DesmocheGame) string {
	if c.GetGameEndFlag() {
		key := "desmoche.gameEndLose"
		if c.GetWinnerIdx() == 0 {
			key = "desmoche.gameEndWin"
		}
		return color.Green(i18n.T(key)) + "\n"
	}
	switch c.GetPhase() {
	case domain.DesmochePhaseDraw:
		return i18n.T("desmoche.promptDraw") + "\n"
	case domain.DesmochePhaseAct:
		// **10 枚上がりが勝利条件なのに、他家のメルドへ付けても自分の枚数は
		// 増えない。** Web は foreignMeldWarning で警告しているが、CUI には
		// 対応する文言が無かった (#5720)。MeldedCount は Owner == player の
		// メルドしか数えない。
		return i18n.T("desmoche.promptAct") + "\n" +
			i18n.T("desmoche.promptActLayoffNote") + "\n"
	case domain.DesmochePhaseRoundEnd:
		var sb strings.Builder
		if w := c.GetRoundWinner(); w >= 0 {
			sb.WriteString(i18n.Tf("desmoche.roundWinner",
				"name", cuiPlayerName(c.GetPlayer(w), w)) + "\n")
		} else {
			// **勝者なしはこの game 固有の結末。**ポットが残ることを明示する。
			sb.WriteString(i18n.Tf("desmoche.roundNoWinner", "pot", strconv.Itoa(c.GetPot())) + "\n")
		}
		sb.WriteString(i18n.T("desmoche.promptNext") + "\n")
		return sb.String()
	case domain.DesmochePhaseGameEnd:
		return ""
	}
	return ""
}

// HintOutput emits the current Desmoche hint.
func (p *DesmocheCuiPresenter) HintOutput(c interfaces.DesmocheGame) string {
	hint := desmocheHint(c)
	key := desmocheHintReasonKeys[hint.Reason]
	if key == "" {
		key = "desmoche.hintNone"
	}
	switch {
	case hint.DrawStock:
		return color.Yellow(i18n.Tf("desmoche.hintDraw", "reason", i18n.T(key))) + "\n"
	case len(hint.CardIndices) > 0:
		parts := make([]string, 0, len(hint.CardIndices))
		for _, i := range hint.CardIndices {
			parts = append(parts, strconv.Itoa(i))
		}
		return color.Yellow(i18n.Tf("desmoche.hintMeld",
			"idxs", strings.Join(parts, ","), "reason", i18n.T(key))) + "\n"
	case hint.CardIndex != nil:
		return color.Yellow(i18n.Tf("desmoche.hintDiscard",
			"idx", strconv.Itoa(*hint.CardIndex), "reason", i18n.T(key))) + "\n"
	default:
		return color.Yellow(i18n.T(key)) + "\n"
	}
}

// desmocheHintReasonKeys maps the reason identifiers desmocheHint returns to
// i18n keys. The Web presenter ships the identifier and the frontend resolves
// it; the CUI must resolve it here or it prints the raw key.
var desmocheHintReasonKeys = map[string]string{
	"desmoche.hint.game_end":      "desmoche.hintReasonGameEnd",
	"desmoche.hint.not_your_turn": "desmoche.hintReasonNotYourTurn",
	"desmoche.hint.draw":          "desmoche.hintReasonDraw",
	"desmoche.hint.meld":          "desmoche.hintReasonMeld",
	"desmoche.hint.discard":       "desmoche.hintReasonDiscard",
	"desmoche.hint.none":          "desmoche.hintNone",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *DesmocheCuiPresenter) ActionLogOutput(c interfaces.DesmocheGame) string {
	return actionLogOutputTextForSeats[*domain.DesmochePlayer](c)
}
