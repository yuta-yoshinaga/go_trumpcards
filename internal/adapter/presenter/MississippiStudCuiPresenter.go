//go:build !js || !wasm || casino

package presenter

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// MississippiStudCuiPresenter ミシシッピ・スタッドCUIプレゼンター
type MississippiStudCuiPresenter struct{}

// Output ゲーム状態を出力する。
func (mp *MississippiStudCuiPresenter) Output(g interfaces.MississippiStudGame, lastErr error) string {
	return buildCuiOutput(i18n.T("mississippistud.title"), func(b *strings.Builder) {
		fmt.Fprintf(b, "%s\n", i18n.Tf("mississippistud.chipsLine", "chips", strconv.Itoa(g.GetChips())))
		fmt.Fprintf(b, "%s\n", i18n.Tf("mississippistud.phaseLine", "phase", mp.phaseStr(g.GetPhase())))

		if g.GetAnteAmount() > 0 {
			fmt.Fprintf(b, "%s\n", i18n.Tf("mississippistud.anteLine", "amount", strconv.Itoa(g.GetAnteAmount())))
		}

		playerHand := g.GetPlayerHand()
		if len(playerHand) > 0 {
			b.WriteString("--- " + color.Bold(i18n.T("mississippistud.playerHeader")) + " ---\n")
			b.WriteString(formatCardSlice(playerHand, cuiCardStr, ","))
			b.WriteString("\n")
		}

		community := g.GetCommunityCards()
		revealed := g.GetCommunityRevealed()
		if len(community) > 0 {
			b.WriteString("--- " + color.Bold(i18n.T("mississippistud.communityHeader")) + " ---\n")
			parts := make([]string, len(community))
			for i, c := range community {
				if i < len(revealed) && revealed[i] {
					parts[i] = cuiCardStr(c)
				} else {
					parts[i] = "??"
				}
			}
			b.WriteString(strings.Join(parts, ","))
			b.WriteString("\n")
		}

		mults := g.GetStreetMultipliers()
		if mults[0] > 0 || mults[1] > 0 || mults[2] > 0 {
			fmt.Fprintf(b, "%s\n", i18n.Tf("mississippistud.streetsLine",
				"s3", mp.multiplierStr(mults[0]),
				"s4", mp.multiplierStr(mults[1]),
				"s5", mp.multiplierStr(mults[2]),
			))
		}

		// During the streets, show the accumulated bet and what a fold forfeits;
		// the game-end block prints totalBet separately, so guard against dupes.
		if !g.GetGameEndFlag() && g.GetTotalBet() > 0 {
			fmt.Fprintf(b, "%s\n", i18n.Tf("mississippistud.totalBetLine", "bet", strconv.Itoa(g.GetTotalBet())))
			fmt.Fprintf(b, "%s\n", i18n.Tf("mississippistud.foldLossLine", "loss", strconv.Itoa(g.GetTotalBet())))
		}

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			if rank := g.GetHandRank(); rank >= 0 && rank < len(domain.PokerHandNames) && !g.GetFolded() {
				fmt.Fprintf(b, "%s\n", i18n.Tf("mississippistud.handLine", "hand", cuiPokerHandName(rank)))
			}
			// **配当がどの倍率から来たのかを説明していなかった** (#5591)。
			// プッシュ (-1) とロス (0) は倍率ではないので出さない ── 出すと
			// 「x-1」「x0」という存在しないオッズに読める。
			if mult := g.GetPayoutMultiplier(); mult > 0 {
				fmt.Fprintf(b, "%s\n", i18n.Tf("mississippistud.payoutMultiplierLine",
					"multiplier", strconv.Itoa(mult)))
			}
			switch g.GetResult() {
			case domain.GameResultWin:
				b.WriteString(color.Green(i18n.T("mississippistud.playerWins")) + "\n")
			case domain.GameResultDraw:
				b.WriteString(color.Yellow(i18n.T("mississippistud.push")) + "\n")
			case domain.GameResultLose:
				b.WriteString(color.Red(i18n.T("mississippistud.playerLoses")) + "\n")
			}
			fmt.Fprintf(b, "%s\n", i18n.Tf("mississippistud.totalBetLine", "bet", strconv.Itoa(g.GetTotalBet())))
			fmt.Fprintf(b, "%s\n", i18n.Tf("mississippistud.totalPayoutLine", "payout", strconv.Itoa(g.GetTotalPayout())))
		}
	})
}

// msHintKeys は推奨アクションから表示用の i18n キーへの対応。
var msHintKeys = map[string]string{
	domain.MSRecommendPlay3x: "mississippistud.hintPlay3x",
	domain.MSRecommendPlay1x: "mississippistud.hintPlay1x",
	domain.MSRecommendFold:   "mississippistud.hintFold",
}

// HintOutput は現在の役・配当対象かどうか・推奨倍率を出力する。
//
// **Web は ms-made-hand に役と配当対象かを常時出し、ヒントも表示している
// のに、CUI にはどちらも無かった (#4710)。**判定はドメインの
// GetCurrentMadeHand / RecommendBet 1か所に置いたので、CUI と Web が
// 違うことを言うことはない。
func (mp *MississippiStudCuiPresenter) HintOutput(g interfaces.MississippiStudGame) string {
	key, ok := msHintKeys[g.RecommendBet()]
	if !ok {
		return i18n.T("mississippistud.hintNone") + "\n"
	}
	var b strings.Builder
	// **役の名前だけでは足りない。**2 のペアは「ワンペア」でも配当が付かない。
	if made := g.GetCurrentMadeHand(); made != nil {
		eligible := i18n.T("mississippistud.madeHandNotEligible")
		if made.PaytableEligible {
			eligible = i18n.T("mississippistud.madeHandEligible")
		}
		b.WriteString(i18n.Tf("mississippistud.madeHandLine",
			"hand", cuiPokerHandName(made.Rank), "eligible", eligible) + "\n")
	}
	b.WriteString(color.Yellow(i18n.T(key)) + "\n")
	return b.String()
}

// ActionLogOutput 棋譜をテキスト出力する。
func (mp *MississippiStudCuiPresenter) ActionLogOutput(g interfaces.MississippiStudGame) string {
	return actionLogOutputText(g)
}

func (mp *MississippiStudCuiPresenter) phaseStr(phase int) string {
	switch phase {
	case domain.MississippiStudPhaseAnte:
		return i18n.T("mississippistud.phaseAnte")
	case domain.MississippiStudPhaseThirdSt:
		return i18n.T("mississippistud.phaseThirdSt")
	case domain.MississippiStudPhaseFourthSt:
		return i18n.T("mississippistud.phaseFourthSt")
	case domain.MississippiStudPhaseFifthSt:
		return i18n.T("mississippistud.phaseFifthSt")
	case domain.MississippiStudPhaseEnd:
		return i18n.T("mississippistud.phaseEnd")
	default:
		return i18n.T("mississippistud.phaseUnknown")
	}
}

func (mp *MississippiStudCuiPresenter) multiplierStr(mult int) string {
	if mult == 0 {
		return "-"
	}
	return fmt.Sprintf("x%d", mult)
}
