package controller

import (
	"errors"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
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
		[]string{
			"f", "fold", "ck", "check", "c", "call", "b", "bet", "ra", "raise",
			"a", "allin", "bl", "bettinglimit", "tm", "tournament",
			"sb", "smallblind", "bb", "bigblind", "lh", "levelhand", "ts", "tablesize",
			"rb", "rebuy", "sr", "skiprebuy", "ad", "addon", "sa", "skipaddon", "m", "muck", "sh", "show",
			"mai", "metaai",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "f", "fold":
				return c.hi.Action(domain.HoldemActionFold, 0, 0), true
			case "ck", "check":
				return c.hi.Action(domain.HoldemActionCheck, 0, 0), true
			case "c", "call":
				return c.hi.Action(domain.HoldemActionCall, 0, 0), true
			case "b", "bet":
				amount, err := parseAmount(args)
				if err != nil {
					return err.Error(), true
				}
				return c.hi.Action(domain.HoldemActionBet, amount, 0), true
			case "ra", "raise":
				amount, err := parseAmount(args)
				if err != nil {
					return err.Error(), true
				}
				return c.hi.Action(domain.HoldemActionRaise, amount, 0), true
			case "a", "allin":
				return c.hi.Action(domain.HoldemActionAllIn, 0, 0), true
			case "bl", "bettinglimit":
				if len(args) < 1 {
					return i18n.T("holdem.bettingLimitRequired"), true
				}
				bl, err := strconv.Atoi(args[0])
				if err != nil {
					return i18n.Tf("holdem.invalidBettingLimit", "val", args[0]), true
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
					return i18n.Tf("holdem.invalidTournamentMode", "val", args[0]), true
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
					return i18n.Tf("holdem.invalidSmallBlind", "val", args[0]), true
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
					return i18n.Tf("holdem.invalidBigBlind", "val", args[0]), true
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
					return i18n.Tf("holdem.invalidLevelHand", "val", args[0]), true
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
					return i18n.Tf("holdem.invalidTableSize", "val", args[0]), true
				}
				cfg := c.hi.GetConfig()
				cfg.TableSize = v
				return c.hi.ResetWithConfig(cfg, nil), true
			case "rb", "rebuy":
				return c.hi.Rebuy(), true
			case "sr", "skiprebuy":
				return c.hi.SkipRebuy(), true
			case "ad", "addon":
				return c.hi.Addon(), true
			case "sa", "skipaddon":
				return c.hi.SkipAddon(), true
			case "m", "muck":
				return c.hi.Muck(), true
			case "sh", "show":
				return c.hi.ShowHand(), true
			case "mai", "metaai":
				if len(args) < 1 {
					return i18n.T("metaAIRequired"), true
				}
				v, err := strconv.Atoi(args[0])
				if err != nil || v < 0 || v > 1 {
					return i18n.Tf("invalidMetaAI", "val", args[0]), true
				}
				cfg := c.hi.GetConfig()
				cfg.CpuMetaAI = v == 1
				return c.hi.ResetWithConfig(cfg, nil), true
			}
			return "", false
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
		return 0, errors.New(i18n.Tf("holdem.invalidAmount", "val", args[0]))
	}
	return amount, nil
}
