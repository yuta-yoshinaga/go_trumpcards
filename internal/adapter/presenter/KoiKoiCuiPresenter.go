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

// KoiKoiCuiPresenter renders the Koi-Koi (こいこい) CUI view (花札; no French suits).
type KoiKoiCuiPresenter struct{}

// koikoiCuiCardStr は花札を "松·光" のように描画する。色は札種に応じて付ける。
func koikoiCuiCardStr(c *domain.Card) string {
	if c == nil {
		return "??"
	}
	label := domain.KoiKoiCardLabel(c)
	switch domain.KoiKoiCardCategory(c) {
	case domain.KoiKoiBright:
		return color.Yellow(label)
	case domain.KoiKoiRibbon:
		if domain.KoiKoiCardRibbonColor(c) == domain.KoiKoiRibbonBlue {
			return label
		}
		return color.Red(label)
	default:
		return label
	}
}

// koikoiCuiFieldStr は場札を "[0]松·光 [1]萩·タネ" 形式で返す。
func koikoiCuiFieldStr(cards []*domain.Card) string {
	if len(cards) == 0 {
		return "-"
	}
	parts := make([]string, len(cards))
	for i, c := range cards {
		parts[i] = "[" + strconv.Itoa(i) + "]" + koikoiCuiCardStr(c)
	}
	return strings.Join(parts, " ")
}

// koikoiCuiHandStr は人間の手札をインデックス付きで描画する。
func koikoiCuiHandStr(p *domain.KoiKoiPlayer) string {
	parts := make([]string, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		parts[i] = "[" + strconv.Itoa(i) + "]" + koikoiCuiCardStr(p.GetCard(i))
	}
	return strings.Join(parts, " ")
}

// koikoiYakuStr は成立役を "光3(五光=10)" 風に描画する。
func koikoiYakuStr(yakus []domain.KoiKoiYaku) string {
	if len(yakus) == 0 {
		return "-"
	}
	parts := make([]string, len(yakus))
	for i, y := range yakus {
		parts[i] = i18n.T("koikoi.yaku."+y.Key) + "(" + strconv.Itoa(y.Points) + ")"
	}
	return strings.Join(parts, ", ")
}

func koikoiPlayerStr(g interfaces.KoiKoiGame, idx int) string {
	player := g.GetPlayer(idx)
	if player == nil {
		return ""
	}
	yakus, pts := g.GetYaku(idx)
	var b strings.Builder
	b.WriteString(i18n.Tf("koikoi.playerLine",
		"name", cuiPlayerName(player, idx),
		"hand", strconv.Itoa(player.GetCardsSize()),
		"captured", strconv.Itoa(player.CapturedCount()),
		"score", strconv.Itoa(player.GetScore()),
		"yaku", koikoiYakuStr(yakus),
		"points", strconv.Itoa(pts)) + "\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(koikoiCuiHandStr(player) + "\n")
	}
	return b.String()
}

// Output renders the current game state for the active locale.
func (p *KoiKoiCuiPresenter) Output(g interfaces.KoiKoiGame, lastErr error) string {
	return buildCuiOutput(i18n.T("koikoi.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("koikoi.roundLine",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"deck", strconv.Itoa(g.GetRemainingDeck()),
			"koikoi", strconv.Itoa(g.GetKoikoiCount())) + "\n")
		b.WriteString(i18n.Tf("koikoi.fieldLine", "field", koikoiCuiFieldStr(g.GetFieldCards())) + "\n")

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(koikoiPlayerStr(g, i))
		}
		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		switch g.GetPhase() {
		case domain.KoiKoiPhasePlay:
			currentIdx := g.GetCurrentTurn()
			b.WriteString(i18n.Tf("koikoi.promptPlay",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
		case domain.KoiKoiPhaseKoiKoiDecision:
			b.WriteString(i18n.Tf("koikoi.promptDecision",
				"yaku", koikoiYakuStr(g.GetPendingYaku()),
				"points", strconv.Itoa(g.GetPendingPoints())) + "\n")
		case domain.KoiKoiPhaseRoundEnd:
			b.WriteString(koikoiRoundResultStr(g) + "\n")
		case domain.KoiKoiPhaseGameEnd:
			b.WriteString(i18n.T("koikoi.promptGameEnd") + "\n")
		}
		b.WriteString(i18n.T("koikoi.promptHelp") + "\n")
	})
}

// koikoiRoundResultStr はラウンド結果の説明文を返す。
func koikoiRoundResultStr(g interfaces.KoiKoiGame) string {
	res := g.GetLastRoundResult()
	if res == nil {
		return i18n.T("koikoi.roundDraw")
	}
	if res.Winner < 0 {
		return i18n.T("koikoi.roundDraw")
	}
	name := "CPU"
	if p := g.GetPlayer(res.Winner); p != nil && p.GetIsHuman() {
		name = i18n.T("cuiPlayerYou")
	}
	return i18n.Tf("koikoi.roundWin",
		"name", name,
		"yaku", koikoiYakuStr(res.Yaku),
		"total", strconv.Itoa(res.Total),
		"mult", strconv.Itoa(res.Multiplier))
}

// HintOutput emits the current Koi-Koi hint.
func (p *KoiKoiCuiPresenter) HintOutput(g interfaces.KoiKoiGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("koikoi.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, koikoiHintReasonKeys)
	if hint.KoiKoi >= 0 {
		action := i18n.T("koikoi.hintShobu")
		if hint.KoiKoi == 1 {
			action = i18n.T("koikoi.hintKoikoi")
		}
		return color.Yellow(i18n.Tf("koikoi.hintDecision", "action", action, "reason", reason)) + "\n"
	}
	card := "-"
	if player := g.GetPlayer(g.GetCurrentTurn()); player != nil &&
		hint.CardIndex >= 0 && hint.CardIndex < player.GetCardsSize() {
		card = "[" + strconv.Itoa(hint.CardIndex) + "]" + koikoiCuiCardStr(player.GetCard(hint.CardIndex))
	}
	return color.Yellow(i18n.Tf("koikoi.hintCard", "card", card, "reason", reason)) + "\n"
}

// koikoiHintReasonKeys maps Koi-Koi-specific hint reasons to i18n keys.
var koikoiHintReasonKeys = map[string]string{
	"capture":        "koikoi.hintReasonCapture",
	"discard_low":    "koikoi.hintReasonDiscard",
	"koikoi_lowyaku": "koikoi.hintReasonKoikoi",
	"stop_secure":    "koikoi.hintReasonStop",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *KoiKoiCuiPresenter) ActionLogOutput(g interfaces.KoiKoiGame) string {
	return actionLogOutputText(g)
}
