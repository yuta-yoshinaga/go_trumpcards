//go:build !js || !wasm || casino

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// BlackJackSwitchCuiPresenter ブラックジャック・スイッチCUIプレゼンター
type BlackJackSwitchCuiPresenter struct{}

// Output ゲーム状態を出力
func (bp *BlackJackSwitchCuiPresenter) Output(g interfaces.BlackJackSwitchGame, lastErr error) string {
	return buildCuiOutput(i18n.T("blackjackswitch.outputTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("blackjackswitch.chipsLine", "chips", strconv.Itoa(g.GetPlayer().GetChips())) + "\n")
		b.WriteString(i18n.Tf("blackjackswitch.phaseLine", "phase", bp.phaseStr(g.GetPhase())) + "\n")

		bp.writeDealer(b, g)
		bp.writeHands(b, g)
		bp.writeSwitchPreview(b, g)

		if g.IsSwitched() {
			b.WriteString(i18n.T("blackjackswitch.switchedLine") + "\n")
		}

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			bp.writeEndSummary(b, g)
		}
	})
}

// ActionLogOutput 棋譜出力
func (bp *BlackJackSwitchCuiPresenter) ActionLogOutput(g interfaces.BlackJackSwitchGame) string {
	return actionLogOutputText(g)
}

func (bp *BlackJackSwitchCuiPresenter) writeDealer(b *strings.Builder, g interfaces.BlackJackSwitchGame) {
	dealer := g.GetDealer()
	if dealer == nil || dealer.GetCardsSize() == 0 {
		return
	}
	if !g.GetGameEndFlag() && dealer.GetCardsSize() >= 2 {
		// アクション中はホールカードを伏せる: 1枚目だけ表示。
		first := dealer.GetCard(0)
		visibleScore := domain.CalculateBlackJackScore([]*domain.Card{first})
		b.WriteString(i18n.Tf("blackjackswitch.dealerLine",
			"score", strconv.Itoa(visibleScore)+"+?",
			"cards", cuiCardStr(first)+",??",
		) + "\n")
		return
	}
	b.WriteString(i18n.Tf("blackjackswitch.dealerLine",
		"score", strconv.Itoa(dealer.GetScore()),
		"cards", cuiCardListStr(dealer),
	) + "\n")
}

func (bp *BlackJackSwitchCuiPresenter) writeHands(b *strings.Builder, g interfaces.BlackJackSwitchGame) {
	hands := g.GetHands()
	curIdx := g.GetCurrentHandIdx()
	for i, h := range hands {
		marker := ""
		if g.GetPhase() == domain.BJSwitchPhaseAction && i == curIdx {
			marker = "> "
		}
		line := i18n.Tf("blackjackswitch.handLine",
			"idx", strconv.Itoa(i),
			"score", strconv.Itoa(h.GetScore()),
			"cards", cuiCardListStr(h),
			"bet", strconv.Itoa(h.GetBet()),
		)
		b.WriteString(marker + line + "\n")
	}
}

// writeSwitchPreview は 2 枚目を入れ替えたときの両ハンドの得点を先に見せる。
//
// **打つまで得か損か分からなかった。**Web はホバーで先読みを出しているのに、
// CUI は `switch` を実行して結果を見るまで比べられなかった (#5586)。
// スイッチフェーズ以外では出さない ── 入れ替えられない局面で得点だけ出しても
// 意味が無く、確定した得点と読み違える。
func (bp *BlackJackSwitchCuiPresenter) writeSwitchPreview(b *strings.Builder, g interfaces.BlackJackSwitchGame) {
	if g.GetPhase() != domain.BJSwitchPhaseSwitch {
		return
	}
	first, second, ok := g.SwitchPreviewScores()
	if !ok {
		return
	}
	b.WriteString(color.Yellow(i18n.Tf("blackjackswitch.switchPreviewLine",
		"first", bjSwitchPreviewScoreStr(first),
		"second", bjSwitchPreviewScoreStr(second))) + "\n")
}

// bjSwitchPreviewScoreStr は得点を出す。21 を超えるならバストと分かる形にする ──
// 数字だけでは、良くなったのか壊れたのかが読み取りにくい。
func bjSwitchPreviewScoreStr(score int) string {
	if score > domain.BlackJackBustOver {
		return i18n.Tf("blackjackswitch.switchPreviewBust", "score", strconv.Itoa(score))
	}
	return strconv.Itoa(score)
}

func (bp *BlackJackSwitchCuiPresenter) writeEndSummary(b *strings.Builder, g interfaces.BlackJackSwitchGame) {
	if g.IsDealerPushed22() {
		b.WriteString(color.Yellow(i18n.T("blackjackswitch.dealer22Push")) + "\n")
	}
	results := g.GetHandResults()
	payouts := g.GetHandPayouts()
	for i := range results {
		line := i18n.Tf("blackjackswitch.handResultLine",
			"idx", strconv.Itoa(i),
			"result", bp.resultStr(results[i]),
			"payout", strconv.Itoa(payouts[i]),
		)
		switch results[i] {
		case domain.GameResultWin:
			b.WriteString(color.Green(line) + "\n")
		case domain.GameResultLose:
			b.WriteString(color.Red(line) + "\n")
		default:
			b.WriteString(color.Yellow(line) + "\n")
		}
	}
	b.WriteString(i18n.Tf("blackjackswitch.totalPayoutLine", "total", strconv.Itoa(g.GetTotalPayout())) + "\n")
	switch g.GetOverallResult() {
	case domain.GameResultWin:
		b.WriteString(color.Green(i18n.T("blackjackswitch.overallWin")) + "\n")
	case domain.GameResultLose:
		b.WriteString(color.Red(i18n.T("blackjackswitch.overallLose")) + "\n")
	default:
		b.WriteString(color.Yellow(i18n.T("blackjackswitch.overallDraw")) + "\n")
	}
}

func (bp *BlackJackSwitchCuiPresenter) phaseStr(phase int) string {
	switch phase {
	case domain.BJSwitchPhaseBet:
		return i18n.T("blackjackswitch.phaseBet")
	case domain.BJSwitchPhaseSwitch:
		return i18n.T("blackjackswitch.phaseSwitch")
	case domain.BJSwitchPhaseAction:
		return i18n.T("blackjackswitch.phaseAction")
	case domain.BJSwitchPhaseEnd:
		return i18n.T("blackjackswitch.phaseEnd")
	default:
		return i18n.T("blackjackswitch.phaseUnknown")
	}
}

func (bp *BlackJackSwitchCuiPresenter) resultStr(r domain.GameResult) string {
	switch r {
	case domain.GameResultWin:
		return i18n.T("blackjackswitch.handWin")
	case domain.GameResultDraw:
		return i18n.T("blackjackswitch.handDraw")
	default:
		return i18n.T("blackjackswitch.handLose")
	}
}
