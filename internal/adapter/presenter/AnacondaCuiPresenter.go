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

// anacondaPassRecipient returns the seat the human passes to — the next still-in
// (not eliminated) participant to the left — or -1 when it can't be determined.
// Mirrors the domain executePass rotation (participants[(k+1)%n]).
func anacondaPassRecipient(g interfaces.AnacondaGame) int {
	humanPos := -1
	var participants []int
	for i := 0; i < g.GetPlayerCnt(); i++ {
		p := g.GetPlayer(i)
		if p == nil || p.GetOut() {
			continue
		}
		if p.GetIsHuman() {
			humanPos = len(participants)
		}
		participants = append(participants, i)
	}
	if humanPos < 0 || len(participants) < 2 {
		return -1
	}
	return participants[(humanPos+1)%len(participants)]
}

// AnacondaCuiPresenter renders the Anaconda CUI view.
type AnacondaCuiPresenter struct{}

// anacondaStatusStr は脱落/フォールド/勝者/手番/待機の状態ラベルを返す。
func anacondaStatusStr(g interfaces.AnacondaGame, player *domain.AnacondaPlayer, idx int) string {
	switch {
	case player.GetOut():
		return i18n.T("anaconda.statusOut")
	case player.GetFolded():
		return i18n.T("anaconda.statusFolded")
	case g.GetPhase() == domain.AnacondaPhaseResult && idx == g.GetWinnerIdx():
		return i18n.T("anaconda.statusWinner")
	case g.GetPhase() == domain.AnacondaPhaseRoll && idx == g.GetCurrentPlayerIdx():
		return i18n.T("anaconda.statusTurn")
	default:
		return i18n.T("anaconda.statusWaiting")
	}
}

// anacondaPlayerStr は 1 プレイヤーの表示文字列を返す。
func anacondaPlayerStr(g interfaces.AnacondaGame, i int) string {
	player := g.GetPlayer(i)
	if player == nil {
		return ""
	}
	var b strings.Builder
	b.WriteString(i18n.Tf("anaconda.playerLine",
		"name", cuiPlayerName(player, i),
		"chips", strconv.Itoa(player.GetChips()),
		"bet", strconv.Itoa(player.GetRoundBet()),
		"status", anacondaStatusStr(g, player, i),
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		line := cuiIndexedCardListStr(player)
		// Once the human's kept hand is fully revealed (final roll / showdown),
		// label its poker category too, matching the CPU rows.
		if g.IsHandFullyRevealed(i) {
			line += "  (" + i18n.T("anaconda.hand."+anacondaHandName(g.GetRevealedCards(i))) + ")"
		}
		b.WriteString(line + "\n")
	} else if revealed := g.GetRevealedCards(i); len(revealed) > 0 {
		line := cuiCardSliceStr(revealed)
		if g.IsHandFullyRevealed(i) {
			line += "  (" + i18n.T("anaconda.hand."+anacondaHandName(revealed)) + ")"
		}
		b.WriteString(line + "\n")
	}
	return b.String()
}

// Output renders the current game state for the active locale.
func (p *AnacondaCuiPresenter) Output(g interfaces.AnacondaGame, lastErr error) string {
	return buildCuiOutput(i18n.T("anaconda.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("anaconda.roundLine",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"pot", strconv.Itoa(g.GetPot()),
			"ante", strconv.Itoa(g.GetAnte()),
		) + "\n")

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(anacondaPlayerStr(g, i))
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			banner := i18n.Tf("anaconda.gameEnd", "player", strconv.Itoa(g.GetMatchWinnerIdx()))
			b.WriteString(color.Green(banner) + "\n")
			return
		}

		switch g.GetPhase() {
		case domain.AnacondaPhasePass:
			b.WriteString(i18n.Tf("anaconda.promptPass", "count", strconv.Itoa(g.GetPassCount())) + "\n")
			// Name who receives the pass (the next still-in player to the left) so
			// the human knows where their discarded cards go.
			if recipient := anacondaPassRecipient(g); recipient >= 0 {
				b.WriteString(i18n.Tf("anaconda.promptPassTo",
					"name", cuiPlayerName(g.GetPlayer(recipient), recipient),
					"count", strconv.Itoa(g.GetPassCount())) + "\n")
			}
		case domain.AnacondaPhaseSet:
			b.WriteString(i18n.T("anaconda.promptKeep") + "\n")
		case domain.AnacondaPhaseRoll:
			b.WriteString(i18n.Tf("anaconda.promptRoll",
				"revealed", strconv.Itoa(g.GetRollIndex()),
				"bet", strconv.Itoa(g.GetCurrentBet()),
			) + "\n")
		case domain.AnacondaPhaseResult:
			b.WriteString(p.resultLine(g))
			b.WriteString(i18n.T("anaconda.promptResult") + "\n")
		}
		b.WriteString(i18n.T("anaconda.promptHelp") + "\n")
	})
}

