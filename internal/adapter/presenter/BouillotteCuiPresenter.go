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

// BouillotteCuiPresenter renders the Bouillotte CUI view.
type BouillotteCuiPresenter struct{}

// bouillotteResolved はラウンドが解決済み (結果フェーズ) かを返す。
func bouillotteResolved(g interfaces.BouillotteGame) bool {
	return g.GetPhase() == domain.BouillottePhaseResult
}

// bouillotteStatusStr は手番/フォールド/脱落の状態ラベルを返す。
func bouillotteStatusStr(g interfaces.BouillotteGame, player *domain.BouillottePlayer, idx int) string {
	switch {
	case player.GetOut():
		return i18n.T("bouillotte.statusOut")
	case player.GetFolded():
		return i18n.T("bouillotte.statusFolded")
	case bouillotteResolved(g):
		if idx == g.GetWinnerIdx() {
			return i18n.T("bouillotte.statusWinner")
		}
		return i18n.T("bouillotte.statusActive")
	case g.GetCurrentPlayerIdx() == idx:
		return i18n.T("bouillotte.statusTurn")
	default:
		return i18n.T("bouillotte.statusWaiting")
	}
}

// bouillottePlayerStr は 1 プレイヤーの表示文字列を返す。
func bouillottePlayerStr(g interfaces.BouillotteGame, i int) string {
	player := g.GetPlayer(i)
	if player == nil {
		return ""
	}
	reveal := bouillotteResolved(g)
	showCards := player.GetIsHuman() || (reveal && !player.GetFolded() && !player.GetOut())

	var b strings.Builder
	b.WriteString(i18n.Tf("bouillotte.playerLine",
		"name", cuiPlayerName(player, i),
		"chips", strconv.Itoa(player.GetChips()),
		"bet", strconv.Itoa(player.GetRoundBet()),
		"status", bouillotteStatusStr(g, player, i),
	))
	b.WriteString("\n")
	if showCards && player.GetCardsSize() > 0 {
		line := cuiIndexedCardListStr(player)
		if showCards && !player.GetIsHuman() && player.GetCardsSize() == domain.BouillotteHandSize {
			line += "  (" + i18n.T("bouillotte.hand."+bouillotteHandName(player, g.GetRetourne())) + ")"
		}
		b.WriteString(line + "\n")
	}
	return b.String()
}

// Output renders the current game state for the active locale.
func (p *BouillotteCuiPresenter) Output(g interfaces.BouillotteGame, lastErr error) string {
	return buildCuiOutput(i18n.T("bouillotte.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("bouillotte.roundLine",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"pot", strconv.Itoa(g.GetPot()),
			"ante", strconv.Itoa(g.GetAnte()),
		) + "\n")

		if r := g.GetRetourne(); r != nil {
			b.WriteString(i18n.Tf("bouillotte.retourneLine", "card", cuiCardStr(r)) + "\n")
		}

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(bouillottePlayerStr(g, i))
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			winIdx := g.GetMatchWinnerIdx()
			banner := i18n.Tf("bouillotte.gameEnd", "name", cuiPlayerName(g.GetPlayer(winIdx), winIdx))
			b.WriteString(color.Green(banner) + "\n")
			return
		}

		switch g.GetPhase() {
		case domain.BouillottePhaseBetting:
			// コールに要る追加額とレイズ後の到達額まで出す (Primero と同じ式)。
			// applyCall/applyRaise がそれぞれ currentBet-roundBet と
			// currentBet+ante を使うので、案内と実際の支払いがずれない。
			actor := g.GetPlayer(g.GetCurrentPlayerIdx())
			need := 0
			if actor != nil {
				// 既払いが現在のベットを超える席では負にしない (applyCall と同じ扱い)。
				if diff := g.GetCurrentBet() - actor.GetRoundBet(); diff > 0 {
					need = diff
				}
			}
			b.WriteString(i18n.Tf("bouillotte.promptBetting",
				"bet", strconv.Itoa(g.GetCurrentBet()),
				"need", strconv.Itoa(need),
				"raiseTo", strconv.Itoa(g.GetCurrentBet()+g.GetAnte()),
			) + "\n")
		case domain.BouillottePhaseResult:
			b.WriteString(p.resultLine(g))
			b.WriteString(i18n.T("bouillotte.promptResult") + "\n")
		}
		b.WriteString(i18n.T("bouillotte.promptHelp") + "\n")
	})
}

// resultLine はラウンド結果の 1 行 (色付き) を返す。
func (p *BouillotteCuiPresenter) resultLine(g interfaces.BouillotteGame) string {
	if g.GetWinnerIdx() < 0 {
		return color.Yellow(i18n.T("bouillotte.result.none")) + "\n"
	}
	switch g.GetResult() {
	case domain.BouillotteResultWin:
		return color.Green(i18n.T("bouillotte.result.win")) + "\n"
	case domain.BouillotteResultLose:
		return color.Red(i18n.T("bouillotte.result.lose")) + "\n"
	default:
		winIdx := g.GetWinnerIdx()
		return color.Yellow(i18n.Tf("bouillotte.result.cpuWin", "name", cuiPlayerName(g.GetPlayer(winIdx), winIdx))) + "\n"
	}
}

// bouillotteHintActionKeys maps hint-action identifiers to i18n keys.
var bouillotteHintActionKeys = map[string]string{
	"call":  "bouillotte.actionCall",
	"raise": "bouillotte.actionRaise",
	"fold":  "bouillotte.actionFold",
}

// bouillotteHintReasonKeys maps hint-reason identifiers to i18n keys.
var bouillotteHintReasonKeys = map[string]string{
	"strong_hand": "bouillotte.hintReasonStrongHand",
	"medium_hand": "bouillotte.hintReasonMediumHand",
	"weak_hand":   "bouillotte.hintReasonWeakHand",
}

// HintOutput emits the current Bouillotte hint.
func (p *BouillotteCuiPresenter) HintOutput(g interfaces.BouillotteGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("bouillotte.hintNone") + "\n"
	}
	action := hintReasonStr(hint.Action, bouillotteHintActionKeys)
	reason := hintReasonStr(hint.Reason, bouillotteHintReasonKeys)
	return color.Yellow(i18n.Tf("bouillotte.hint", "action", action, "reason", reason)) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *BouillotteCuiPresenter) ActionLogOutput(g interfaces.BouillotteGame) string {
	return actionLogOutputText(g)
}
