package controller

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuimsg"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
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
		unknownCommandMessage,
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "e", "exchange":
				indices := []int{}
				for _, p := range args {
					idx, err := strconv.Atoi(p)
					if err == nil && 0 <= idx && idx <= 4 {
						indices = append(indices, idx)
					}
				}
				return pcc.pi.Exchange(indices), true
			case "s", "stand":
				return pcc.pi.Stand(), true
			case "b", "bet":
				amount := 0
				if len(args) > 0 {
					if a, err := strconv.Atoi(args[0]); err == nil && a > 0 {
						amount = a
					}
				}
				return pcc.pi.Action(domain.PokerActionBet, amount), true
			case "c", "call":
				return pcc.pi.Action(domain.PokerActionCall, 0), true
			case "ra", "raise":
				amount := 0
				if len(args) > 0 {
					if a, err := strconv.Atoi(args[0]); err == nil && a > 0 {
						amount = a
					}
				}
				return pcc.pi.Action(domain.PokerActionRaise, amount), true
			case "f", "fold":
				return pcc.pi.Action(domain.PokerActionFold, 0), true
			case "ck", "check":
				return pcc.pi.Action(domain.PokerActionCheck, 0), true
			case "a", "allin":
				return pcc.pi.Action(domain.PokerActionAllIn, 0), true
			case "bl", "bettinglimit":
				if len(args) < 1 {
					return "Betting limit type is required (0=Fixed, 1=PotLimit, 2=NoLimit).", true
				}
				bl, err := strconv.Atoi(args[0])
				if err != nil || bl < 0 || bl > 2 {
					return cuimsg.InvalidOutOfRange("betting limit", args[0], "Please enter 0-2."), true
				}
				cfg := pcc.pi.GetConfig()
				cfg.BettingLimit = domain.BettingLimitType(bl)
				return pcc.pi.ResetWithConfig(cfg), true
			case "scc", "setcpucount":
				if len(args) < 1 {
					return cuimsg.Required("CPU player count"), true
				}
				count, err := strconv.Atoi(args[0])
				if err != nil || count < 1 || count > 3 {
					return cuimsg.InvalidOutOfRange("CPU player count", args[0], "Please enter 1-3."), true
				}
				cfg := pcc.pi.GetConfig()
				cfg.CpuCount = count
				return pcc.pi.ResetWithConfig(cfg), true
			case "sjc", "setjokercount":
				if len(args) < 1 {
					return cuimsg.Required("Joker count"), true
				}
				count, err := strconv.Atoi(args[0])
				if err != nil || count < 0 || count > 2 {
					return cuimsg.InvalidOutOfRange("joker count", args[0], "Please enter 0-2."), true
				}
				cfg := pcc.pi.GetConfig()
				cfg.JokerCount = count
				return pcc.pi.ResetWithConfig(cfg), true
			case "o", "odds":
				indices := []int{}
				for _, p := range args {
					idx, err := strconv.Atoi(p)
					if err == nil && 0 <= idx && idx <= 4 {
						indices = append(indices, idx)
					}
				}
				return pcc.pi.Odds(indices), true
			case "lw", "lowball":
				cfg := pcc.pi.GetConfig()
				cfg.IsLowball = !cfg.IsLowball
				return pcc.pi.ResetWithConfig(cfg), true
			}
			return "", false
		},
	)
}
