package controller

import (
	"fmt"
	"strconv"
	"strings"

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

// daifugoRuleKeys is the ordered list of all boolean rule keys for sr commands.
var daifugoRuleKeys = []string{
	"8cut", "11back", "seq", "exchange", "5skip", "7pass", "10discard",
	"spade3", "capital", "9reverse", "coupdetat", "numberlock", "sandstorm",
	"emperor", "seqrev", "seqlock", "illegal", "12bomber", "blindexchange",
}

// getDaifugoRule returns the current value of a boolean rule flag. Returns (false, false) if unknown.
func getDaifugoRule(cfg *domain.DaifugoConfig, key string) (value bool, ok bool) {
	switch key {
	case "8cut":
		return cfg.EightCutEnabled, true
	case "11back":
		return cfg.ElevenBackEnabled, true
	case "seq":
		return cfg.SequenceEnabled, true
	case "exchange":
		return cfg.CardExchangeEnabled, true
	case "5skip":
		return cfg.FiveSkipEnabled, true
	case "7pass":
		return cfg.SevenPassEnabled, true
	case "10discard":
		return cfg.TenDiscardEnabled, true
	case "spade3":
		return cfg.SpadeThreeEnabled, true
	case "capital":
		return cfg.CapitalFallEnabled, true
	case "9reverse":
		return cfg.NineReverseEnabled, true
	case "coupdetat":
		return cfg.CoupDetatEnabled, true
	case "numberlock":
		return cfg.NumberLockEnabled, true
	case "sandstorm":
		return cfg.SandstormEnabled, true
	case "emperor":
		return cfg.EmperorEnabled, true
	case "seqrev":
		return cfg.SequenceRevolutionEnabled, true
	case "seqlock":
		return cfg.SequenceLockEnabled, true
	case "illegal":
		return cfg.IllegalFinishEnabled, true
	case "12bomber":
		return cfg.QueenBomberEnabled, true
	case "blindexchange":
		return cfg.BlindExchangeEnabled, true
	default:
		return false, false
	}
}

// formatDaifugoRuleList returns a formatted string showing all rules with their current ON/OFF status.
func formatDaifugoRuleList(cfg *domain.DaifugoConfig) string {
	var b strings.Builder
	for _, key := range daifugoRuleKeys {
		val, _ := getDaifugoRule(cfg, key)
		status := "OFF"
		if val {
			status = "ON"
		}
		fmt.Fprintf(&b, "  %-16s %s\n", key, status)
	}
	return b.String()
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
				// Refuse before playing. PrependSkippedWarning ran the move first and
				// put the warning above the new board, so a mistyped index was dropped
				// and the remaining ones played as a different, legal move (issue #5390).
				if len(skipped) > 0 {
					return invalidArg("invalidCardIndex", "val", strings.Join(skipped, ", ")), true
				}
				return c.dgi.Play(indices), true
			case "sort":
				mode := domain.DaifugoSortByStrength
				if len(args) > 0 {
					if m, err := strconv.Atoi(args[0]); err == nil && m >= int(domain.DaifugoSortByStrength) && m <= int(domain.DaifugoSortByNumber) {
						mode = domain.DaifugoSortMode(m)
					}
				}
				return c.dgi.Sort(mode), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequiredAlt", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.dgi.GetConfig()
					cfg.CpuDifficulty = domain.DaifugoCpuDifficulty(v)
					return c.dgi.ResetWithConfig(cfg)
				})
			case "sj", "setjoker":
				return cuiutil.WithParsedIntKeys(args, "jokerCountRequired02", "invalidJokerCount02", 0, 2, func(v int) string {
					cfg := c.dgi.GetConfig()
					cfg.JokerCount = v
					return c.dgi.ResetWithConfig(cfg)
				})
			case "sr", "setrule":
				if len(args) >= 1 && args[0] == "list" {
					cfg := c.dgi.GetConfig()
					return "Rules:\n" + formatDaifugoRuleList(&cfg), true
				}
				if len(args) < 2 {
					return "Usage: sr <rule> <0|1> | sr list\nRules: " + strings.Join(daifugoRuleKeys, ", ") + ".\nUse 'suitlockmode' for suit lock (0-2), '5skipcount' for skip count.", true
				}
				if !setDaifugoRule(nil, args[0], false) {
					return invalidArg("unknownRule", "val", fmt.Sprint(args[0])), true
				}
				v, err := strconv.Atoi(args[1])
				if err != nil || v < 0 || v > 1 {
					return invalidArg("invalidValue0Or1Raw", "val", fmt.Sprint(args[1])), true
				}
				cfg := c.dgi.GetConfig()
				setDaifugoRule(&cfg, args[0], v == 1)
				return c.dgi.ResetWithConfig(cfg), true
			case "suitlockmode":
				return cuiutil.WithParsedIntKeys(args, "suitLockModeRequired", "invalidSuitLockMode", 0, 2, func(v int) string {
					cfg := c.dgi.GetConfig()
					cfg.SuitLockMode = domain.DaifugoSuitLockMode(v)
					return c.dgi.ResetWithConfig(cfg)
				})
			case "5skipcount":
				return cuiutil.WithParsedIntKeys(args, "fiveSkipCountRequired15", "invalidFiveSkipCount15", 1, 5, func(v int) string {
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
