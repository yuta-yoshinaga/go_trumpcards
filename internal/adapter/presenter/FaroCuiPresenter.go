//go:build !js || !wasm || extra2

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// FaroCuiPresenter はファロのCUIプレゼンター。
type FaroCuiPresenter struct{}

// Output ゲーム状態を出力する。
func (fp *FaroCuiPresenter) Output(f interfaces.FaroGame, lastErr error) string {
	var sb strings.Builder

	sb.WriteString("----------\n")
	sb.WriteString(i18n.Tf("faro.chipsLine", "chips", strconv.Itoa(f.GetChips())) + "\n")
	sb.WriteString(i18n.Tf("faro.phaseLine", "phase", fp.phaseStr(f.GetPhase())) + "\n")
	sb.WriteString(i18n.Tf("faro.turnsLine", "played", strconv.Itoa(f.GetTurnsPlayed()), "total", strconv.Itoa(f.GetTurnsTotal())) + "\n")
	sb.WriteString(i18n.Tf("faro.remainingLine", "count", strconv.Itoa(f.GetRemainingCount())) + "\n")

	// **ケースキーパーはこのゲームの中核。**Web はランク別の残数を常時出すのに、
	// CUI は総枚数しか出しておらず、勘だけで賭けることになっていた (#4894)。
	remaining := f.FaroRemainingByRank()
	parts := make([]string, 0, domain.FaroMaxRank)
	for r := 1; r <= domain.FaroMaxRank; r++ {
		parts = append(parts, cuiRankLabel(r)+":"+strconv.Itoa(remaining[r]))
	}
	sb.WriteString(i18n.Tf("faro.caseKeeper", "counts", strings.Join(parts, " ")) + "\n")

	ranks := f.GetBetRanks()
	bets := f.GetBets()
	if len(ranks) > 0 {
		sb.WriteString("--- " + color.Bold(i18n.T("faro.betsHeader")) + " ---\n")
		for _, r := range ranks {
			b := bets[r]
			tag := ""
			if b.Copper {
				tag = " " + i18n.T("faro.copperTag")
			}
			sb.WriteString(i18n.Tf("faro.betLine", "rank", cuiRankLabel(r), "amount", strconv.Itoa(b.Amount)) + tag + "\n")
		}
	}

	if lt := f.GetLastTurn(); lt != nil {
		sb.WriteString("--- " + color.Bold(i18n.T("faro.lastTurnHeader")) + " ---\n")
		sb.WriteString(i18n.T("faro.losingLabel") + " " + cuiCardStr(lt.LosingCard) + "\n")
		sb.WriteString(i18n.T("faro.winningLabel") + " " + cuiCardStr(lt.WinningCard) + "\n")
		if lt.Split {
			sb.WriteString(color.Yellow(i18n.T("faro.splitLine")) + "\n")
		}
	}

	if cards := f.GetCallCards(); len(cards) > 0 {
		sb.WriteString("--- " + color.Bold(i18n.T("faro.callHeader")) + " ---\n")
		parts := make([]string, len(cards))
		for i, c := range cards {
			parts[i] = cuiCardStr(c)
		}
		sb.WriteString(strings.Join(parts, ",") + "\n")
	}

	sb.WriteString("----------\n")

	if lastErr != nil {
		sb.WriteString(color.Red(lastErr.Error()) + "\n")
	}

	if f.GetPhase() == domain.FaroPhaseRoundEnd || f.GetPhase() == domain.FaroPhaseGameEnd {
		if f.GetCallOrder() != nil {
			if f.GetCallWon() {
				sb.WriteString(color.Green(i18n.T("faro.callWon")) + "\n")
			} else {
				sb.WriteString(color.Red(i18n.T("faro.callLost")) + "\n")
			}
		}
		sb.WriteString(i18n.Tf("faro.payoutLine", "payout", strconv.Itoa(f.GetTotalPayout())) + "\n")
		sb.WriteString("----------\n")
	}
	if f.GetGameEndFlag() {
		sb.WriteString(color.Red(i18n.T("faro.gameEnd")) + "\n")
	}

	return sb.String()
}

// ActionLogOutput 棋譜をテキスト出力する。
func (fp *FaroCuiPresenter) ActionLogOutput(f interfaces.FaroGame) string {
	return actionLogOutputText(f)
}

// phaseStr フェーズ文字列。
func (fp *FaroCuiPresenter) phaseStr(phase int) string {
	switch phase {
	case domain.FaroPhaseBetting:
		return i18n.T("faro.phaseBetting")
	case domain.FaroPhaseTurn:
		return i18n.T("faro.phaseTurn")
	case domain.FaroPhaseCall:
		return i18n.T("faro.phaseCall")
	case domain.FaroPhaseRoundEnd:
		return i18n.T("faro.phaseRoundEnd")
	case domain.FaroPhaseGameEnd:
		return i18n.T("faro.phaseGameEnd")
	default:
		return i18n.T("faro.phaseUnknown")
	}
}
