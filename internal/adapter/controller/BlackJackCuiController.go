package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// bjNoArgCommands 引数なしのブラックジャックコマンドマップ
var bjNoArgCommands = map[string]func(usecase.BlackJackInteractorIF) string{
	"h":                     usecase.BlackJackInteractorIF.Hit,
	"hit":                   usecase.BlackJackInteractorIF.Hit,
	"s":                     usecase.BlackJackInteractorIF.Stand,
	"stand":                 usecase.BlackJackInteractorIF.Stand,
	"d":                     usecase.BlackJackInteractorIF.DoubleDown,
	"doubledown":            usecase.BlackJackInteractorIF.DoubleDown,
	"sp":                    usecase.BlackJackInteractorIF.Split,
	"split":                 usecase.BlackJackInteractorIF.Split,
	"i":                     usecase.BlackJackInteractorIF.Insurance,
	"insurance":             usecase.BlackJackInteractorIF.Insurance,
	"di":                    usecase.BlackJackInteractorIF.DeclineInsurance,
	"declineinsurance":      usecase.BlackJackInteractorIF.DeclineInsurance,
	"sur":                   usecase.BlackJackInteractorIF.Surrender,
	"surrender":             usecase.BlackJackInteractorIF.Surrender,
	"es":                    usecase.BlackJackInteractorIF.EarlySurrender,
	"earlysurrender":        usecase.BlackJackInteractorIF.EarlySurrender,
	"des":                   usecase.BlackJackInteractorIF.DeclineEarlySurrender,
	"declineearlysurrender": usecase.BlackJackInteractorIF.DeclineEarlySurrender,
	"hint":                  usecase.BlackJackInteractorIF.ToggleHint,
	"togglehint":            usecase.BlackJackInteractorIF.ToggleHint,
	"soft17":                usecase.BlackJackInteractorIF.ToggleSoft17,
	"togglesoft17":          usecase.BlackJackInteractorIF.ToggleSoft17,
	"counting":              usecase.BlackJackInteractorIF.ToggleCounting,
	"togglecounting":        usecase.BlackJackInteractorIF.ToggleCounting,
	"das":                   usecase.BlackJackInteractorIF.ToggleDAS,
	"toggledas":             usecase.BlackJackInteractorIF.ToggleDAS,
}

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
			"log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			if fn, ok := bjNoArgCommands[cmd]; ok {
				return fn(bcc.bji), true
			}
			switch cmd {
			case "b", "bet":
				amount, errMsg, ok := cuiutil.ParseIntArg(args, "Bet amount is required.", "Invalid bet amount. Please enter a number.", 1, math.MaxInt)
				if !ok {
					return errMsg, true
				}
				ppBet := cuiutil.ParseOptionalInt(args, 1, 0)
				t3Bet := cuiutil.ParseOptionalInt(args, 2, 0)
				handCount := cuiutil.ParseOptionalInt(args, 3, 0)
				return bcc.bji.Bet(amount, ppBet, t3Bet, handCount), true
			case "ssr", "setsurrenderrule":
				return cuiutil.WithParsedInt(args, "Surrender rule is required.", "Invalid surrender rule: %s. Please enter a number (0-2).", 0, domain.BJSurrenderMax, bcc.bji.SetSurrenderRule)
			case "sd", "setdeckcount":
				return cuiutil.WithParsedInt(args, "Deck count is required.", "Invalid deck count. Please enter a number.", 1, math.MaxInt, bcc.bji.SetDeckCount)
			case "scc", "setcpucount":
				return cuiutil.WithParsedInt(args, "CPU player count is required.", "Invalid CPU player count: %s. Please enter a number (0-3).", 0, domain.BJMaxCpuPlayers, bcc.bji.SetCpuPlayerCount)
			case "scs", "setcountingsystem":
				return cuiutil.WithParsedInt(args, "Counting system is required.", "Invalid counting system: %s. Please enter a number (0-3).", 0, domain.BJCountingMax, bcc.bji.SetCountingSystem)
			case "pen", "setpenetration":
				return cuiutil.WithParsedInt(args, "Penetration rate is required.", "Invalid penetration rate. Please enter a number.", cuiutil.NoMin, cuiutil.NoMax, bcc.bji.SetDeckPenetration)
			default:
				return handleCuiLog(cmd, bcc.bji.ActionLog)
			}
		},
	)
}
