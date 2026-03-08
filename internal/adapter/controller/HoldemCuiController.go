package controller

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

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
		unknownCommandMessage,
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "f", "fold":
				return c.hi.Action(domain.HoldemActionFold, 0), true
			case "ck", "check":
				return c.hi.Action(domain.HoldemActionCheck, 0), true
			case "c", "call":
				return c.hi.Action(domain.HoldemActionCall, 0), true
			case "b", "bet":
				amount, err := parseAmount(args)
				if err != nil {
					return err.Error(), true
				}
				return c.hi.Action(domain.HoldemActionBet, amount), true
			case "ra", "raise":
				amount, err := parseAmount(args)
				if err != nil {
					return err.Error(), true
				}
				return c.hi.Action(domain.HoldemActionRaise, amount), true
			case "a", "allin":
				return c.hi.Action(domain.HoldemActionAllIn, 0), true
			case "bl", "bettinglimit":
				if len(args) < 1 {
					return "Betting limit type is required (0=Fixed, 1=PotLimit, 2=NoLimit).", true
				}
				bl, err := strconv.Atoi(args[0])
				if err != nil || bl < 0 || bl > 2 {
					return fmt.Sprintf("Invalid betting limit: %s. Please enter 0-2.", args[0]), true
				}
				cfg := c.hi.GetConfig()
				cfg.BettingLimit = domain.BettingLimitType(bl)
				return c.hi.ResetWithConfig(cfg), true
			case "tm", "tournament":
				if len(args) < 1 {
					return "Tournament mode is required (0=OFF, 1=ON).", true
				}
				v, err := strconv.Atoi(args[0])
				if err != nil || v < 0 || v > 1 {
					return fmt.Sprintf("Invalid tournament mode: %s. Please enter 0-1.", args[0]), true
				}
				cfg := c.hi.GetConfig()
				cfg.TournamentMode = v == 1
				return c.hi.ResetWithConfig(cfg), true
			case "sb", "smallblind":
				if len(args) < 1 {
					return "Small blind amount is required (>=1).", true
				}
				v, err := strconv.Atoi(args[0])
				if err != nil || v < 1 {
					return fmt.Sprintf("Invalid small blind: %s. Please enter 1 or more.", args[0]), true
				}
				cfg := c.hi.GetConfig()
				if v >= cfg.BigBlind {
					return fmt.Sprintf("Small blind must be less than big blind (%d).", cfg.BigBlind), true
				}
				cfg.SmallBlind = v
				return c.hi.ResetWithConfig(cfg), true
			case "bb", "bigblind":
				if len(args) < 1 {
					return "Big blind amount is required (>=2).", true
				}
				v, err := strconv.Atoi(args[0])
				if err != nil || v < 2 {
					return fmt.Sprintf("Invalid big blind: %s. Please enter 2 or more.", args[0]), true
				}
				cfg := c.hi.GetConfig()
				if v <= cfg.SmallBlind {
					return fmt.Sprintf("Big blind must be greater than small blind (%d).", cfg.SmallBlind), true
				}
				cfg.BigBlind = v
				return c.hi.ResetWithConfig(cfg), true
			case "lh", "levelhand":
				if len(args) < 1 {
					return "Level-up hands is required (>=1).", true
				}
				v, err := strconv.Atoi(args[0])
				if err != nil || v < 1 {
					return fmt.Sprintf("Invalid level-up hands: %s. Please enter 1 or more.", args[0]), true
				}
				cfg := c.hi.GetConfig()
				cfg.BlindLevelHands = v
				return c.hi.ResetWithConfig(cfg), true
			case "ts", "tablesize":
				if len(args) < 1 {
					return "Table size is required (4, 6, or 9).", true
				}
				v, err := strconv.Atoi(args[0])
				if err != nil || !domain.IsValidHoldemTableSize(v) {
					return fmt.Sprintf("Invalid table size: %s. Please enter 4, 6, or 9.", args[0]), true
				}
				cfg := c.hi.GetConfig()
				cfg.TableSize = v
				return c.hi.ResetWithConfig(cfg), true
			case "rb", "rebuy":
				return c.hi.Rebuy(), true
			case "sr", "skiprebuy":
				return c.hi.SkipRebuy(), true
			case "ad", "addon":
				return c.hi.Addon(), true
			case "sa", "skipaddon":
				return c.hi.SkipAddon(), true
			}
			return "", false
		},
	)
}

// parseAmount 引数スライスからベット額を抽出する
func parseAmount(args []string) (int, error) {
	if len(args) < 1 {
		return 0, fmt.Errorf("ベット/レイズには金額の指定が必要です。")
	}
	amount, err := strconv.Atoi(args[0])
	if err != nil || amount <= 0 {
		return 0, fmt.Errorf("無効な金額です: %s", args[0])
	}
	return amount, nil
}
