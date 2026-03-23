package controller

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

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
				return c.oi.Action(domain.OmahaActionFold, 0, 0), true
			case "ck", "check":
				return c.oi.Action(domain.OmahaActionCheck, 0, 0), true
			case "c", "call":
				return c.oi.Action(domain.OmahaActionCall, 0, 0), true
			case "b", "bet":
				amount, err := parseAmount(args)
				if err != nil {
					return err.Error(), true
				}
				return c.oi.Action(domain.OmahaActionBet, amount, 0), true
			case "ra", "raise":
				amount, err := parseAmount(args)
				if err != nil {
					return err.Error(), true
				}
				return c.oi.Action(domain.OmahaActionRaise, amount, 0), true
			case "a", "allin":
				return c.oi.Action(domain.OmahaActionAllIn, 0, 0), true
			case "bl", "bettinglimit":
				if len(args) < 1 {
					return i18n.T("holdem.bettingLimitRequired"), true
				}
				bl, err := strconv.Atoi(args[0])
				if err != nil {
					return i18n.Tf("holdem.invalidBettingLimit", "val", args[0]), true
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
					return i18n.Tf("holdem.invalidTournamentMode", "val", args[0]), true
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
					return i18n.Tf("holdem.invalidSmallBlind", "val", args[0]), true
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
					return i18n.Tf("holdem.invalidBigBlind", "val", args[0]), true
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
					return i18n.Tf("holdem.invalidLevelHand", "val", args[0]), true
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
					return i18n.Tf("holdem.invalidTableSize", "val", args[0]), true
				}
				cfg := c.oi.GetConfig()
				cfg.TableSize = v
				return c.oi.ResetWithConfig(cfg, nil), true
			case "rb", "rebuy":
				return c.oi.Rebuy(), true
			case "sr", "skiprebuy":
				return c.oi.SkipRebuy(), true
			case "ad", "addon":
				return c.oi.Addon(), true
			case "sa", "skipaddon":
				return c.oi.SkipAddon(), true
			case "m", "muck":
				return c.oi.Muck(), true
			case "sh", "show":
				return c.oi.ShowHand(), true
			case "mai", "metaai":
				if len(args) < 1 {
					return i18n.T("metaAIRequired"), true
				}
				v, err := strconv.Atoi(args[0])
				if err != nil || v < 0 || v > 1 {
					return i18n.Tf("invalidMetaAI", "val", args[0]), true
				}
				cfg := c.oi.GetConfig()
				cfg.CpuMetaAI = v == 1
				return c.oi.ResetWithConfig(cfg, nil), true
			}
			return "", false
		},
	)
}
