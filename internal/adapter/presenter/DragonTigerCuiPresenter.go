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

// dragonTigerHistoryMaxShown caps the trailing run of results shown so a long
// shoe does not overflow the terminal line.
const dragonTigerHistoryMaxShown = 20

// DragonTigerCuiPresenter ドラゴンタイガーCUIプレゼンタークラス
type DragonTigerCuiPresenter struct{}

// Output ゲーム状態を出力
func (dp *DragonTigerCuiPresenter) Output(dt interfaces.DragonTigerGame, lastErr error) string {
	return buildCuiOutput(i18n.T("dragontiger.outputTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("dragontiger.chipsLine", "chips", strconv.Itoa(dt.GetChips())) + "\n")
		b.WriteString(i18n.Tf("dragontiger.phaseLine", "phase", dp.phaseStr(dt.GetPhase())) + "\n")

		if dt.GetBetAmount() > 0 {
			b.WriteString(i18n.Tf("dragontiger.betLine",
				"amount", strconv.Itoa(dt.GetBetAmount()),
				"type", dp.betTypeStr(dt.GetBetType()),
			) + "\n")
		}

		dragon, tiger := dt.GetDragonCard(), dt.GetTigerCard()
		if dragon != nil || tiger != nil {
			b.WriteString("----------\n")
			if dragon != nil {
				b.WriteString(i18n.Tf("dragontiger.dragonLine", "card", cuiCardStr(dragon)) + "\n")
			}
			if tiger != nil {
				b.WriteString(i18n.Tf("dragontiger.tigerLine", "card", cuiCardStr(tiger)) + "\n")
			}
		}

		cuiErrorBlock(b, lastErr)

		if dt.GetGameEndFlag() {
			// Color is player-centric: green when the player's bet wins,
			// red when it loses. The result message itself names the game-side
			// winner (Dragon / Tiger) regardless of which side the player took.
			betType := dt.GetBetType()
			switch dt.GetResult() {
			case domain.GameResultWin: // Dragon wins
				msg := i18n.T("dragontiger.dragonWins")
				if betType == domain.DragonTigerBetDragon {
					b.WriteString(color.Green(msg) + "\n")
				} else {
					b.WriteString(color.Red(msg) + "\n")
				}
			case domain.GameResultLose: // Tiger wins
				msg := i18n.T("dragontiger.tigerWins")
				if betType == domain.DragonTigerBetTiger {
					b.WriteString(color.Green(msg) + "\n")
				} else {
					b.WriteString(color.Red(msg) + "\n")
				}
			case domain.GameResultDraw:
				if betType == domain.DragonTigerBetTie {
					b.WriteString(color.Green(i18n.T("dragontiger.tieWin")) + "\n")
				} else {
					b.WriteString(color.Yellow(i18n.T("dragontiger.tieRefund")) + "\n")
				}
			default:
			}
			// Show the bet type and its odds (Tie pays 8:1, Dragon/Tiger 1:1) so the
			// payout figure below is self-explanatory.
			odds := 1
			betTypeKey := "dragontiger.betTypeDragon"
			switch betType {
			case domain.DragonTigerBetTiger:
				betTypeKey = "dragontiger.betTypeTiger"
			case domain.DragonTigerBetTie:
				betTypeKey = "dragontiger.betTypeTie"
				odds = 8
			}
			b.WriteString(i18n.Tf("dragontiger.oddsLine", "type", i18n.T(betTypeKey), "odds", strconv.Itoa(odds)) + "\n")
			b.WriteString(i18n.Tf("dragontiger.payoutLine", "payout", strconv.Itoa(dt.GetPayout())) + "\n")
		}

		// Big-road history: recent results as a colored D/T/= row plus totals.
		if history := dt.GetHistory(); len(history) > 0 {
			dCount, tCount, tieCount := 0, 0, 0
			for _, r := range history {
				switch r {
				case domain.DragonTigerResultDragon:
					dCount++
				case domain.DragonTigerResultTiger:
					tCount++
				case domain.DragonTigerResultTie:
					tieCount++
				}
			}
			shown := history
			if len(shown) > dragonTigerHistoryMaxShown {
				shown = shown[len(shown)-dragonTigerHistoryMaxShown:]
			}
			syms := make([]string, len(shown))
			for i, r := range shown {
				switch r {
				case domain.DragonTigerResultDragon:
					syms[i] = color.Red("D")
				case domain.DragonTigerResultTiger:
					syms[i] = color.Yellow("T")
				case domain.DragonTigerResultTie:
					syms[i] = "="
				default:
					syms[i] = "?"
				}
			}
			b.WriteString(i18n.Tf("dragontiger.historyLine", "symbols", strings.Join(syms, " ")) + "\n")
			b.WriteString(i18n.Tf("dragontiger.historyCounts",
				"d", strconv.Itoa(dCount),
				"t", strconv.Itoa(tCount),
				"tie", strconv.Itoa(tieCount)) + "\n")
		}
	})
}

// ActionLogOutput 棋譜をテキスト出力
func (dp *DragonTigerCuiPresenter) ActionLogOutput(dt interfaces.DragonTigerGame) string {
	return actionLogOutputText(dt)
}

// phaseStr フェーズ文字列
func (dp *DragonTigerCuiPresenter) phaseStr(phase int) string {
	switch phase {
	case domain.DragonTigerPhaseBet:
		return i18n.T("dragontiger.phaseBet")
	case domain.DragonTigerPhaseEnd:
		return i18n.T("dragontiger.phaseEnd")
	default:
		return i18n.T("dragontiger.phaseUnknown")
	}
}

// betTypeStr ベットタイプ文字列
func (dp *DragonTigerCuiPresenter) betTypeStr(betType int) string {
	switch betType {
	case domain.DragonTigerBetDragon:
		return i18n.T("dragontiger.betTypeDragon")
	case domain.DragonTigerBetTiger:
		return i18n.T("dragontiger.betTypeTiger")
	case domain.DragonTigerBetTie:
		return i18n.T("dragontiger.betTypeTie")
	default:
		return i18n.T("dragontiger.betTypeUnknown")
	}
}
