//go:build !js || !wasm || casino

package controller

import (
	"errors"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// IndianPokerCuiController インディアンポーカーCUIコントローラークラス
type IndianPokerCuiController struct {
	ipi usecase.IndianPokerInteractorIF
}

// NewIndianPokerCuiController コンストラクタ
func NewIndianPokerCuiController(ipi usecase.IndianPokerInteractorIF) *IndianPokerCuiController {
	return &IndianPokerCuiController{ipi: ipi}
}

// Exec コマンド実行
func (c *IndianPokerCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.ipi.Reset() },
		[]string{
			"f", "fold", "ck", "check", "c", "call", "b", "bet", "ra", "raise",
			"a", "allin", "bl", "bettinglimit", "mai", "metaai", "an", "ante",
			"log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "f", "fold":
				return c.ipi.Action(domain.IndianPokerActionFold, 0, 0), true
			case "ck", "check":
				return c.ipi.Action(domain.IndianPokerActionCheck, 0, 0), true
			case "c", "call":
				return c.ipi.Action(domain.IndianPokerActionCall, 0, 0), true
			case "b", "bet":
				amount, err := indianPokerParseAmount(args)
				if err != nil {
					return err.Error(), true
				}
				return c.ipi.Action(domain.IndianPokerActionBet, amount, 0), true
			case "ra", "raise":
				amount, err := indianPokerParseAmount(args)
				if err != nil {
					return err.Error(), true
				}
				return c.ipi.Action(domain.IndianPokerActionRaise, amount, 0), true
			case "a", "allin":
				return c.ipi.Action(domain.IndianPokerActionAllIn, 0, 0), true
			case "bl", "bettinglimit":
				if len(args) < 1 {
					return i18n.T("indianpoker.bettingLimitRequired"), true
				}
				bl, err := strconv.Atoi(args[0])
				if err != nil {
					return invalidArg("indianpoker.invalidBettingLimit", "val", args[0]), true
				}
				cfg := c.ipi.GetConfig()
				cfg.BettingLimit = domain.BettingLimitType(bl)
				return c.ipi.ResetWithConfig(cfg, nil), true
			case "mai", "metaai":
				if len(args) < 1 {
					return i18n.T("metaAIRequired"), true
				}
				v, err := strconv.Atoi(args[0])
				if err != nil || v < 0 || v > 1 {
					return invalidArg("invalidMetaAI", "val", args[0]), true
				}
				cfg := c.ipi.GetConfig()
				cfg.CpuMetaAI = v == 1
				return c.ipi.ResetWithConfig(cfg, nil), true
			case "an", "ante":
				if len(args) < 1 {
					return i18n.T("indianpoker.anteRequired"), true
				}
				v, err := strconv.Atoi(args[0])
				if err != nil || v < 1 {
					return invalidArg("indianpoker.invalidAnte", "val", args[0]), true
				}
				cfg := c.ipi.GetConfig()
				cfg.Ante = v
				return c.ipi.ResetWithConfig(cfg, nil), true
			default:
				return handleCuiLog(cmd, c.ipi.ActionLog)
			}
		},
	)
}

// indianPokerParseAmount 引数スライスからベット額を抽出する
func indianPokerParseAmount(args []string) (int, error) {
	if len(args) < 1 {
		return 0, errors.New(i18n.T("indianpoker.amountRequired"))
	}
	amount, err := strconv.Atoi(args[0])
	if err != nil || amount <= 0 {
		return 0, errors.New(invalidArg("indianpoker.invalidAmount", "val", args[0]))
	}
	return amount, nil
}
