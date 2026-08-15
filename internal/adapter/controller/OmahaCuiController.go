//go:build !js || !wasm || casino

package controller

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// omahaNoArgCommands maps no-arg CUI commands to Omaha interactor calls.
// Action(domain.OmahaActionX, 0, 0) is wrapped in a thin closure because
// CommandMap binds func(T) string and the action enum/zero amounts have to be
// supplied at registration time.
var omahaNoArgCommands = cuiutil.NewCommandMap[usecase.OmahaInteractorIF]().
	Add(func(oi usecase.OmahaInteractorIF) string {
		return oi.Action(domain.OmahaActionFold, 0, 0)
	}, "f", "fold").
	Add(func(oi usecase.OmahaInteractorIF) string {
		return oi.Action(domain.OmahaActionCheck, 0, 0)
	}, "ck", "check").
	Add(func(oi usecase.OmahaInteractorIF) string {
		return oi.Action(domain.OmahaActionCall, 0, 0)
	}, "c", "call").
	Add(func(oi usecase.OmahaInteractorIF) string {
		return oi.Action(domain.OmahaActionAllIn, 0, 0)
	}, "a", "allin").
	Add(usecase.OmahaInteractorIF.Rebuy, "rb", "rebuy").
	Add(usecase.OmahaInteractorIF.SkipRebuy, "sr", "skiprebuy").
	Add(usecase.OmahaInteractorIF.Addon, "ad", "addon").
	Add(usecase.OmahaInteractorIF.SkipAddon, "sa", "skipaddon").
	Add(usecase.OmahaInteractorIF.Muck, "m", "muck").
	Add(usecase.OmahaInteractorIF.ShowHand, "sh", "show").
	Add(usecase.OmahaInteractorIF.ActionLog, "log", "l")

// omahaArgfulCommands lists alias names for argful commands handled in the
// Exec switch.
var omahaArgfulCommands = []string{
	"b", "bet", "ra", "raise",
	"bl", "bettinglimit", "tm", "tournament",
	"sb", "smallblind", "bb", "bigblind",
	"lh", "levelhand", "ts", "tablesize",
	"mai", "metaai",
}

// OmahaCuiController オマハホールデムCUIコントローラークラス
type OmahaCuiController struct {
	oi usecase.OmahaInteractorIF
}

// NewOmahaCuiController コンストラクタ
func NewOmahaCuiController(oi usecase.OmahaInteractorIF) *OmahaCuiController {
	return &OmahaCuiController{oi: oi}
}

// Exec コマンド実行
func (c *OmahaCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.oi.Reset() },
		append(omahaNoArgCommands.Names(), omahaArgfulCommands...),
		func(cmd string, args []string) (string, bool) {
			if fn, ok := omahaNoArgCommands.Lookup(cmd); ok {
				return fn(c.oi), true
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
				return c.oi.Action(domain.OmahaActionBet, amount, 0), true
			case "ra", "raise":
				if len(args) < 1 {
					return cuiutil.PromptRequest(i18n.T("promptRaiseAmount"), "ra {0}"), true
				}
				amount, err := parseAmount(args)
				if err != nil {
					return err.Error(), true
				}
				return c.oi.Action(domain.OmahaActionRaise, amount, 0), true
			case "bl", "bettinglimit":
				if len(args) < 1 {
					return i18n.T("holdem.bettingLimitRequired"), true
				}
				bl, err := strconv.Atoi(args[0])
				if err != nil {
					return invalidArg("holdem.invalidBettingLimit", "val", args[0]), true
				}
				cfg := c.oi.GetConfig()
				cfg.BettingLimit = domain.BettingLimitType(bl)
				return c.oi.ResetWithConfig(cfg, nil), true
			case "tm", "tournament":
				if len(args) < 1 {
					return i18n.T("holdem.tournamentModeRequired"), true
				}
				v, err := strconv.Atoi(args[0])
				if err != nil || v < 0 || v > 1 {
					return invalidArg("holdem.invalidTournamentMode", "val", args[0]), true
				}
				cfg := c.oi.GetConfig()
				cfg.TournamentMode = v == 1
				return c.oi.ResetWithConfig(cfg, nil), true
			case "sb", "smallblind":
				if len(args) < 1 {
					return i18n.T("holdem.smallBlindRequired"), true
				}
				v, err := strconv.Atoi(args[0])
				if err != nil {
					return invalidArg("holdem.invalidSmallBlind", "val", args[0]), true
				}
				cfg := c.oi.GetConfig()
				cfg.SmallBlind = v
				return c.oi.ResetWithConfig(cfg, nil), true
			case "bb", "bigblind":
				if len(args) < 1 {
					return i18n.T("holdem.bigBlindRequired"), true
				}
				v, err := strconv.Atoi(args[0])
				if err != nil {
					return invalidArg("holdem.invalidBigBlind", "val", args[0]), true
				}
				cfg := c.oi.GetConfig()
				cfg.BigBlind = v
				return c.oi.ResetWithConfig(cfg, nil), true
			case "lh", "levelhand":
				if len(args) < 1 {
					return i18n.T("holdem.levelHandRequired"), true
				}
				v, err := strconv.Atoi(args[0])
				if err != nil {
					return invalidArg("holdem.invalidLevelHand", "val", args[0]), true
				}
				cfg := c.oi.GetConfig()
				cfg.BlindLevelHands = v
				return c.oi.ResetWithConfig(cfg, nil), true
			case "ts", "tablesize":
				if len(args) < 1 {
					return i18n.T("holdem.tableSizeRequired"), true
				}
				v, err := strconv.Atoi(args[0])
				if err != nil {
					return invalidArg("holdem.invalidTableSize", "val", args[0]), true
				}
				cfg := c.oi.GetConfig()
				cfg.TableSize = v
				return c.oi.ResetWithConfig(cfg, nil), true
			case "mai", "metaai":
				if len(args) < 1 {
					return i18n.T("metaAIRequired"), true
				}
				v, err := strconv.Atoi(args[0])
				if err != nil || v < 0 || v > 1 {
					return invalidArg("invalidMetaAI", "val", args[0]), true
				}
				cfg := c.oi.GetConfig()
				cfg.CpuMetaAI = v == 1
				return c.oi.ResetWithConfig(cfg, nil), true
			default:
				return "", false
			}
		},
	)
}
