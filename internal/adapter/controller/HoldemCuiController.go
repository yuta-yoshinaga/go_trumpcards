//go:build !js || !wasm || casino

package controller

import (
	"errors"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// holdemNoArgCommands maps no-arg CUI commands to Holdem interactor calls.
// Action(domain.HoldemActionX, 0, 0) is wrapped in a thin closure because
// CommandMap binds func(T) string and the action enum/zero amounts have to be
// supplied at registration time. argful commands ("b 100", "sb 50", etc.) stay
// in the switch in Exec because they need argument parsing.
var holdemNoArgCommands = cuiutil.NewCommandMap[usecase.HoldemInteractorIF]().
	Add(func(hi usecase.HoldemInteractorIF) string {
		return hi.Action(domain.HoldemActionFold, 0, 0)
	}, "f", "fold").
	Add(func(hi usecase.HoldemInteractorIF) string {
		return hi.Action(domain.HoldemActionCheck, 0, 0)
	}, "ck", "check").
	Add(func(hi usecase.HoldemInteractorIF) string {
		return hi.Action(domain.HoldemActionCall, 0, 0)
	}, "c", "call").
	Add(func(hi usecase.HoldemInteractorIF) string {
		return hi.Action(domain.HoldemActionAllIn, 0, 0)
	}, "a", "allin").
	Add(usecase.HoldemInteractorIF.Rebuy, "rb", "rebuy").
	Add(usecase.HoldemInteractorIF.SkipRebuy, "sr", "skiprebuy").
	Add(usecase.HoldemInteractorIF.Addon, "ad", "addon").
	Add(usecase.HoldemInteractorIF.SkipAddon, "sa", "skipaddon").
	Add(usecase.HoldemInteractorIF.Muck, "m", "muck").
	Add(usecase.HoldemInteractorIF.ShowHand, "sh", "show").
	Add(usecase.HoldemInteractorIF.ActionLog, "log", "l")

// holdemArgfulCommands lists alias names for the argful commands handled in
// the Exec switch. The CommandMap covers no-arg aliases automatically; these
// are listed by hand because they aren't bound through CommandMap.
var holdemArgfulCommands = []string{
	"b", "bet", "ra", "raise",
	"bl", "bettinglimit", "tm", "tournament",
	"sb", "smallblind", "bb", "bigblind",
	"lh", "levelhand", "ts", "tablesize",
	"mai", "metaai",
}

// HoldemCuiController テキサスホールデムCUIコントローラークラス
type HoldemCuiController struct {
	hi usecase.HoldemInteractorIF
}

// NewHoldemCuiController コンストラクタ
func NewHoldemCuiController(hi usecase.HoldemInteractorIF) *HoldemCuiController {
	return &HoldemCuiController{hi: hi}
}

// Exec コマンド実行
func (c *HoldemCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.hi.Reset() },
		append(holdemNoArgCommands.Names(), holdemArgfulCommands...),
		func(cmd string, args []string) (string, bool) {
			if fn, ok := holdemNoArgCommands.Lookup(cmd); ok {
				return fn(c.hi), true
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
				return c.hi.Action(domain.HoldemActionBet, amount, 0), true
			case "ra", "raise":
				if len(args) < 1 {
					return cuiutil.PromptRequest(i18n.T("promptRaiseAmount"), "ra {0}"), true
				}
				amount, err := parseAmount(args)
				if err != nil {
					return err.Error(), true
				}
				return c.hi.Action(domain.HoldemActionRaise, amount, 0), true
			case "bl", "bettinglimit":
				if len(args) < 1 {
					return i18n.T("holdem.bettingLimitRequired"), true
				}
				bl, err := strconv.Atoi(args[0])
				if err != nil {
					return invalidArg("holdem.invalidBettingLimit", "val", args[0]), true
				}
				cfg := c.hi.GetConfig()
				cfg.BettingLimit = domain.BettingLimitType(bl)
				return c.hi.ResetWithConfig(cfg, nil), true
			case "tm", "tournament":
				if len(args) < 1 {
					return i18n.T("holdem.tournamentModeRequired"), true
				}
				v, err := strconv.Atoi(args[0])
				if err != nil || v < 0 || v > 1 {
					return invalidArg("holdem.invalidTournamentMode", "val", args[0]), true
				}
				cfg := c.hi.GetConfig()
				cfg.TournamentMode = v == 1
				return c.hi.ResetWithConfig(cfg, nil), true
			case "sb", "smallblind":
				if len(args) < 1 {
					return i18n.T("holdem.smallBlindRequired"), true
				}
				v, err := strconv.Atoi(args[0])
				if err != nil {
					return invalidArg("holdem.invalidSmallBlind", "val", args[0]), true
				}
				cfg := c.hi.GetConfig()
				cfg.SmallBlind = v
				return c.hi.ResetWithConfig(cfg, nil), true
			case "bb", "bigblind":
				if len(args) < 1 {
					return i18n.T("holdem.bigBlindRequired"), true
				}
				v, err := strconv.Atoi(args[0])
				if err != nil {
					return invalidArg("holdem.invalidBigBlind", "val", args[0]), true
				}
				cfg := c.hi.GetConfig()
				cfg.BigBlind = v
				return c.hi.ResetWithConfig(cfg, nil), true
			case "lh", "levelhand":
				if len(args) < 1 {
					return i18n.T("holdem.levelHandRequired"), true
				}
				v, err := strconv.Atoi(args[0])
				if err != nil {
					return invalidArg("holdem.invalidLevelHand", "val", args[0]), true
				}
				cfg := c.hi.GetConfig()
				cfg.BlindLevelHands = v
				return c.hi.ResetWithConfig(cfg, nil), true
			case "ts", "tablesize":
				if len(args) < 1 {
					return i18n.T("holdem.tableSizeRequired"), true
				}
				v, err := strconv.Atoi(args[0])
				if err != nil {
					return invalidArg("holdem.invalidTableSize", "val", args[0]), true
				}
				cfg := c.hi.GetConfig()
				cfg.TableSize = v
				return c.hi.ResetWithConfig(cfg, nil), true
			case "mai", "metaai":
				if len(args) < 1 {
					return i18n.T("metaAIRequired"), true
				}
				v, err := strconv.Atoi(args[0])
				if err != nil || v < 0 || v > 1 {
					return invalidArg("invalidMetaAI", "val", args[0]), true
				}
				cfg := c.hi.GetConfig()
				cfg.CpuMetaAI = v == 1
				return c.hi.ResetWithConfig(cfg, nil), true
			default:
				return "", false
			}
		},
	)
}

// parseAmount 引数スライスからベット額を抽出する
func parseAmount(args []string) (int, error) {
	if len(args) < 1 {
		return 0, errors.New(i18n.T("holdem.amountRequired"))
	}
	amount, err := strconv.Atoi(args[0])
	if err != nil || amount <= 0 {
		return 0, errors.New(invalidArg("holdem.invalidAmount", "val", args[0]))
	}
	return amount, nil
}
