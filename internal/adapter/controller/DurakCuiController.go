package controller

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// DurakCuiController ドゥラークCUIコントローラークラス
type DurakCuiController struct {
	di usecase.DurakInteractorIF
}

// NewDurakCuiController コンストラクタ
func NewDurakCuiController(di usecase.DurakInteractorIF) *DurakCuiController {
	return &DurakCuiController{di: di}
}

// Exec コマンド実行
// a <idx>: 攻撃, d <atkIdx> <handIdx>: 防御, p: パス, t: 引き取り
func (c *DurakCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.di.GetConfig()
			return c.di.ResetWithConfig(cfg)
		},
		[]string{
			"a", "attack", "d", "defend", "p", "pass", "t", "take",
			"sort", "sd", "setdifficulty", "h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "a", "attack":
				// No argument keeps the index-0 shorthand; a typed argument that does
				// not parse is refused rather than silently read as 0, which played a
				// card the player never chose (issue #5390).
				idx := 0
				if len(args) > 0 {
					v, err := strconv.Atoi(args[0])
					if err != nil {
						return invalidArg("invalidCardIndex", "val", args[0]), true
					}
					idx = v
				}
				return c.di.Attack(idx), true
			case "d", "defend":
				atkIdx := 0
				handIdx := 0
				if len(args) > 0 {
					if v, err := strconv.Atoi(args[0]); err == nil {
						atkIdx = v
					}
				}
				if len(args) > 1 {
					if v, err := strconv.Atoi(args[1]); err == nil {
						handIdx = v
					}
				}
				return c.di.Defend(atkIdx, handIdx), true
			case "p", "pass":
				return c.di.Pass(), true
			case "t", "take":
				return c.di.TakeCards(), true
			case "sort":
				mode := domain.DurakSortBySuit
				if len(args) > 0 {
					if m, err := strconv.Atoi(args[0]); err == nil && m >= int(domain.DurakSortBySuit) && m <= int(domain.DurakSortByValue) {
						mode = domain.DurakSortMode(m)
					}
				}
				return c.di.Sort(mode), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequiredAlt", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.di.GetConfig()
					cfg.CpuDifficulty = domain.DurakCpuDifficulty(v)
					return c.di.ResetWithConfig(cfg)
				})
			default:
				return handleCuiHintAndLog(cmd, c.di.Hint, c.di.ActionLog)
			}
		},
	)
}
