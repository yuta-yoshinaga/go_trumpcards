package controller

import (
	"fmt"
	"strconv"

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

// daifugoRuleKeys setruleコマンドで使えるルールキー
var daifugoRuleKeys = map[string]func(*domain.DaifugoConfig, bool){
	"8cut":       func(c *domain.DaifugoConfig, v bool) { c.EightCutEnabled = v },
	"11back":     func(c *domain.DaifugoConfig, v bool) { c.ElevenBackEnabled = v },
	"seq":        func(c *domain.DaifugoConfig, v bool) { c.SequenceEnabled = v },
	"exchange":   func(c *domain.DaifugoConfig, v bool) { c.CardExchangeEnabled = v },
	"5skip":      func(c *domain.DaifugoConfig, v bool) { c.FiveSkipEnabled = v },
	"7pass":      func(c *domain.DaifugoConfig, v bool) { c.SevenPassEnabled = v },
	"10discard":  func(c *domain.DaifugoConfig, v bool) { c.TenDiscardEnabled = v },
	"spade3":     func(c *domain.DaifugoConfig, v bool) { c.SpadeThreeEnabled = v },
	"capital":    func(c *domain.DaifugoConfig, v bool) { c.CapitalFallEnabled = v },
	"9reverse":   func(c *domain.DaifugoConfig, v bool) { c.NineReverseEnabled = v },
	"coupdetat":  func(c *domain.DaifugoConfig, v bool) { c.CoupDetatEnabled = v },
	"numberlock": func(c *domain.DaifugoConfig, v bool) { c.NumberLockEnabled = v },
	"sandstorm":  func(c *domain.DaifugoConfig, v bool) { c.SandstormEnabled = v },
	"emperor":    func(c *domain.DaifugoConfig, v bool) { c.EmperorEnabled = v },
	"seqrev":     func(c *domain.DaifugoConfig, v bool) { c.SequenceRevolutionEnabled = v },
	"illegal":    func(c *domain.DaifugoConfig, v bool) { c.IllegalFinishEnabled = v },
	"12bomber":   func(c *domain.DaifugoConfig, v bool) { c.QueenBomberEnabled = v },
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
		unknownCommandMessage,
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "p", "play":
				indices := []int{}
				for _, f := range args {
					if idx, err := strconv.Atoi(f); err == nil {
						indices = append(indices, idx)
					}
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
				if len(args) < 1 {
					return "CPU difficulty is required (0=Normal, 1=Easy, 2=Hard).", true
				}
				v, err := strconv.Atoi(args[0])
				if err != nil || v < 0 || v > 2 {
					return fmt.Sprintf("Invalid CPU difficulty: %s. Please enter 0-2.", args[0]), true
				}
				cfg := c.dgi.GetConfig()
				cfg.CpuDifficulty = domain.DaifugoCpuDifficulty(v)
				return c.dgi.ResetWithConfig(cfg), true
			case "sj", "setjoker":
				if len(args) < 1 {
					return "Joker count is required (0-2).", true
				}
				v, err := strconv.Atoi(args[0])
				if err != nil || v < 0 || v > 2 {
					return fmt.Sprintf("Invalid joker count: %s. Please enter 0-2.", args[0]), true
				}
				cfg := c.dgi.GetConfig()
				cfg.JokerCount = v
				return c.dgi.ResetWithConfig(cfg), true
			case "sr", "setrule":
				if len(args) < 2 {
					return "Usage: sr <rule> <0|1>. Rules: 8cut, 11back, seq, exchange, 5skip, 7pass, 10discard, spade3, capital, 9reverse, coupdetat, numberlock, sandstorm, emperor, seqrev, illegal, 12bomber. Use 'suitlockmode' for suit lock (0-2), '5skipcount' for skip count.", true
				}
				setter, ok := daifugoRuleKeys[args[0]]
				if !ok {
					return fmt.Sprintf("Unknown rule: %s.", args[0]), true
				}
				v, err := strconv.Atoi(args[1])
				if err != nil || v < 0 || v > 1 {
					return fmt.Sprintf("Invalid value: %s. Please enter 0 or 1.", args[1]), true
				}
				cfg := c.dgi.GetConfig()
				setter(&cfg, v == 1)
				return c.dgi.ResetWithConfig(cfg), true
			case "suitlockmode":
				if len(args) < 1 {
					return "Suit lock mode is required (0=none, 1=partial, 2=full).", true
				}
				v, err := strconv.Atoi(args[0])
				if err != nil || v < 0 || v > 2 {
					return fmt.Sprintf("Invalid suit lock mode: %s. Please enter 0-2.", args[0]), true
				}
				cfg := c.dgi.GetConfig()
				cfg.SuitLockMode = domain.DaifugoSuitLockMode(v)
				return c.dgi.ResetWithConfig(cfg), true
			case "5skipcount":
				if len(args) < 1 {
					return "Five skip count is required (1-10).", true
				}
				v, err := strconv.Atoi(args[0])
				if err != nil || v < 1 || v > 10 {
					return fmt.Sprintf("Invalid five skip count: %s. Please enter 1-10.", args[0]), true
				}
				cfg := c.dgi.GetConfig()
				cfg.FiveSkipCount = v
				return c.dgi.ResetWithConfig(cfg), true
			}
			return "", false
		},
	)
}
