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

// GoStopCuiPresenter renders the Go-Stop (고스톱) CUI view (花札; no French suits).
type GoStopCuiPresenter struct{}

// gostopCuiCardStr は花札を "松·光" のように描画する。色は札種に応じて付ける。
func gostopCuiCardStr(c *domain.Card) string {
	if c == nil {
		return "??"
	}
	label := domain.GoStopCardLabel(c)
	switch domain.GoStopCardCategory(c) {
	case domain.GoStopGwang:
		return color.Yellow(label)
	case domain.GoStopTti:
		if domain.GoStopCardRibbonColor(c) == domain.GoStopRibbonBlue {
			return label
		}
		return color.Red(label)
	default:
		return label
	}
}

// gostopCuiFieldStr は場札を "[0]松·光 [1]萩·열" 形式で返す。
func gostopCuiFieldStr(cards []*domain.Card) string {
	if len(cards) == 0 {
		return "-"
	}
	parts := make([]string, len(cards))
	for i, c := range cards {
		parts[i] = "[" + strconv.Itoa(i) + "]" + gostopCuiCardStr(c)
	}
	return strings.Join(parts, " ")
}

// gostopCuiHandStr は人間の手札をインデックス付きで描画する。
func gostopCuiHandStr(p *domain.GoStopPlayer) string {
	parts := make([]string, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		parts[i] = "[" + strconv.Itoa(i) + "]" + gostopCuiCardStr(p.GetCard(i))
	}
	return strings.Join(parts, " ")
}

// gostopScoreStr は得点内訳を "光:3 고도리:5 (合計8)" 風に描画する。
func gostopScoreStr(bd *domain.GoStopBreakdown) string {
	if bd == nil || bd.Base == 0 {
		return "-"
	}
	var parts []string
	add := func(key string, pts int) {
		if pts > 0 {
			parts = append(parts, i18n.T("gostop.cat."+key)+":"+strconv.Itoa(pts))
		}
	}
	add("gwang", bd.Gwang)
	add("godori", bd.Godori)
	add("tti", bd.Tti)
	add("yeol", bd.Yeol)
	add("pi", bd.Pi)
	return strings.Join(parts, " ") + " (" + strconv.Itoa(bd.Base) + ")"
}

// gostopCategoryCountStr counts captured cards by card category (光/열끗/띠/피),
// letting CLI players gauge role progress before a category actually scores.
// Captured piles are public in Go-Stop, so this is shown for every player.
func gostopCategoryCountStr(captured []*domain.Card) string {
	var gwang, yeol, tti, pi int
	for _, c := range captured {
		switch domain.GoStopCardCategory(c) {
		case domain.GoStopGwang:
			gwang++
		case domain.GoStopYeol:
			yeol++
		case domain.GoStopTti:
			tti++
		case domain.GoStopPi:
			pi++
		}
	}
	return i18n.Tf("gostop.categoryCounts",
		"gwang", strconv.Itoa(gwang),
		"yeol", strconv.Itoa(yeol),
		"tti", strconv.Itoa(tti),
		"pi", strconv.Itoa(pi))
}

func gostopPlayerStr(g interfaces.GoStopGame, idx int) string {
	player := g.GetPlayer(idx)
	if player == nil {
		return ""
	}
	bd, _ := g.GetScore(idx)
	var b strings.Builder
	b.WriteString(i18n.Tf("gostop.playerLine",
		"name", cuiPlayerName(player, idx),
		"hand", strconv.Itoa(player.GetCardsSize()),
		"captured", strconv.Itoa(player.CapturedCount()),
		"score", strconv.Itoa(player.GetScore()),
		"go", strconv.Itoa(player.GetGoCount()),
		"breakdown", gostopScoreStr(bd)) + "\n")
	if player.CapturedCount() > 0 {
		b.WriteString(gostopCategoryCountStr(player.GetCapturedCards()) + "\n")
	}
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(gostopCuiHandStr(player) + "\n")
	}
	return b.String()
}

