//go:build !js || !wasm || casino

package controller

import (
	"math"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// dramahaNoArgCommands maps no-arg CUI commands to Dramaha interactor calls.
// Action(domain.DramahaActionX, 0, 0) is wrapped in a thin closure because
// CommandMap binds func(T) string and the action enum/zero amounts have to be
// supplied at registration time.
var dramahaNoArgCommands = cuiutil.NewCommandMap[usecase.DramahaInteractorIF]().
	Add(func(oi usecase.DramahaInteractorIF) string {
		return oi.Action(domain.DramahaActionFold, 0, 0)
	}, "f", "fold").
	Add(func(oi usecase.DramahaInteractorIF) string {
		return oi.Action(domain.DramahaActionCheck, 0, 0)
	}, "ck", "check").
	Add(func(oi usecase.DramahaInteractorIF) string {
		return oi.Action(domain.DramahaActionCall, 0, 0)
	}, "c", "call").
	Add(func(oi usecase.DramahaInteractorIF) string {
		return oi.Action(domain.DramahaActionAllIn, 0, 0)
	}, "a", "allin").
	Add(usecase.DramahaInteractorIF.Rebuy, "rb", "rebuy").
	Add(usecase.DramahaInteractorIF.SkipRebuy, "sr", "skiprebuy").
	Add(usecase.DramahaInteractorIF.Addon, "ad", "addon").
	Add(usecase.DramahaInteractorIF.SkipAddon, "sa", "skipaddon").
	Add(usecase.DramahaInteractorIF.Muck, "m", "muck").
	Add(usecase.DramahaInteractorIF.ShowHand, "sh", "show").
	Add(usecase.DramahaInteractorIF.ActionLog, "log", "l")

// dramahaArgfulCommands lists alias names for argful commands handled in the
// Exec switch.
var dramahaArgfulCommands = []string{
	"b", "bet", "ra", "raise",
	"bl", "bettinglimit", "tm", "tournament",
	"sb", "smallblind", "bb", "bigblind",
	"lh", "levelhand", "ts", "tablesize",
	"mai", "metaai",
	"d", "draw",
}

// DramahaCuiController ドラマハホールデムCUIコントローラークラス
type DramahaCuiController struct {
	oi usecase.DramahaInteractorIF
}

// NewDramahaCuiController コンストラクタ
func NewDramahaCuiController(oi usecase.DramahaInteractorIF) *DramahaCuiController {
	return &DramahaCuiController{oi: oi}
}

// Exec コマンド実行
func (c *DramahaCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.oi.Reset() },
		append(dramahaNoArgCommands.Names(), dramahaArgfulCommands...),
		func(cmd string, args []string) (string, bool) {
			if fn, ok := dramahaNoArgCommands.Lookup(cmd); ok {
				return fn(c.oi), true
			}
			switch cmd {
			case "d", "draw":
				// **番号は画面と同じ 1 始まり。** 0 始まりで打たせると、
				// 意図と 1 枚ずれた札が捨てられる。引数なしは「交換しない」。
				indices := make([]int, 0, len(args))
				for i := range args {
					n, errMsg, ok := cuiutil.ParseIntArgKeys(args[i:], "", "invalidCardIndex", 1, math.MaxInt)
					if !ok {
						return errMsg, true
					}
					indices = append(indices, n-1)
				}
				return c.oi.Draw(indices), true
			case "b", "bet":
				if len(args) < 1 {
					return cuiutil.PromptRequest(i18n.T("promptBetAmount"), "b {0}"), true
				}
				amount, err := parseAmount(args)
				if err != nil {
					return err.Error(), true
				}
				return c.oi.Action(domain.DramahaActionBet, amount, 0), true
			case "ra", "raise":
				if len(args) < 1 {
					return cuiutil.PromptRequest(i18n.T("promptRaiseAmount"), "ra {0}"), true
				}
				amount, err := parseAmount(args)
				if err != nil {
					return err.Error(), true
				}
				return c.oi.Action(domain.DramahaActionRaise, amount, 0), true
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
				// 収まらない席数を断る判断は DramahaInteractor が持つ
				// (CUI と Web API の両方がそこを通る)。ここでは渡すだけ。
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
