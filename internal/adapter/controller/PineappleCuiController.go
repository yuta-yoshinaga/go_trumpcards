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

// pineappleNoArgCommands maps no-arg CUI commands to Pineapple interactor calls.
// Action(domain.PineappleActionX, 0, 0) is wrapped in a thin closure because
// CommandMap binds func(T) string and the action enum/zero amounts have to be
// supplied at registration time.
var pineappleNoArgCommands = cuiutil.NewCommandMap[usecase.PineappleInteractorIF]().
	Add(func(pi usecase.PineappleInteractorIF) string {
		return pi.Action(domain.PineappleActionFold, 0, 0)
	}, "f", "fold").
	Add(func(pi usecase.PineappleInteractorIF) string {
		return pi.Action(domain.PineappleActionCheck, 0, 0)
	}, "ck", "check").
	Add(func(pi usecase.PineappleInteractorIF) string {
		return pi.Action(domain.PineappleActionCall, 0, 0)
	}, "c", "call").
	Add(func(pi usecase.PineappleInteractorIF) string {
		return pi.Action(domain.PineappleActionAllIn, 0, 0)
	}, "a", "allin").
	Add(usecase.PineappleInteractorIF.Rebuy, "rb", "rebuy").
	Add(usecase.PineappleInteractorIF.SkipRebuy, "sr", "skiprebuy").
	Add(usecase.PineappleInteractorIF.Addon, "ad", "addon").
	Add(usecase.PineappleInteractorIF.SkipAddon, "sa", "skipaddon").
	Add(usecase.PineappleInteractorIF.Muck, "m", "muck").
	Add(usecase.PineappleInteractorIF.ShowHand, "sh", "show").
	Add(usecase.PineappleInteractorIF.ActionLog, "log", "l")

// pineappleArgfulCommands lists alias names for argful commands handled in the
// Exec switch.
var pineappleArgfulCommands = []string{
	"b", "bet", "ra", "raise", "d", "discard",
	"bl", "bettinglimit", "tm", "tournament",
	"sb", "smallblind", "bb", "bigblind",
	"lh", "levelhand", "ts", "tablesize",
	"mai", "metaai",
}

// PineappleCuiController パイナップルポーカーCUIコントローラークラス
type PineappleCuiController struct {
	pi usecase.PineappleInteractorIF
}

// NewPineappleCuiController コンストラクタ
func NewPineappleCuiController(pi usecase.PineappleInteractorIF) *PineappleCuiController {
	return &PineappleCuiController{pi: pi}
}

// Exec コマンド実行
func (c *PineappleCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.pi.Reset() },
		append(pineappleNoArgCommands.Names(), pineappleArgfulCommands...),
		func(cmd string, args []string) (string, bool) {
			if fn, ok := pineappleNoArgCommands.Lookup(cmd); ok {
				return fn(c.pi), true
			}
			switch cmd {
			case "b", "bet":
				if len(args) < 1 {
					return cuiutil.PromptRequest(i18n.T("promptBetAmount"), "b {0}"), true
				}
				amount, err := parsePineappleAmount(args)
				if err != nil {
					return err.Error(), true
				}
				return c.pi.Action(domain.PineappleActionBet, amount, 0), true
			case "ra", "raise":
				if len(args) < 1 {
					return cuiutil.PromptRequest(i18n.T("promptRaiseAmount"), "ra {0}"), true
				}
				amount, err := parsePineappleAmount(args)
				if err != nil {
					return err.Error(), true
				}
				return c.pi.Action(domain.PineappleActionRaise, amount, 0), true
			case "d", "discard":
				if len(args) < 1 {
					return i18n.T("pineapple.discardIdxRequired"), true
				}
				idx, err := strconv.Atoi(args[0])
				if err != nil {
					return invalidArg("pineapple.invalidDiscardIdx", "val", args[0]), true
				}
				return c.pi.Discard(idx), true
			case "bl", "bettinglimit":
				if len(args) < 1 {
					return i18n.T("holdem.bettingLimitRequired"), true
				}
				bl, err := strconv.Atoi(args[0])
				if err != nil {
					return invalidArg("holdem.invalidBettingLimit", "val", args[0]), true
				}
				cfg := c.pi.GetConfig()
				cfg.BettingLimit = domain.BettingLimitType(bl)
				return c.pi.ResetWithConfig(cfg, nil), true
			case "tm", "tournament":
				if len(args) < 1 {
					return i18n.T("holdem.tournamentModeRequired"), true
				}
				v, err := strconv.Atoi(args[0])
				if err != nil || v < 0 || v > 1 {
					return invalidArg("holdem.invalidTournamentMode", "val", args[0]), true
				}
				cfg := c.pi.GetConfig()
				cfg.TournamentMode = v == 1
				return c.pi.ResetWithConfig(cfg, nil), true
			case "sb", "smallblind":
				if len(args) < 1 {
					return i18n.T("holdem.smallBlindRequired"), true
				}
				v, err := strconv.Atoi(args[0])
				if err != nil {
					return invalidArg("holdem.invalidSmallBlind", "val", args[0]), true
				}
				cfg := c.pi.GetConfig()
				cfg.SmallBlind = v
				return c.pi.ResetWithConfig(cfg, nil), true
			case "bb", "bigblind":
				if len(args) < 1 {
					return i18n.T("holdem.bigBlindRequired"), true
				}
				v, err := strconv.Atoi(args[0])
				if err != nil {
					return invalidArg("holdem.invalidBigBlind", "val", args[0]), true
				}
				cfg := c.pi.GetConfig()
				cfg.BigBlind = v
				return c.pi.ResetWithConfig(cfg, nil), true
			case "lh", "levelhand":
				if len(args) < 1 {
					return i18n.T("holdem.levelHandRequired"), true
				}
				v, err := strconv.Atoi(args[0])
				if err != nil {
					return invalidArg("holdem.invalidLevelHand", "val", args[0]), true
				}
				cfg := c.pi.GetConfig()
				cfg.BlindLevelHands = v
				return c.pi.ResetWithConfig(cfg, nil), true
			case "ts", "tablesize":
				if len(args) < 1 {
					return i18n.T("holdem.tableSizeRequired"), true
				}
				v, err := strconv.Atoi(args[0])
				if err != nil {
					return invalidArg("holdem.invalidTableSize", "val", args[0]), true
				}
				cfg := c.pi.GetConfig()
				cfg.TableSize = v
				return c.pi.ResetWithConfig(cfg, nil), true
			case "mai", "metaai":
				if len(args) < 1 {
					return i18n.T("metaAIRequired"), true
				}
				v, err := strconv.Atoi(args[0])
				if err != nil || v < 0 || v > 1 {
					return invalidArg("invalidMetaAI", "val", args[0]), true
				}
				cfg := c.pi.GetConfig()
				cfg.CpuMetaAI = v == 1
				return c.pi.ResetWithConfig(cfg, nil), true
			default:
				return "", false
			}
		},
	)
}

// parsePineappleAmount 引数スライスからベット額を抽出する
func parsePineappleAmount(args []string) (int, error) {
	if len(args) < 1 {
		return 0, errors.New(i18n.T("holdem.amountRequired"))
	}
	amount, err := strconv.Atoi(args[0])
	if err != nil || amount <= 0 {
		return 0, errors.New(invalidArg("holdem.invalidAmount", "val", args[0]))
	}
	return amount, nil
}