// resultLine はラウンド結果の 1 行 (色付き) を返す。
func (p *AnacondaCuiPresenter) resultLine(g interfaces.AnacondaGame) string {
	if g.GetWinnerIdx() < 0 {
		return color.Yellow(i18n.T("anaconda.result.none")) + "\n"
	}
	switch g.GetResult() {
	case domain.AnacondaResultWin:
		return color.Green(i18n.T("anaconda.result.win")) + "\n"
	case domain.AnacondaResultLose:
		return color.Red(i18n.T("anaconda.result.lose")) + "\n"
	default:
		return color.Yellow(i18n.Tf("anaconda.result.cpuWin", "player", strconv.Itoa(g.GetWinnerIdx()))) + "\n"
	}
}

// HintOutput emits the current Anaconda hint.
func (p *AnacondaCuiPresenter) HintOutput(g interfaces.AnacondaGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("anaconda.hintNone") + "\n"
	}
	action := i18n.T("anaconda.action." + hint.Action)
	reason := hintReasonStr(hint.Reason, anacondaHintReasonKeys)
	line := i18n.Tf("anaconda.hint", "action", action, "reason", reason)
	// **どの札かまで出す。**ドメインは pass/keep の推奨インデックスを計算して
	// 返しているのに、行は「3枚パス（弱いため）」で止まっていた (#4851)。
	// call/raise/fold では CardIndices は nil なので、この行は付かない。
	if cards := anacondaHintCards(g, hint.CardIndices); cards != "" {
		line += " " + i18n.Tf("anaconda.hintCards", "cards", cards)
	}
	return color.Yellow(line) + "\n"
}

// anacondaHintCards renders the recommended indices as "[0]SPADE 5, [2]HEART 9",
// or "" when the hint carries no indices.
func anacondaHintCards(g interfaces.AnacondaGame, indices []int) string {
	if len(indices) == 0 {
		return ""
	}
	// 席順は固定でないので GetIsHuman で探す (このファイルの他の箇所と同じ)。
	var human *domain.AnacondaPlayer
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if p := g.GetPlayer(i); p != nil && p.GetIsHuman() {
			human = p
			break
		}
	}
	if human == nil {
		return ""
	}
	parts := make([]string, 0, len(indices))
	for _, idx := range indices {
		if idx < 0 || idx >= human.GetCardsSize() {
			continue
		}
		parts = append(parts, "["+strconv.Itoa(idx)+"]"+cuiCardStr(human.GetCard(idx)))
	}
	return strings.Join(parts, ", ")
}

// anacondaHintReasonKeys maps hint-reason identifiers to i18n keys.
var anacondaHintReasonKeys = map[string]string{
	"pass_weakest": "anaconda.hintReasonPassWeakest",
	"keep_best":    "anaconda.hintReasonKeepBest",
	"strong_hand":  "anaconda.hintReasonStrongHand",
	"medium_hand":  "anaconda.hintReasonMediumHand",
	"weak_hand":    "anaconda.hintReasonWeakHand",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *AnacondaCuiPresenter) ActionLogOutput(g interfaces.AnacondaGame) string {
	return actionLogOutputTextForSeats[*domain.AnacondaPlayer](g)
}
