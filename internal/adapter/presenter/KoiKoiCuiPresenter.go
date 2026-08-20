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

// koikoiDecisionInfoStr describes the stakes of a koi-koi / shobu decision:
// the points locked in by calling shobu now (pending × the koi-koi multiplier)
// and both players' current cumulative scores, so the CLI player can weigh the
// reward against the risk of being overtaken by continuing.
func koikoiDecisionInfoStr(g interfaces.KoiKoiGame) string {
	// Calling shobu now scores pending × 2 once any koi-koi has been declared
	// this round (mirrors KoiKoi.endRound).
	mult := 1
	if g.GetKoikoiCount() >= 1 {
		mult = 2
	}
	confirmed := g.GetPendingPoints() * mult

	humanIdx, oppIdx := 0, 1
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if g.GetPlayer(i).GetIsHuman() {
			humanIdx = i
			oppIdx = (i + 1) % g.GetPlayerCnt()
			break
		}
	}
	return i18n.Tf("koikoi.promptDecisionInfo",
		"confirmed", strconv.Itoa(confirmed),
		"mult", strconv.Itoa(mult),
		"you", strconv.Itoa(g.GetPlayer(humanIdx).GetScore()),
		"opp", strconv.Itoa(g.GetPlayer(oppIdx).GetScore()))
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
			b.WriteString(koikoiDecisionInfoStr(g) + "\n")
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
	// 人間側だけ i18n を通していて、CPU は英語リテラルのままだった。
	// 名前の組み立ては他の表示と同じ cuiPlayerName に任せる (#4855)。
	name := cuiPlayerName(g.GetPlayer(res.Winner), res.Winner)
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
	return actionLogOutputTextForSeats[*domain.KoiKoiPlayer](g)
}
