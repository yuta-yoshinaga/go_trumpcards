//go:build !js || !wasm || extra4

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// GutsCuiPresenter renders the Guts CUI view.
type GutsCuiPresenter struct{}

// gutsDeclareGuideLine は人間の手札の診断行を返す (診断できなければ空文字)。
// 判定は domain.GutsEvaluateGuide がただ一つの出どころで、Web の
// evaluateGutsGuide とは golden vector で結び付けてある。
func gutsDeclareGuideLine(g interfaces.GutsGame) string {
	// 座席 0 は必ず人間 (gutsNewPlayers)。nil 検査だけ残すのはインタフェース越しに
	// 呼ばれるため。
	human := g.GetPlayer(0)
	if human == nil {
		return ""
	}
	cards := make([]*domain.Card, 0, human.GetCardsSize())
	for i := 0; i < human.GetCardsSize(); i++ {
		cards = append(cards, human.GetCard(i))
	}
	guide := domain.GutsEvaluateGuide(cards)
	if guide == nil {
		return ""
	}
	// 役名は結果表示と同じキーを使う (別に持つと片方だけ直る)。
	hand := i18n.T("guts.hand.highcard")
	if guide.Pair {
		hand = i18n.T("guts.hand.pair")
	}
	return i18n.Tf("guts.declareGuide", "hand", hand, "tier", i18n.T(gutsGuideTierKeys[guide.Tier])) + "\n"
}

// gutsGuideTierKeys maps guide tiers to i18n keys.
var gutsGuideTierKeys = map[string]string{
	domain.GutsGuideTierHigh:   "guts.guideTierHigh",
	domain.GutsGuideTierMedium: "guts.guideTierMedium",
	domain.GutsGuideTierLow:    "guts.guideTierLow",
}

// gutsStatusStr は in/out/脱落の状態ラベルを返す。
func gutsStatusStr(g interfaces.GutsGame, player *domain.GutsPlayer, idx int) string {
	switch {
	case player.GetOut():
		return i18n.T("guts.statusOut")
	case i18nGutsResolved(g) && player.GetIn():
		if idx == g.GetWinnerIdx() {
			return i18n.T("guts.statusWinner")
		}
		if g.IsMatcher(idx) {
			return i18n.T("guts.statusMatched")
		}
		return i18n.T("guts.statusIn")
	case player.GetIn():
		return i18n.T("guts.statusIn")
	default:
		return i18n.T("guts.statusWaiting")
	}
}

// i18nGutsResolved はラウンドが解決済み (結果フェーズ) かを返す。
func i18nGutsResolved(g interfaces.GutsGame) bool {
	return g.GetPhase() == domain.GutsPhaseResult
}

// gutsPlayerStr は 1 プレイヤーの表示文字列を返す。
func gutsPlayerStr(g interfaces.GutsGame, i int) string {
	player := g.GetPlayer(i)
	if player == nil {
		return ""
	}
	reveal := i18nGutsResolved(g)
	showCards := player.GetIsHuman() || (reveal && player.GetIn() && !player.GetOut())

	var b strings.Builder
	b.WriteString(i18n.Tf("guts.playerLine",
		"name", cuiPlayerName(player, i),
		"chips", strconv.Itoa(player.GetChips()),
		"bet", strconv.Itoa(player.GetRoundBet()),
		"status", gutsStatusStr(g, player, i),
	))
	b.WriteString("\n")
	if showCards && player.GetCardsSize() > 0 {
		line := cuiIndexedCardListStr(player)
		if showCards && !player.GetIsHuman() && player.GetCardsSize() == domain.GutsHandSize {
			line += "  (" + i18n.T("guts.hand."+gutsHandName(player)) + ")"
		}
		b.WriteString(line + "\n")
	}
	return b.String()
}

// Output renders the current game state for the active locale.
func (p *GutsCuiPresenter) Output(g interfaces.GutsGame, lastErr error) string {
	return buildCuiOutput(i18n.T("guts.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("guts.roundLine",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"pot", strconv.Itoa(g.GetPot()),
			"ante", strconv.Itoa(g.GetAnte()),
		) + "\n")

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(gutsPlayerStr(g, i))
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			banner := i18n.Tf("guts.gameEnd", "player", strconv.Itoa(g.GetMatchWinnerIdx()))
			b.WriteString(color.Green(banner) + "\n")
			return
		}

		switch g.GetPhase() {
		case domain.GutsPhaseDeclare:
			// Web は宣言中つねに手役と勝ち目の目安を出している。ヒントの
			// オン/オフとは無関係のガイドなので、CUI でも常時出す (#5697)。
			if line := gutsDeclareGuideLine(g); line != "" {
				b.WriteString(line)
			}
			b.WriteString(i18n.T("guts.promptDeclare") + "\n")
		case domain.GutsPhaseResult:
			b.WriteString(p.resultLine(g))
			b.WriteString(i18n.T("guts.promptResult") + "\n")
		}
		b.WriteString(i18n.T("guts.promptHelp") + "\n")
	})
}

// resultLine はラウンド結果の 1 行 (色付き) を返す。
func (p *GutsCuiPresenter) resultLine(g interfaces.GutsGame) string {
	if g.GetWinnerIdx() < 0 {
		return color.Yellow(i18n.Tf("guts.result.carry",
			"pot", strconv.Itoa(g.GetCarryPot()),
			"count", strconv.Itoa(g.GetCarryCount()))) + "\n"
	}
	switch g.GetResult() {
	case domain.GutsResultWin:
		return color.Green(i18n.T("guts.result.win")) + "\n"
	case domain.GutsResultLose:
		return color.Red(i18n.T("guts.result.lose")) + "\n"
	default:
		return color.Yellow(i18n.Tf("guts.result.cpuWin", "player", strconv.Itoa(g.GetWinnerIdx()))) + "\n"
	}
}

// HintOutput emits the current Guts hint.
func (p *GutsCuiPresenter) HintOutput(g interfaces.GutsGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("guts.hintNone") + "\n"
	}
	action := i18n.T("guts.declareOut")
	if hint.Declaration == domain.GutsDeclarationIn {
		action = i18n.T("guts.declareIn")
	}
	reason := hintReasonStr(hint.Reason, gutsHintReasonKeys)
	return color.Yellow(i18n.Tf("guts.hint", "action", action, "reason", reason)) + "\n"
}

// gutsHintReasonKeys maps hint-reason identifiers to i18n keys.
var gutsHintReasonKeys = map[string]string{
	"strong_hand": "guts.hintReasonStrongHand",
	"weak_hand":   "guts.hintReasonWeakHand",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *GutsCuiPresenter) ActionLogOutput(g interfaces.GutsGame) string {
	return actionLogOutputTextForSeats[*domain.GutsPlayer](g)
}
