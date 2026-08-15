//go:build !js || !wasm || casino

package controller

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// fiveCardStudNoArgCommands maps no-arg CUI commands to FiveCardStud calls.
// Action(domain.FiveCardStudActionX, 0, 0) is wrapped in a thin closure
// because CommandMap binds func(T) string and the action enum/zero amounts
// have to be supplied at registration time.
var fiveCardStudNoArgCommands = cuiutil.NewCommandMap[usecase.FiveCardStudInteractorIF]().
	Add(func(si usecase.FiveCardStudInteractorIF) string {
		return si.Action(domain.FiveCardStudActionFold, 0, 0)
	}, "f", "fold").
	Add(func(si usecase.FiveCardStudInteractorIF) string {
		return si.Action(domain.FiveCardStudActionCheck, 0, 0)
	}, "ck", "check").
	Add(func(si usecase.FiveCardStudInteractorIF) string {
		return si.Action(domain.FiveCardStudActionCall, 0, 0)
	}, "c", "call").
	Add(func(si usecase.FiveCardStudInteractorIF) string {
		return si.Action(domain.FiveCardStudActionAllIn, 0, 0)
	}, "a", "allin").
	Add(usecase.FiveCardStudInteractorIF.Rebuy, "rb", "rebuy").
	Add(usecase.FiveCardStudInteractorIF.SkipRebuy, "sr", "skiprebuy").
	Add(usecase.FiveCardStudInteractorIF.Addon, "ad", "addon").
	Add(usecase.FiveCardStudInteractorIF.SkipAddon, "sa", "skipaddon").
	Add(usecase.FiveCardStudInteractorIF.Muck, "m", "muck").
	Add(usecase.FiveCardStudInteractorIF.ShowHand, "sh", "show").
	Add(usecase.FiveCardStudInteractorIF.ActionLog, "log", "l")

// fiveCardStudArgfulCommands lists alias names for argful commands handled in
// the Exec switch.
var fiveCardStudArgfulCommands = []string{
	"b", "bet", "ra", "raise",
	"bl", "bettinglimit", "tm", "tournament",
	"ante", "bi", "bringin", "sb", "smallbet", "bb", "bigbet",
	"lh", "levelhand", "ts", "tablesize",
	"mai", "metaai",
}

// FiveCardStudCuiController ファイブカードスタッドCUIコントローラークラス
type FiveCardStudCuiController struct {
	si usecase.FiveCardStudInteractorIF
}

// NewFiveCardStudCuiController コンストラクタ
func NewFiveCardStudCuiController(si usecase.FiveCardStudInteractorIF) *FiveCardStudCuiController {
	return &FiveCardStudCuiController{si: si}
}

// Exec コマンド実行
func (c *FiveCardStudCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.si.Reset() },
		append(fiveCardStudNoArgCommands.Names(), fiveCardStudArgfulCommands...),
		func(cmd string, args []string) (string, bool) {
			if fn, ok := fiveCardStudNoArgCommands.Lookup(cmd); ok {
				return fn(c.si), true
			}
			switch cmd {
			case "b", "bet":
				if len(args) < 1 {
					return cuiutil.PromptRequest(i18n.T("promptBetAmount"), "b {0}"), true
				}
				amount, err := parseAmount(args)
				if err != nil {
					return err.Error(), true
				}
				return c.si.Action(domain.FiveCardStudActionBet, amount, 0), true
			case "ra", "raise":
				if len(args) < 1 {
					return cuiutil.PromptRequest(i18n.T("promptRaiseAmount"), "ra {0}"), true
				}
				amount, err := parseAmount(args)
				if err != nil {
					return err.Error(), true
				}
				return c.si.Action(domain.FiveCardStudActionRaise, amount, 0), true
			case "bl", "bettinglimit":
				if len(args) < 1 {
					return i18n.T("holdem.bettingLimitRequired"), true
				}
				bl, err := strconv.Atoi(args[0])
				if err != nil {
					return invalidArg("holdem.invalidBettingLimit", "val", args[0]), true
				}
				cfg := c.si.GetConfig()
				cfg.BettingLimit = domain.BettingLimitType(bl)
				return c.si.ResetWithConfig(cfg, nil), true
			case "tm", "tournament":
				if len(args) < 1 {
					return i18n.T("holdem.tournamentModeRequired"), true
				}
				v, err := strconv.Atoi(args[0])
				if err != nil || v < 0 || v > 1 {
					return invalidArg("holdem.invalidTournamentMode", "val", args[0]), true
				}
				cfg := c.si.GetConfig()
				cfg.TournamentMode = v == 1
				return c.si.ResetWithConfig(cfg, nil), true
			case "ante":
				if len(args) < 1 {
					return i18n.T("fivecardstud.anteRequired"), true
				}
				v, err := strconv.Atoi(args[0])
				if err != nil {
					return invalidArg("fivecardstud.invalidAnte", "val", args[0]), true
				}
				cfg := c.si.GetConfig()
				cfg.Ante = v
				return c.si.ResetWithConfig(cfg, nil), true
			case "bi", "bringin":
				if len(args) < 1 {
					return i18n.T("fivecardstud.bringInRequired"), true
				}
				v, err := strconv.Atoi(args[0])
				if err != nil {
					return invalidArg("fivecardstud.invalidBringIn", "val", args[0]), true
				}
				cfg := c.si.GetConfig()
				cfg.BringIn = v
				return c.si.ResetWithConfig(cfg, nil), true
			case "sb", "smallbet":
				if len(args) < 1 {
					return i18n.T("fivecardstud.smallBetRequired"), true
				}
				v, err := strconv.Atoi(args[0])
				if err != nil {
					return invalidArg("fivecardstud.invalidSmallBet", "val", args[0]), true
				}
				cfg := c.si.GetConfig()
				cfg.SmallBet = v
				return c.si.ResetWithConfig(cfg, nil), true
			case "bb", "bigbet":
				if len(args) < 1 {
					return i18n.T("fivecardstud.bigBetRequired"), true
				}
				v, err := strconv.Atoi(args[0])
				if err != nil {
					return invalidArg("fivecardstud.invalidBigBet", "val", args[0]), true
				}
				cfg := c.si.GetConfig()
				cfg.BigBet = v
				return c.si.ResetWithConfig(cfg, nil), true
			case "lh", "levelhand":
				if len(args) < 1 {
					return i18n.T("holdem.levelHandRequired"), true
				}
				v, err := strconv.Atoi(args[0])
				if err != nil {
					return invalidArg("holdem.invalidLevelHand", "val", args[0]), true
				}
				cfg := c.si.GetConfig()
				cfg.AnteLevelHands = v
				return c.si.ResetWithConfig(cfg, nil), true
			case "ts", "tablesize":
				if len(args) < 1 {
					return i18n.T("holdem.tableSizeRequired"), true
				}
				v, err := strconv.Atoi(args[0])
				if err != nil {
					return invalidArg("holdem.invalidTableSize", "val", args[0]), true
				}
				cfg := c.si.GetConfig()
				cfg.TableSize = v
				return c.si.ResetWithConfig(cfg, nil), true
			case "mai", "metaai":
				if len(args) < 1 {
					return i18n.T("metaAIRequired"), true
				}
				v, err := strconv.Atoi(args[0])
				if err != nil || v < 0 || v > 1 {
					return invalidArg("invalidMetaAI", "val", args[0]), true
				}
				cfg := c.si.GetConfig()
				cfg.CpuMetaAI = v == 1
				return c.si.ResetWithConfig(cfg, nil), true
			default:
				return "", false
			}
		},
	)
}
