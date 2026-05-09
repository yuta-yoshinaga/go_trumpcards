package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

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
			switch dt.GetResult() {
			case domain.GameResultWin:
				b.WriteString(color.Green(i18n.T("dragontiger.dragonWins")) + "\n")
			case domain.GameResultLose:
				b.WriteString(color.Red(i18n.T("dragontiger.tigerWins")) + "\n")
			case domain.GameResultDraw:
				if dt.GetBetType() == domain.DragonTigerBetTie {
					b.WriteString(color.Green(i18n.T("dragontiger.tieWin")) + "\n")
				} else {
					b.WriteString(color.Yellow(i18n.T("dragontiger.tieRefund")) + "\n")
				}
			default:
			}
			b.WriteString(i18n.Tf("dragontiger.payoutLine", "payout", strconv.Itoa(dt.GetPayout())) + "\n")
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
