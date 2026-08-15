//go:build !js || !wasm || casino

package controller

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// PokerCuiController ポーカーCUIコントローラークラス
type PokerCuiController struct {
	pi usecase.PokerInteractorIF
}

// NewPokerCuiController コンストラクタ
func NewPokerCuiController(pi usecase.PokerInteractorIF) *PokerCuiController {
	return &PokerCuiController{
		pi: pi,
	}
}

// Exec ゲーム実行
// コマンド例: "r", "e 0 2 4", "s", "b 20", "c", "ra 30", "f", "ck", "a", "q"
func (pcc *PokerCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return pcc.pi.Reset() },
		[]string{
			"e", "exchange", "s", "stand", "b", "bet", "c", "call", "ra", "raise",
			"f", "fold", "ck", "check", "a", "allin", "bl", "bettinglimit",
			"scc", "setcpucount", "sjc", "setjokercount", "o", "odds", "lw", "lowball",
			"mai", "metaai",
			"log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "e", "exchange":
				indices, skipped := cuiutil.ParseBoundedIntSlice(args, 0, 4)
				return cuiutil.PrependSkippedWarning(pcc.pi.Exchange(indices), skipped), true
			case "s", "stand":
				return pcc.pi.Stand(), true
			case "b", "bet":
				// No argument keeps the "bet zero" shorthand; a typed amount that does
				// not parse, or is not positive, is refused rather than silently turned
				// into zero -- the player asked to bet (issue #5390).
				amount := 0
				if len(args) > 0 {
					a, err := strconv.Atoi(args[0])
					if err != nil || a <= 0 {
						return invalidArg("invalidBetAmount", "val", args[0]), true
					}
					amount = a
				}
				return pcc.pi.Action(domain.PokerActionBet, amount, 0), true
			case "c", "call":
				return pcc.pi.Action(domain.PokerActionCall, 0, 0), true
			case "ra", "raise":
				// No argument keeps the "bet zero" shorthand; a typed amount that does
				// not parse, or is not positive, is refused rather than silently turned
				// into zero -- the player asked to bet (issue #5390).
				amount := 0
				if len(args) > 0 {
					a, err := strconv.Atoi(args[0])
					if err != nil || a <= 0 {
						return invalidArg("invalidBetAmount", "val", args[0]), true
					}
					amount = a
				}
				return pcc.pi.Action(domain.PokerActionRaise, amount, 0), true
			case "f", "fold":
				return pcc.pi.Action(domain.PokerActionFold, 0, 0), true
			case "ck", "check":
				return pcc.pi.Action(domain.PokerActionCheck, 0, 0), true
			case "a", "allin":
				return pcc.pi.Action(domain.PokerActionAllIn, 0, 0), true
			case "bl", "bettinglimit":
				return cuiutil.WithParsedInt(args, "Betting limit type is required (0=Fixed, 1=PotLimit, 2=NoLimit).", "Invalid betting limit: %s. Please enter 0-2.", 0, 2, func(v int) string {
					cfg := pcc.pi.GetConfig()
					cfg.BettingLimit = domain.BettingLimitType(v)
					return pcc.pi.ResetWithConfig(cfg, nil)
				})
			case "scc", "setcpucount":
				return cuiutil.WithParsedInt(args, "CPU player count is required.", "Invalid CPU player count: %s. Please enter 1-3.", 1, 3, func(v int) string {
					cfg := pcc.pi.GetConfig()
					cfg.CpuCount = v
					return pcc.pi.ResetWithConfig(cfg, nil)
				})
			case "sjc", "setjokercount":
				return cuiutil.WithParsedInt(args, "Joker count is required.", "Invalid joker count: %s. Please enter 0-2.", 0, 2, func(v int) string {
					cfg := pcc.pi.GetConfig()
					cfg.JokerCount = v
					return pcc.pi.ResetWithConfig(cfg, nil)
				})
			case "o", "odds":
				indices, skipped := cuiutil.ParseBoundedIntSlice(args, 0, 4)
				return cuiutil.PrependSkippedWarning(pcc.pi.Odds(indices), skipped), true
			case "lw", "lowball":
				cfg := pcc.pi.GetConfig()
				cfg.IsLowball = !cfg.IsLowball
				return pcc.pi.ResetWithConfig(cfg, nil), true
			case "mai", "metaai":
				if len(args) < 1 {
					return i18n.T("metaAIRequired"), true
				}
				v, err := strconv.Atoi(args[0])
				if err != nil || v < 0 || v > 1 {
					return invalidArg("invalidMetaAI", "val", args[0]), true
				}
				cfg := pcc.pi.GetConfig()
				cfg.CpuMetaAI = v == 1
				return pcc.pi.ResetWithConfig(cfg, nil), true
			default:
				return handleCuiLog(cmd, pcc.pi.ActionLog)
			}
		},
	)
}
