//go:build !js || !wasm || casino

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// bjNoArgCommands maps no-arg CUI commands to BlackJack interactor methods.
// Each Add registers one action under all of its aliases — no per-alias rows
// or duplicated validCommands list. argful commands ("b 100", "sd 5", etc.)
// stay in the switch in Exec because they need argument parsing.
var bjNoArgCommands = cuiutil.NewCommandMap[usecase.BlackJackInteractorIF]().
	Add(usecase.BlackJackInteractorIF.Hit, "h", "hit").
	Add(usecase.BlackJackInteractorIF.Stand, "s", "stand").
	Add(usecase.BlackJackInteractorIF.DoubleDown, "d", "doubledown").
	Add(usecase.BlackJackInteractorIF.Split, "sp", "split").
	Add(usecase.BlackJackInteractorIF.Insurance, "i", "insurance").
	Add(usecase.BlackJackInteractorIF.DeclineInsurance, "di", "declineinsurance").
	Add(usecase.BlackJackInteractorIF.Surrender, "sur", "surrender").
	Add(usecase.BlackJackInteractorIF.EarlySurrender, "es", "earlysurrender").
	Add(usecase.BlackJackInteractorIF.DeclineEarlySurrender, "des", "declineearlysurrender").
	Add(usecase.BlackJackInteractorIF.ToggleHint, "hint", "togglehint").
	Add(usecase.BlackJackInteractorIF.ToggleSoft17, "soft17", "togglesoft17").
	Add(usecase.BlackJackInteractorIF.ToggleCounting, "counting", "togglecounting").
	Add(usecase.BlackJackInteractorIF.ToggleDAS, "das", "toggledas")

// bjArgfulCommands lists alias names for the argful commands handled in the
// Exec switch. The CommandMap covers no-arg aliases automatically; these have
// to be listed by hand because they aren't bound through CommandMap.
var bjArgfulCommands = []string{
	"b", "bet",
	"ssr", "setsurrenderrule",
	"sd", "setdeckcount",
	"scc", "setcpucount",
	"scs", "setcountingsystem",
	"pen", "setpenetration",
	"log", "l",
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
		append(bjNoArgCommands.Names(), bjArgfulCommands...),
		func(cmd string, args []string) (string, bool) {
			if fn, ok := bjNoArgCommands.Lookup(cmd); ok {
				return fn(bcc.bji), true
			}
			switch cmd {
			case "b", "bet":
				amount, errMsg, ok := cuiutil.ParseIntArgKeys(args, "betAmountRequired", "invalidBetAmount", 1, math.MaxInt)
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
