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

// ThreeCardRummyCuiPresenter スリーカード・ラミーCUIプレゼンタークラス
type ThreeCardRummyCuiPresenter struct {
}

// Output ゲーム状態を出力
func (tp *ThreeCardRummyCuiPresenter) Output(tc interfaces.ThreeCardRummyGame, lastErr error) string {
	var sb strings.Builder

	sb.WriteString("----------\n")
	sb.WriteString(i18n.Tf("threecardrummy.chipsLine", "chips", strconv.Itoa(tc.GetChips())) + "\n")
	sb.WriteString(i18n.Tf("threecardrummy.phaseLine", "phase", tp.phaseStr(tc.GetPhase())) + "\n")
	if tc.GetPhase() == domain.ThreeCardRummyPhaseBet {
		// **低いほど強い** はこのゲーム最大の意外性。賭ける前に必ず読ませる。
		sb.WriteString(i18n.T("threecardrummy.scoringNote") + "\n")
		sb.WriteString(i18n.T("threecardrummy.qualifyNote") + "\n")
	}

	playerHand := tc.GetPlayerHand()
	if len(playerHand) > 0 {
		sb.WriteString("--- " + color.Bold(i18n.T("threecardrummy.playerHeader")) + " ---\n")
		// **点数を出す。役名ではない。** 低いほど強いので、数字がそのまま強さ。
		sb.WriteString(i18n.Tf("threecardrummy.scoreLine",
			"score", threeCardRummyScoreStr(tc.GetPlayerScore())) + "\n")
		parts := make([]string, len(playerHand))
		for i, card := range playerHand {
			parts[i] = cuiCardStr(card)
		}
		sb.WriteString(strings.Join(parts, ","))
		sb.WriteString("\n")
	}

	dealerHand := tc.GetDealerHand()
	if len(dealerHand) > 0 && tc.GetPhase() == domain.ThreeCardRummyPhaseEnd {
		sb.WriteString("--- " + color.Bold(i18n.T("threecardrummy.dealerHeader")) + " ---\n")
		sb.WriteString(i18n.Tf("threecardrummy.scoreLine",
			"score", threeCardRummyScoreStr(tc.GetDealerScore())) + "\n")
		if tc.GetDealerQualified() {
			sb.WriteString(i18n.T("threecardrummy.qualified") + "\n")
		} else {
			sb.WriteString(i18n.T("threecardrummy.notQualified") + "\n")
			// **配当の帰結は勝負した場合だけ。** 降りた手にこれを付けると、
			// 没収されたアンテについて「アンテのみ配当」と書くことになる。
			if tc.GetPlayBet() > 0 {
				sb.WriteString(i18n.T("threecardrummy.notQualifiedPayout") + "\n")
			}
		}
		parts := make([]string, len(dealerHand))
		for i, card := range dealerHand {
			parts[i] = cuiCardStr(card)
		}
		sb.WriteString(strings.Join(parts, ","))
		sb.WriteString("\n")
	}

	sb.WriteString("----------\n")

	if lastErr != nil {
		sb.WriteString(i18n.MarkErrorLine(color.Red(lastErr.Error())) + "\n")
	}

	if tc.GetGameEndFlag() {
		sb.WriteString(i18n.Tf("threecardrummy.anteLine", "ante", strconv.Itoa(tc.GetAnteBet())) + "\n")
		if tc.GetPlayBet() > 0 {
			sb.WriteString(i18n.Tf("threecardrummy.playLine", "play", strconv.Itoa(tc.GetPlayBet())) + "\n")
		}
		// **負けた側注も見せる。** 配当行は 0 のとき省くので、side bet を置いた
		// ことすら結果画面に残らず、賭け金が黙って消えたように見えていた。
		if tc.GetLowBonusBet() > 0 {
			sb.WriteString(i18n.Tf("threecardrummy.lowBonusLine", "bonus", strconv.Itoa(tc.GetLowBonusBet())) + "\n")
		}
		switch tc.GetResult() {
		case domain.GameResultWin:
			sb.WriteString(color.Green(i18n.T("threecardrummy.playerWins")) + "\n")
		case domain.GameResultLose:
			if tc.GetPlayBet() == 0 {
				sb.WriteString(color.Red(i18n.T("threecardrummy.playerFolded")) + "\n")
			} else {
				sb.WriteString(color.Red(i18n.T("threecardrummy.dealerWins")) + "\n")
			}
		case domain.GameResultDraw:
			sb.WriteString(color.Yellow(i18n.T("threecardrummy.push")) + "\n")
		default:
		}
		// Side-bet / bonus payout breakdown (omitted when zero to stay concise).
		if bonus := tc.GetAnteBonusPayout(); bonus != 0 {
			sb.WriteString(i18n.Tf("threecardrummy.anteBonusPayoutLine", "payout", strconv.Itoa(bonus)) + "\n")
		}
		if lowBonus := tc.GetLowBonusPayout(); lowBonus != 0 {
			sb.WriteString(i18n.Tf("threecardrummy.lowBonusPayoutLine", "payout", strconv.Itoa(lowBonus)) + "\n")
		}
		sb.WriteString(i18n.Tf("threecardrummy.totalPayoutLine", "payout", strconv.Itoa(tc.GetTotalPayout())) + "\n")
		sb.WriteString("----------\n")
	}

	return sb.String()
}

