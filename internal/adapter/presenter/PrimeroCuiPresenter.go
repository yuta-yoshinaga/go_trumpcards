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

// PrimeroCuiPresenter renders the Primero CUI view.
type PrimeroCuiPresenter struct{}

// primeroResolved はラウンドが解決済み (結果フェーズ) かを返す。
func primeroResolved(g interfaces.PrimeroGame) bool {
	return g.GetPhase() == domain.PrimeroPhaseResult
}

// primeroStatusStr は手番/フォールド/脱落の状態ラベルを返す。
func primeroStatusStr(g interfaces.PrimeroGame, player *domain.PrimeroPlayer, idx int) string {
	switch {
	case player.GetOut():
		return i18n.T("primero.statusOut")
	case player.GetFolded():
		return i18n.T("primero.statusFolded")
	case primeroResolved(g):
		if idx == g.GetWinnerIdx() {
			return i18n.T("primero.statusWinner")
		}
		return i18n.T("primero.statusActive")
	case g.GetCurrentPlayerIdx() == idx:
		return i18n.T("primero.statusTurn")
	default:
		return i18n.T("primero.statusWaiting")
	}
}

// primeroPlayerStr は 1 プレイヤーの表示文字列を返す。
func primeroPlayerStr(g interfaces.PrimeroGame, i int) string {
	player := g.GetPlayer(i)
	if player == nil {
		return ""
	}
	reveal := primeroResolved(g)
	showCards := player.GetIsHuman() || (reveal && !player.GetFolded() && !player.GetOut())

	var b strings.Builder
	b.WriteString(i18n.Tf("primero.playerLine",
		"name", cuiPlayerName(player, i),
		"chips", strconv.Itoa(player.GetChips()),
		"bet", strconv.Itoa(player.GetRoundBet()),
		"status", primeroStatusStr(g, player, i),
	))
	b.WriteString("\n")
	if showCards && player.GetCardsSize() > 0 {
		line := cuiIndexedCardListStr(player)
		if showCards && !player.GetIsHuman() && player.GetCardsSize() == domain.PrimeroHandSize {
			line += "  (" + i18n.T("primero.hand."+primeroHandName(player)) + ")"
		}
		b.WriteString(line + "\n")
	}
	return b.String()
}

// Output renders the current game state for the active locale.
func (p *PrimeroCuiPresenter) Output(g interfaces.PrimeroGame, lastErr error) string {
	return buildCuiOutput(i18n.T("primero.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("primero.roundLine",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"pot", strconv.Itoa(g.GetPot()),
			"ante", strconv.Itoa(g.GetAnte()),
		) + "\n")

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(primeroPlayerStr(g, i))
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			banner := i18n.Tf("primero.gameEnd", "player", strconv.Itoa(g.GetMatchWinnerIdx()))
			b.WriteString(color.Green(banner) + "\n")
			return
		}

		switch g.GetPhase() {
		case domain.PrimeroPhaseBetting:
			// Spell out what the acting player must add to call and the total a
			// raise reaches, so the human isn't left computing it by hand. The
			// acting seat is always valid here and never bets past the current
			// level, so need = currentBet - roundBet stays non-negative.
			actor := g.GetPlayer(g.GetCurrentPlayerIdx())
			b.WriteString(i18n.Tf("primero.promptBetting",
				"bet", strconv.Itoa(g.GetCurrentBet()),
				"need", strconv.Itoa(g.GetCurrentBet()-actor.GetRoundBet()),
				"raiseTo", strconv.Itoa(g.GetCurrentBet()+g.GetAnte()),
			) + "\n")
		case domain.PrimeroPhaseResult:
			b.WriteString(p.resultLine(g))
			b.WriteString(i18n.T("primero.promptResult") + "\n")
		}
		// fluxus / supremus は一般的なポーカー用語ではないので、強い順と条件を
		// 1 行で添える (Web の常設 legend と同じ内容)。並び順は domain の
		// PrimeroHand* 定数の降順で、その一致は infrastructure のガードが見る。
		b.WriteString(i18n.Tf("primero.handRanking",
			"fluxus", i18n.T("primero.hand."+primeroCategoryLabel(domain.PrimeroHandFluxus)),
			"supremus", i18n.T("primero.hand."+primeroCategoryLabel(domain.PrimeroHandSupremus)),
			"primero", i18n.T("primero.hand."+primeroCategoryLabel(domain.PrimeroHandPrimero)),
			"numerus", i18n.T("primero.hand."+primeroCategoryLabel(domain.PrimeroHandNumerus))) + "\n")
		b.WriteString(i18n.T("primero.promptHelp") + "\n")
	})
}

// resultLine はラウンド結果の 1 行 (色付き) を返す。
func (p *PrimeroCuiPresenter) resultLine(g interfaces.PrimeroGame) string {
	if g.GetWinnerIdx() < 0 {
		return color.Yellow(i18n.T("primero.result.none")) + "\n"
	}
	switch g.GetResult() {
	case domain.PrimeroResultWin:
		return color.Green(i18n.T("primero.result.win")) + "\n"
	case domain.PrimeroResultLose:
		return color.Red(i18n.T("primero.result.lose")) + "\n"
	default:
		return color.Yellow(i18n.Tf("primero.result.cpuWin", "player", strconv.Itoa(g.GetWinnerIdx()))) + "\n"
	}
}

// primeroHintActionKeys maps hint-action identifiers to i18n keys.
var primeroHintActionKeys = map[string]string{
	"call":  "primero.actionCall",
	"raise": "primero.actionRaise",
	"fold":  "primero.actionFold",
}

// primeroHintReasonKeys maps hint-reason identifiers to i18n keys.
var primeroHintReasonKeys = map[string]string{
	"strong_hand": "primero.hintReasonStrongHand",
	"medium_hand": "primero.hintReasonMediumHand",
	"weak_hand":   "primero.hintReasonWeakHand",
}

// HintOutput emits the current Primero hint.
func (p *PrimeroCuiPresenter) HintOutput(g interfaces.PrimeroGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("primero.hintNone") + "\n"
	}
	action := hintReasonStr(hint.Action, primeroHintActionKeys)
	reason := hintReasonStr(hint.Reason, primeroHintReasonKeys)
	return color.Yellow(i18n.Tf("primero.hint", "action", action, "reason", reason)) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *PrimeroCuiPresenter) ActionLogOutput(g interfaces.PrimeroGame) string {
	return actionLogOutputText(g)
}