// Output renders the current game state for the active locale.
func (p *GoStopCuiPresenter) Output(g interfaces.GoStopGame, lastErr error) string {
	return buildCuiOutput(i18n.T("gostop.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("gostop.roundLine",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"deck", strconv.Itoa(g.GetRemainingDeck())) + "\n")
		b.WriteString(i18n.Tf("gostop.fieldLine", "field", gostopCuiFieldStr(g.GetFieldCards())) + "\n")

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(gostopPlayerStr(g, i))
		}
		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		switch g.GetPhase() {
		case domain.GoStopPhasePlay:
			currentIdx := g.GetCurrentTurn()
			b.WriteString(i18n.Tf("gostop.promptPlay",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
		case domain.GoStopPhaseGoDecision:
			b.WriteString(i18n.Tf("gostop.promptDecision",
				"points", strconv.Itoa(g.GetPendingPoints())) + "\n")
			// 続けるかどうかは「あと何枚でどの役か」で決まる (Web の
			// gostop-yaku-preview と同じ判定を domain から取る)。
			if line := gostopNearYakuLine(g.GetPendingBreakdown()); line != "" {
				b.WriteString(line)
			}
		case domain.GoStopPhaseRoundEnd:
			b.WriteString(gostopRoundResultStr(g) + "\n")
		case domain.GoStopPhaseGameEnd:
			b.WriteString(i18n.T("gostop.promptGameEnd") + "\n")
		}
		b.WriteString(i18n.T("gostop.promptHelp") + "\n")
	})
}

// gostopRoundResultStr はラウンド結果の説明文を返す。
func gostopRoundResultStr(g interfaces.GoStopGame) string {
	res := g.GetLastRoundResult()
	if res == nil || res.Winner < 0 {
		return i18n.T("gostop.roundDraw")
	}
	// 人間側だけ i18n を通していて、CPU は英語リテラルのままだった。
	// 名前の組み立ては他の表示と同じ cuiPlayerName に任せる (#4855)。
	name := cuiPlayerName(g.GetPlayer(res.Winner), res.Winner)
	bak := gostopBakStr(res)
	return i18n.Tf("gostop.roundWin",
		"name", name,
		"score", gostopScoreStr(res.Breakdown),
		"total", strconv.Itoa(res.Total),
		"bak", strconv.Itoa(res.BakMult),
		"bakinfo", bak)
}

// gostopBakStr は成立したバクを列挙する。
func gostopBakStr(res *domain.GoStopRoundResult) string {
	var parts []string
	if res.GwangBak {
		parts = append(parts, i18n.T("gostop.bak.gwang"))
	}
	if res.PiBak {
		parts = append(parts, i18n.T("gostop.bak.pi"))
	}
	if res.GoBak {
		parts = append(parts, i18n.T("gostop.bak.go"))
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, ",")
}

// gostopNearYakuLine renders the "almost scoring" preview (empty when nothing is close).
// 役名のキーは Web の preview.<target> と同じ識別子を domain から受け取る。
func gostopNearYakuLine(bd *domain.GoStopBreakdown) string {
	near := domain.GoStopComputeNearYaku(bd)
	if len(near) == 0 {
		return ""
	}
	items := make([]string, 0, len(near))
	for _, y := range near {
		items = append(items, i18n.Tf("gostop.previewItem",
			"name", i18n.T("gostop.preview"+strings.ToUpper(y.Target[:1])+y.Target[1:]),
			"remaining", strconv.Itoa(y.Remaining)))
	}
	return i18n.Tf("gostop.previewTitle", "items", strings.Join(items, ", ")) + "\n"
}

// HintOutput emits the current Go-Stop hint.
func (p *GoStopCuiPresenter) HintOutput(g interfaces.GoStopGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("gostop.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, gostopHintReasonKeys)
	if hint.Go >= 0 {
		action := i18n.T("gostop.hintStop")
		if hint.Go == 1 {
			action = i18n.T("gostop.hintGo")
		}
		return color.Yellow(i18n.Tf("gostop.hintDecision", "action", action, "reason", reason)) + "\n"
	}
	card := "-"
	if player := g.GetPlayer(g.GetCurrentTurn()); player != nil &&
		hint.CardIndex >= 0 && hint.CardIndex < player.GetCardsSize() {
		card = "[" + strconv.Itoa(hint.CardIndex) + "]" + gostopCuiCardStr(player.GetCard(hint.CardIndex))
	}
	return color.Yellow(i18n.Tf("gostop.hintCard", "card", card, "reason", reason)) + "\n"
}

// gostopHintReasonKeys maps Go-Stop-specific hint reasons to i18n keys.
var gostopHintReasonKeys = map[string]string{
	"capture":     "gostop.hintReasonCapture",
	"discard_low": "gostop.hintReasonDiscard",
	"go_lowscore": "gostop.hintReasonGo",
	"stop_secure": "gostop.hintReasonStop",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *GoStopCuiPresenter) ActionLogOutput(g interfaces.GoStopGame) string {
	return actionLogOutputTextForSeats[*domain.GoStopPlayer](g)
}
