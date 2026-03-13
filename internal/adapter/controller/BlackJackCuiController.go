package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BlackJackCuiController ブラックジャックCUIコントローラークラス
type BlackJackCuiController struct {
	bji usecase.BlackJackInteractorIF
}

// NewBlackJackCuiController コンストラクタ
func NewBlackJackCuiController(bji usecase.BlackJackInteractorIF) *BlackJackCuiController {
	return &BlackJackCuiController{
		bji: bji,
	}
}

// Exec ゲーム実行
// コマンド例: "r", "h", "s", "b 100", "d", "sp", "i", "di", "q"
func (bcc *BlackJackCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return bcc.bji.Reset() },
		[]string{
			"h", "hit", "s", "stand", "b", "bet", "d", "doubledown", "sp", "split",
			"i", "insurance", "di", "declineinsurance", "sur", "surrender",
			"es", "earlysurrender", "des", "declineearlysurrender",
			"ssr", "setsurrenderrule", "hint", "togglehint", "soft17", "togglesoft17",
			"counting", "togglecounting", "das", "toggledas",
			"sd", "setdeckcount", "scc", "setcpucount", "scs", "setcountingsystem", "pen", "setpenetration",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "h", "hit":
				return bcc.bji.Hit(), true
			case "s", "stand":
				return bcc.bji.Stand(), true
			case "b", "bet":
				amount, errMsg, ok := cuiutil.ParseIntArg(args, "Bet amount is required.", "Invalid bet amount. Please enter a number.", 1, math.MaxInt)
				if !ok {
					return errMsg, true
				}
				ppBet := cuiutil.ParseOptionalInt(args, 1, 0)
				t3Bet := cuiutil.ParseOptionalInt(args, 2, 0)
				handCount := cuiutil.ParseOptionalInt(args, 3, 0)
				return bcc.bji.Bet(amount, ppBet, t3Bet, handCount), true
			case "d", "doubledown":
				return bcc.bji.DoubleDown(), true
			case "sp", "split":
				return bcc.bji.Split(), true
			case "i", "insurance":
				return bcc.bji.Insurance(), true
			case "di", "declineinsurance":
				return bcc.bji.DeclineInsurance(), true
			case "sur", "surrender":
				return bcc.bji.Surrender(), true
			case "es", "earlysurrender":
				return bcc.bji.EarlySurrender(), true
			case "des", "declineearlysurrender":
				return bcc.bji.DeclineEarlySurrender(), true
			case "ssr", "setsurrenderrule":
				rule, errMsg, ok := cuiutil.ParseIntArg(args, "Surrender rule is required.", "Invalid surrender rule: %s. Please enter a number (0-2).", 0, domain.BJSurrenderMax)
				if !ok {
					return errMsg, true
				}
				return bcc.bji.SetSurrenderRule(rule), true
			case "hint", "togglehint":
				return bcc.bji.ToggleHint(), true
			case "soft17", "togglesoft17":
				return bcc.bji.ToggleSoft17(), true
			case "counting", "togglecounting":
				return bcc.bji.ToggleCounting(), true
			case "das", "toggledas":
				return bcc.bji.ToggleDAS(), true
			case "sd", "setdeckcount":
				count, errMsg, ok := cuiutil.ParseIntArg(args, "Deck count is required.", "Invalid deck count. Please enter a number.", 1, math.MaxInt)
				if !ok {
					return errMsg, true
				}
				return bcc.bji.SetDeckCount(count), true
			case "scc", "setcpucount":
				count, errMsg, ok := cuiutil.ParseIntArg(args, "CPU player count is required.", "Invalid CPU player count: %s. Please enter a number (0-3).", 0, domain.BJMaxCpuPlayers)
				if !ok {
					return errMsg, true
				}
				return bcc.bji.SetCpuPlayerCount(count), true
			case "scs", "setcountingsystem":
				system, errMsg, ok := cuiutil.ParseIntArg(args, "Counting system is required.", "Invalid counting system: %s. Please enter a number (0-3).", 0, domain.BJCountingMax)
				if !ok {
					return errMsg, true
				}
				return bcc.bji.SetCountingSystem(system), true
			case "pen", "setpenetration":
				pen, errMsg, ok := cuiutil.ParseIntArg(args, "Penetration rate is required.", "Invalid penetration rate. Please enter a number.", cuiutil.NoMin, cuiutil.NoMax)
				if !ok {
					return errMsg, true
				}
				return bcc.bji.SetDeckPenetration(pen), true
			}
			return "", false
		},
	)
}
