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

// buraPlayerStr returns the display string for a single Bura player.
// CPU の手札は枚数のみを出す。ここで札を出すと CUI 上で相手の手が丸見えになる。
func buraPlayerStr(player *domain.BuraPlayer, idx, points int) string {
	var b strings.Builder
	b.WriteString(i18n.Tf("bura.playerLine",
		"name", cuiPlayerName(player, idx),
		"points", strconv.Itoa(points),
		"cards", strconv.Itoa(player.GetCardsSize()),
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// BuraCuiPresenter renders the Bura CUI view.
type BuraCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *BuraCuiPresenter) Output(b interfaces.BuraGame, lastErr error) string {
	return buildCuiOutput(i18n.T("bura.helpTitle"), func(sb *strings.Builder) {
		sb.WriteString(i18n.Tf("bura.header",
			"trick", strconv.Itoa(b.GetTrickNumber()),
			"stock", strconv.Itoa(len(b.GetStock()))) + "\n")

		if tc := b.GetTrumpCard(); tc != nil {
			sb.WriteString(i18n.Tf("bura.trumpLine", "card", cuiCardStr(tc)) + "\n")
		} else {
			sb.WriteString(i18n.Tf("bura.trumpLineNone",
				"suit", cuiSuitName(b.GetTrumpSuit())) + "\n")
		}
		sb.WriteString(i18n.Tf("bura.targetLine",
			"target", strconv.Itoa(domain.BuraWinThreshold)) + "\n")
		if line := buraWinningCombosLine(); line != "" {
			sb.WriteString(line + "\n")
		}

		for i, player := range b.GetPlayers() {
			sb.WriteString(buraPlayerStr(player, i, b.GetPlayerPoints(i)))
		}

		sb.WriteString("----------\n")

		if lead := b.GetCurrentLead(); len(lead) > 0 {
			parts := make([]string, 0, len(lead))
			for _, c := range lead {
				parts = append(parts, cuiCardStr(c))
			}
			sb.WriteString(i18n.Tf("bura.leadLine",
				"name", cuiPlayerName(b.GetPlayer(b.GetLeadPlayerIdx()), b.GetLeadPlayerIdx()),
				"cards", strings.Join(parts, " ")) + "\n")
		}

		cuiErrorBlock(sb, lastErr)

		if b.GetGameEndFlag() {
			p0 := b.GetPlayerPoints(0)
			p1 := b.GetPlayerPoints(1)
			var banner string
			switch {
			case b.IsDraw():
				banner = i18n.Tf("bura.gameEndDraw", "p0", strconv.Itoa(p0), "p1", strconv.Itoa(p1))
			case b.GetWinnerIdx() == 0:
				banner = i18n.Tf("bura.gameEndP0", "p0", strconv.Itoa(p0), "p1", strconv.Itoa(p1))
			case b.GetWinnerIdx() < 0:
				banner = i18n.Tf("bura.gameEndDraw", "p0", strconv.Itoa(p0), "p1", strconv.Itoa(p1))
			default:
				banner = i18n.Tf("bura.gameEndP1", "p0", strconv.Itoa(p0), "p1", strconv.Itoa(p1))
			}
			sb.WriteString(color.Green(banner) + "\n")
			return
		}

		currentIdx := b.GetCurrentPlayerIdx()
		if currentIdx < 0 {
			return
		}
		sb.WriteString(i18n.Tf("bura.promptCurrentPlayer",
			"name", cuiPlayerName(b.GetPlayer(currentIdx), currentIdx)) + "\n")
		if len(b.GetCurrentLead()) > 0 {
			sb.WriteString(i18n.Tf("bura.promptRespond",
				"n", strconv.Itoa(len(b.GetCurrentLead()))) + "\n")
		} else {
			sb.WriteString(i18n.T("bura.promptLead") + "\n")
		}
	})
}

// HintOutput emits the current Bura hint.
func (p *BuraCuiPresenter) HintOutput(b interfaces.BuraGame) string {
	indices, reason := buraHint(b)
	key := buraHintReasonKeys[reason]
	if key == "" {
		key = "bura.hintNone"
	}
	if len(indices) == 0 {
		return color.Yellow(i18n.T(key)) + "\n"
	}
	player := b.GetPlayer(0)
	parts := make([]string, 0, len(indices))
	for _, i := range indices {
		parts = append(parts, strconv.Itoa(i)+":"+cuiCardStr(player.GetCard(i)))
	}
	return color.Yellow(i18n.Tf("bura.hintCards",
		"cards", strings.Join(parts, " "),
		"reason", i18n.T(key))) + "\n"
}

// buraHintReasonKeys maps the reason identifiers buraHint returns to their i18n
// keys. The Web presenter ships the identifier itself and the frontend does its
// own lookup; the CUI has to resolve it here.
var buraHintReasonKeys = map[string]string{
	"bura.hint.game_end":      "bura.hintReasonGameEnd",
	"bura.hint.not_your_turn": "bura.hintReasonNotYourTurn",
	"bura.hint.declare":       "bura.hintReasonDeclare",
	"bura.hint.claim":         "bura.hintReasonClaim",
	"bura.hint.respond":       "bura.hintReasonRespond",
	"bura.hint.lead":          "bura.hintReasonLead",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *BuraCuiPresenter) ActionLogOutput(b interfaces.BuraGame) string {
	return actionLogOutputTextForSeats[*domain.BuraPlayer](b)
}

// buraWinningCombosLine はドメイン定義の役一覧から表示行を生成する。
// 役の定義順・一覧は domain.BuraWinningCombinations() を唯一の出所とし、
// 未翻訳キー（i18n.T がキー名をそのまま返した場合）や空キーは黙って落とす。
func buraWinningCombosLine() string {
	combos := domain.BuraWinningCombinations()
	parts := make([]string, 0, len(combos))
	for _, c := range combos {
		key := c.Key()
		if key == "" {
			continue
		}
		fullKey := "bura.combo." + key
		val := i18n.T(fullKey)
		if val == fullKey {
			continue
		}
		parts = append(parts, val)
	}
	if len(parts) == 0 {
		return ""
	}
	return i18n.Tf("bura.combosLine", "combos", strings.Join(parts, " / "))
}
