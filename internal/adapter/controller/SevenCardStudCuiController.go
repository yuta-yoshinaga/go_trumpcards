package controller

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// SevenCardStudCuiController セブンカードスタッドCUIコントローラークラス
type SevenCardStudCuiController struct {
	si usecase.SevenCardStudInteractorIF
}

// NewSevenCardStudCuiController コンストラクタ
func NewSevenCardStudCuiController(si usecase.SevenCardStudInteractorIF) *SevenCardStudCuiController {
	return &SevenCardStudCuiController{si: si}
}

// Exec コマンド実行
func (c *SevenCardStudCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.si.Reset() },
		[]string{
			"f", "fold", "ck", "check", "c", "call", "b", "bet", "ra", "raise",
			"a", "allin", "bl", "bettinglimit", "tm", "tournament",
			"ante", "bi", "bringin", "sb", "smallbet", "bb", "bigbet",
			"lh", "levelhand", "ts", "tablesize",
			"rb", "rebuy", "sr", "skiprebuy", "ad", "addon", "sa", "skipaddon", "m", "muck", "sh", "show",
			"mai", "metaai",
			"log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "f", "fold":
				return c.si.Action(domain.SevenCardStudActionFold, 0, 0), true
			case "ck", "check":
				return c.si.Action(domain.SevenCardStudActionCheck, 0, 0), true
			case "c", "call":
				return c.si.Action(domain.SevenCardStudActionCall, 0, 0), true
			case "b", "bet":
				if len(args) < 1 {
					return cuiutil.PromptRequest(i18n.T("promptBetAmount"), "b {0}"), true
				}
				amount, err := parseAmount(args)
				if err != nil {
					return err.Error(), true
				}
				return c.si.Action(domain.SevenCardStudActionBet, amount, 0), true
			case "ra", "raise":
				if len(args) < 1 {
					return cuiutil.PromptRequest(i18n.T("promptRaiseAmount"), "ra {0}"), true
				}
				amount, err := parseAmount(args)
				if err != nil {
					return err.Error(), true
				}
				return c.si.Action(domain.SevenCardStudActionRaise, amount, 0), true
			case "a", "allin":
				return c.si.Action(domain.SevenCardStudActionAllIn, 0, 0), true
			case "bl", "bettinglimit":
				if len(args) < 1 {
					return i18n.T("holdem.bettingLimitRequired"), true
				}
				bl, err := strconv.Atoi(args[0])
				if err != nil {
					return i18n.Tf("holdem.invalidBettingLimit", "val", args[0]), true
				}
				cfg := c.si.GetConfig()
				cfg.BettingLimit = domain.BettingLimitType(bl)
				return c.si.ResetWithConfig(cfg, nil), true
			case "tm", "tournament":
				if len(args) < 1 {
					return i18n.T("holdem.tournamentModeRequired"), true
				}
				v, err := strconv.Atoi(args[0])
				if err != nil || v < 0 || v > 1 {
					return i18n.Tf("holdem.invalidTournamentMode", "val", args[0]), true
				}
				cfg := c.si.GetConfig()
				cfg.TournamentMode = v == 1
				return c.si.ResetWithConfig(cfg, nil), true
			case "ante":
				if len(args) < 1 {
					return i18n.T("sevencardstud.anteRequired"), true
				}
				v, err := strconv.Atoi(args[0])
				if err != nil {
					return i18n.Tf("sevencardstud.invalidAnte", "val", args[0]), true
				}
				cfg := c.si.GetConfig()
				cfg.Ante = v
				return c.si.ResetWithConfig(cfg, nil), true
			case "bi", "bringin":
				if len(args) < 1 {
					return i18n.T("sevencardstud.bringInRequired"), true
				}
				v, err := strconv.Atoi(args[0])
				if err != nil {
					return i18n.Tf("sevencardstud.invalidBringIn", "val", args[0]), true
				}
				cfg := c.si.GetConfig()
				cfg.BringIn = v
				return c.si.ResetWithConfig(cfg, nil), true
			case "sb", "smallbet":
				if len(args) < 1 {
					return i18n.T("sevencardstud.smallBetRequired"), true
				}
				v, err := strconv.Atoi(args[0])
				if err != nil {
					return i18n.Tf("sevencardstud.invalidSmallBet", "val", args[0]), true
				}
				cfg := c.si.GetConfig()
				cfg.SmallBet = v
				return c.si.ResetWithConfig(cfg, nil), true
			case "bb", "bigbet":
				if len(args) < 1 {
					return i18n.T("sevencardstud.bigBetRequired"), true
				}
				v, err := strconv.Atoi(args[0])
				if err != nil {
					return i18n.Tf("sevencardstud.invalidBigBet", "val", args[0]), true
				}
				cfg := c.si.GetConfig()
				cfg.BigBet = v
				return c.si.ResetWithConfig(cfg, nil), true
			case "lh", "levelhand":
				if len(args) < 1 {
					return i18n.T("holdem.levelHandRequired"), true
				}
				v, err := strconv.Atoi(args[0])
				if err != nil {
					return i18n.Tf("holdem.invalidLevelHand", "val", args[0]), true
				}
				cfg := c.si.GetConfig()
				cfg.AnteLevelHands = v
				return c.si.ResetWithConfig(cfg, nil), true
			case "ts", "tablesize":
				if len(args) < 1 {
					return i18n.T("holdem.tableSizeRequired"), true
				}
				v, err := strconv.Atoi(args[0])
				if err != nil {
					return i18n.Tf("holdem.invalidTableSize", "val", args[0]), true
				}
				cfg := c.si.GetConfig()
				cfg.TableSize = v
				return c.si.ResetWithConfig(cfg, nil), true
			case "rb", "rebuy":
				return c.si.Rebuy(), true
			case "sr", "skiprebuy":
				return c.si.SkipRebuy(), true
			case "ad", "addon":
				return c.si.Addon(), true
			case "sa", "skipaddon":
				return c.si.SkipAddon(), true
			case "m", "muck":
				return c.si.Muck(), true
			case "sh", "show":
				return c.si.ShowHand(), true
			case "mai", "metaai":
				if len(args) < 1 {
					return i18n.T("metaAIRequired"), true
				}
				v, err := strconv.Atoi(args[0])
				if err != nil || v < 0 || v > 1 {
					return i18n.Tf("invalidMetaAI", "val", args[0]), true
				}
				cfg := c.si.GetConfig()
				cfg.CpuMetaAI = v == 1
				return c.si.ResetWithConfig(cfg, nil), true
			default:
				return handleCuiLog(cmd, c.si.ActionLog)
			}
		},
	)
}