// threeCardRummyShouldPlay reports whether the hand is worth the play bet.
//
// **しきい値は点数。** クローン元の Q-6-4 はポーカー役の話で、ここでは意味を
// 持たない。ディーラーは 20 点以下でクオリファイするので、それより十分低ければ
// 勝ち目がある。20 点ちょうどは引き分け含みなので、余裕を見て 1 点下に置く。
func threeCardRummyShouldPlay(score int) bool {
	return score < domain.ThreeCardRummyDealerQualifyMax
}

// HintOutput emits a play/fold recommendation during the action phase; other
// phases have no decision to advise.
func (tp *ThreeCardRummyCuiPresenter) HintOutput(tc interfaces.ThreeCardRummyGame) string {
	if tc.GetPhase() != domain.ThreeCardRummyPhaseAction {
		return i18n.T("threecardrummy.hintNone") + "\n"
	}
	if threeCardRummyShouldPlay(tc.GetPlayerScore()) {
		return color.Yellow(i18n.T("threecardrummy.hintPlay")) + "\n"
	}
	return color.Yellow(i18n.T("threecardrummy.hintFold")) + "\n"
}

// ActionLogOutput 棋譜をテキスト出力
func (tp *ThreeCardRummyCuiPresenter) ActionLogOutput(tc interfaces.ThreeCardRummyGame) string {
	return actionLogOutputText(tc)
}

// phaseStr フェーズ文字列
func (tp *ThreeCardRummyCuiPresenter) phaseStr(phase int) string {
	switch phase {
	case domain.ThreeCardRummyPhaseBet:
		return i18n.T("threecardrummy.phaseBet")
	case domain.ThreeCardRummyPhaseAction:
		return i18n.T("threecardrummy.phaseAction")
	case domain.ThreeCardRummyPhaseEnd:
		return i18n.T("threecardrummy.phaseEnd")
	default:
		return i18n.T("threecardrummy.phaseUnknown")
	}
}

// threeCardRummyScoreStr は点数の表示文字列を返す。
//
// **0 点は「役」なので別の文言にする。** 同ランク 3 枚か同スート連番 3 枚で、
// このゲームの最強手。ただの 0 と書くと、ただ点が無いだけに見える。
func threeCardRummyScoreStr(score int) string {
	if score == domain.ThreeCardRummyPerfectScore {
		return i18n.T("threecardrummy.perfectHand")
	}
	return strconv.Itoa(score)
}
