package controller

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// DaifugoCuiController 大富豪CUIコントローラークラス
type DaifugoCuiController struct {
	dgi usecase.DaifugoInteractorIF
}

// NewDaifugoCuiController コンストラクタ
func NewDaifugoCuiController(dgi usecase.DaifugoInteractorIF) *DaifugoCuiController {
	return &DaifugoCuiController{dgi: dgi}
}

// setDaifugoRule sets a boolean rule flag on the config by key. Returns false if the key is unknown.
// cfg may be nil for key-only validation.
func setDaifugoRule(cfg *domain.DaifugoConfig, key string, value bool) bool {
	switch key {
	case "8cut", "11back", "seq", "exchange", "5skip", "7pass", "10discard",
		"spade3", "capital", "9reverse", "coupdetat", "numberlock", "sandstorm",
		"emperor", "seqrev", "seqlock", "illegal", "12bomber", "blindexchange":
		// valid key
	default:
		return false
	}
	if cfg == nil {
		return true
	}
	switch key {
	case "8cut":
		cfg.EightCutEnabled = value
	case "11back":
		cfg.ElevenBackEnabled = value
	case "seq":
		cfg.SequenceEnabled = value
	case "exchange":
		cfg.CardExchangeEnabled = value
	case "5skip":
		cfg.FiveSkipEnabled = value
	case "7pass":
		cfg.SevenPassEnabled = value
	case "10discard":
		cfg.TenDiscardEnabled = value
	case "spade3":
		cfg.SpadeThreeEnabled = value
	case "capital":
		cfg.CapitalFallEnabled = value
	case "9reverse":
		cfg.NineReverseEnabled = value
	case "coupdetat":
		cfg.CoupDetatEnabled = value
	case "numberlock":
		cfg.NumberLockEnabled = value
	case "sandstorm":
		cfg.SandstormEnabled = value
	case "emperor":
		cfg.EmperorEnabled = value
	case "seqrev":
		cfg.SequenceRevolutionEnabled = value
	case "seqlock":
		cfg.SequenceLockEnabled = value
	case "illegal":
		cfg.IllegalFinishEnabled = value
	case "12bomber":
		cfg.QueenBomberEnabled = value
	case "blindexchange":
		cfg.BlindExchangeEnabled = value
	}
	return true
}

// Exec コマンド実行
// play コマンドは "p 0 2" または "play 0 2" の形式でカードインデックスを指定。
// インデックスなしの場合はパス扱い。例: "p" → パス / "p 0 2" → 0番と2番のカードを出す
func (c *DaifugoCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.dgi.GetConfig()
			return c.dgi.ResetWithConfig(cfg)
		},
		[]string{
			"p", "play", "sort", "sd", "setdifficulty", "sj", "setjoker",
			"sr", "setrule", "suitlockmode", "5skipcount",
			"log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "p", "play":
				indices, skipped := cuiutil.ParseIntSlice(args)
				return cuiutil.PrependSkippedWarning(c.dgi.Play(indices), skipped), true
			case "sort":
				mode := domain.DaifugoSortByStrength
				if len(args) > 0 {
					if m, err := strconv.Atoi(args[0]); err == nil && m >= int(domain.DaifugoSortByStrength) && m <= int(domain.DaifugoSortByNumber) {
						mode = domain.DaifugoSortMode(m)
					}
				}
				return c.dgi.Sort(mode), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedInt(args, "CPU difficulty is required (0=Normal, 1=Easy, 2=Hard).", "Invalid CPU difficulty: %s. Please enter 0-2.", 0, 2, func(v int) string {
					cfg := c.dgi.GetConfig()
					cfg.CpuDifficulty = domain.DaifugoCpuDifficulty(v)
					return c.dgi.ResetWithConfig(cfg)
				})
			case "sj", "setjoker":
				return cuiutil.WithParsedInt(args, "Joker count is required (0-2).", "Invalid joker count: %s. Please enter 0-2.", 0, 2, func(v int) string {
					cfg := c.dgi.GetConfig()
					cfg.JokerCount = v
					return c.dgi.ResetWithConfig(cfg)
				})
			case "sr", "setrule":
				if len(args) < 2 {
					return "Usage: sr <rule> <0|1>. Rules: 8cut, 11back, seq, exchange, 5skip, 7pass, 10discard, spade3, capital, 9reverse, coupdetat, numberlock, sandstorm, emperor, seqrev, seqlock, illegal, 12bomber, blindexchange. Use 'suitlockmode' for suit lock (0-2), '5skipcount' for skip count.", true
				}
				if !setDaifugoRule(nil, args[0], false) {
					return fmt.Sprintf("Unknown rule: %s.", args[0]), true
				}
				v, err := strconv.Atoi(args[1])
				if err != nil || v < 0 || v > 1 {
					return fmt.Sprintf("Invalid value: %s. Please enter 0 or 1.", args[1]), true
				}
				cfg := c.dgi.GetConfig()
				setDaifugoRule(&cfg, args[0], v == 1)
				return c.dgi.ResetWithConfig(cfg), true
			case "suitlockmode":
				return cuiutil.WithParsedInt(args, "Suit lock mode is required (0=none, 1=partial, 2=full).", "Invalid suit lock mode: %s. Please enter 0-2.", 0, 2, func(v int) string {
					cfg := c.dgi.GetConfig()
					cfg.SuitLockMode = domain.DaifugoSuitLockMode(v)
					return c.dgi.ResetWithConfig(cfg)
				})
			case "5skipcount":
				return cuiutil.WithParsedInt(args, "Five skip count is required (1-5).", "Invalid five skip count: %s. Please enter 1-5.", 1, 5, func(v int) string {
					cfg := c.dgi.GetConfig()
					cfg.FiveSkipCount = v
					return c.dgi.ResetWithConfig(cfg)
				})
			default:
				return handleCuiLog(cmd, c.dgi.ActionLog)
			}
		},
	)
}
