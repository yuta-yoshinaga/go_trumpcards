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

// SevenTwentySevenCuiPresenter renders the SevenTwentySeven CUI view.
type SevenTwentySevenCuiPresenter struct{}

// i18nSevenTwentySevenResolved はラウンドが解決済み (結果フェーズ) かを返す。
func i18nSevenTwentySevenResolved(g interfaces.SevenTwentySevenGame) bool {
	return g.GetPhase() == domain.SevenTwentySevenPhaseResult
}

// sevenTwentySevenStatusStr は「引く / 止まった / 脱落」の状態ラベルを返す。
func sevenTwentySevenStatusStr(player *domain.SevenTwentySevenPlayer) string {
	switch {
	case player.GetOut():
		return i18n.T("seventwentyseven.statusOut")
	case player.GetStanding():
		return i18n.T("seventwentyseven.statusStanding")
	default:
		return i18n.T("seventwentyseven.statusDrawing")
	}
}

// sevenTwentySevenScoreStr は 7 側 / 27 側の得点を「6.5 / 21」の形で返す。
// 超過した側は「-」。**両方を必ず出す** —— 片側だけでは、いま何を狙えるのかが
// 読めない。
func sevenTwentySevenScoreStr(g interfaces.SevenTwentySevenGame, idx int) string {
	fmtSide := func(side int) string {
		v, ok := g.GetScore(idx, side)
		if !ok {
			return i18n.T("seventwentyseven.bust")
		}
		return domain.SevenTwentySevenFormat(v)
	}
	return fmtSide(domain.SevenTwentySevenSideLow) + " / " + fmtSide(domain.SevenTwentySevenSideHigh)
}

// sevenTwentySevenPlayerStr は 1 プレイヤーの表示文字列を返す。
func sevenTwentySevenPlayerStr(g interfaces.SevenTwentySevenGame, i int) string {
	player := g.GetPlayer(i)
	if player == nil {
		return ""
	}
	reveal := i18nSevenTwentySevenResolved(g)
	showCards := player.GetIsHuman() || (reveal && !player.GetOut())

	var b strings.Builder
	b.WriteString(i18n.Tf("seventwentyseven.playerLine",
		"name", cuiPlayerName(player, i),
		"chips", strconv.Itoa(player.GetChips()),
		"bet", strconv.Itoa(player.GetRoundBet()),
		"status", sevenTwentySevenStatusStr(player),
	))
	b.WriteString("\n")
	if showCards && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "  (" + sevenTwentySevenScoreStr(g, i) + ")\n")
	}
	return b.String()
}

// Output renders the current game state for the active locale.
func (p *SevenTwentySevenCuiPresenter) Output(g interfaces.SevenTwentySevenGame, lastErr error) string {
	return buildCuiOutput(i18n.T("seventwentyseven.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("seventwentyseven.roundLine",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"pot", strconv.Itoa(g.GetPot()),
			"ante", strconv.Itoa(g.GetAnte()),
		) + "\n")

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(sevenTwentySevenPlayerStr(g, i))
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			banner := i18n.Tf("seventwentyseven.gameEnd", "player", strconv.Itoa(g.GetMatchWinnerIdx()))
			b.WriteString(color.Green(banner) + "\n")
			return
		}

		switch g.GetPhase() {
		case domain.SevenTwentySevenPhaseDraw:
			// **狙える 2 つの目標を毎回書く。** 「7 と 27 のどちらに寄せるか」が
			// このゲームそのもので、書いていなければ何を選んでいるのか読めない。
			b.WriteString(i18n.T("seventwentyseven.targetsNote") + "\n")
			b.WriteString(i18n.Tf("seventwentyseven.yourScore",
				"score", sevenTwentySevenScoreStr(g, 0)) + "\n")
			b.WriteString(i18n.T("seventwentyseven.promptDraw") + "\n")
		case domain.SevenTwentySevenPhaseResult:
			b.WriteString(p.resultLine(g))
			b.WriteString(i18n.T("seventwentyseven.promptResult") + "\n")
		}
		b.WriteString(i18n.T("seventwentyseven.promptHelp") + "\n")
	})
}

// resultLine はラウンド結果の 1 行 (色付き) を返す。
func (p *SevenTwentySevenCuiPresenter) resultLine(g interfaces.SevenTwentySevenGame) string {
	low, high := g.GetLowWinner(), g.GetHighWinner()
	if low < 0 && high < 0 {
		return color.Yellow(i18n.Tf("seventwentyseven.result.carry",
			"pot", strconv.Itoa(g.GetCarryPot()),
			"count", strconv.Itoa(g.GetCarryCount()))) + "\n"
	}

	var b strings.Builder
	// **両側の勝者を名指しする。** どちらを取ったのかが分からないと、
	// なぜ半分なのか / なぜ総取りなのかが読めない。
	if low >= 0 && low == high {
		b.WriteString(color.Green(i18n.Tf("seventwentyseven.result.scoop",
			"name", cuiPlayerName(g.GetPlayer(low), low))) + "\n")
	} else {
		if low >= 0 {
			b.WriteString(i18n.Tf("seventwentyseven.result.lowWinner",
				"name", cuiPlayerName(g.GetPlayer(low), low),
				"score", domain.SevenTwentySevenFormat(mustScore(g, low, domain.SevenTwentySevenSideLow))) + "\n")
		} else {
			b.WriteString(i18n.T("seventwentyseven.result.lowEmpty") + "\n")
		}
		if high >= 0 {
			b.WriteString(i18n.Tf("seventwentyseven.result.highWinner",
				"name", cuiPlayerName(g.GetPlayer(high), high),
				"score", domain.SevenTwentySevenFormat(mustScore(g, high, domain.SevenTwentySevenSideHigh))) + "\n")
		} else {
			b.WriteString(i18n.T("seventwentyseven.result.highEmpty") + "\n")
		}
	}

	switch g.GetResult() {
	case domain.SevenTwentySevenResultWin:
		b.WriteString(color.Green(i18n.T("seventwentyseven.result.win")) + "\n")
	case domain.SevenTwentySevenResultLose:
		b.WriteString(color.Red(i18n.T("seventwentyseven.result.lose")) + "\n")
	}
	return b.String()
}

// mustScore は勝者の得点を返す。勝者として選ばれている以上その側で生きている。
func mustScore(g interfaces.SevenTwentySevenGame, idx, side int) int {
	v, _ := g.GetScore(idx, side)
	return v
}

// HintOutput emits the current SevenTwentySeven hint.
func (p *SevenTwentySevenCuiPresenter) HintOutput(g interfaces.SevenTwentySevenGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("seventwentyseven.hintNone") + "\n"
	}
	action := i18n.T("seventwentyseven.stand")
	if hint.Draw {
		action = i18n.T("seventwentyseven.takeCard")
	}
	reason := hintReasonStr(hint.Reason, sevenTwentySevenHintReasonKeys)
	return color.Yellow(i18n.Tf("seventwentyseven.hint", "action", action, "reason", reason)) + "\n"
}

// sevenTwentySevenHintReasonKeys maps hint-reason identifiers to i18n keys.
var sevenTwentySevenHintReasonKeys = map[string]string{
	"chase_seven":         "seventwentyseven.hintReasonChaseSeven",
	"chase_twentyseven":   "seventwentyseven.hintReasonChaseTwentySeven",
	"exactly_seven":       "seventwentyseven.hintReasonExactlySeven",
	"exactly_twentyseven": "seventwentyseven.hintReasonExactlyTwentySeven",
	"stand_pat":           "seventwentyseven.hintReasonStandPat",
	"bust_both":           "seventwentyseven.hintReasonBustBoth",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *SevenTwentySevenCuiPresenter) ActionLogOutput(g interfaces.SevenTwentySevenGame) string {
	return actionLogOutputTextForSeats[*domain.SevenTwentySevenPlayer](g)
}
