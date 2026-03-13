package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// OldMaidCuiController ババ抜きCUIコントローラークラス
type OldMaidCuiController struct {
	omi usecase.OldMaidInteractorIF
}

// NewOldMaidCuiController コンストラクタ
func NewOldMaidCuiController(omi usecase.OldMaidInteractorIF) *OldMaidCuiController {
	return &OldMaidCuiController{omi: omi}
}

// Exec コマンド実行
// draw コマンドは "d N" または "draw N" の形式でカードインデックスを指定可能。
// 例: "d 2" → インデックス2のカードを引く / "d" → ランダムに引く
func (c *OldMaidCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.omi.Reset(c.omi.GetConfig()) },
		[]string{
			"d", "draw", "s", "shuffle", "ro", "reorder", "sm", "setmode",
			"sps", "setplacementstrategy", "smetaai", "smai",
			"rp", "resetprofile", "sma", "setmemoryai",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "d", "draw":
				return c.omi.Draw(cuiutil.ParseOptionalInt(args, 0, -1)), true
			case "s", "shuffle":
				return c.omi.Shuffle(), true
			case "ro", "reorder":
				indices, skipped := cuiutil.ParseIntSlice(args)
				result := c.omi.Reorder(indices)
				if w := cuiutil.FormatSkippedWarning(skipped); w != "" {
					result = w + "\n" + result
				}
				return result, true
			case "sm", "setmode":
				m, errMsg, ok := cuiutil.ParseIntArg(args, "Game mode is required (0=Normal, 1=JijiNuki).", "Invalid game mode: %s. Please enter 0-1.", 0, 1)
				if !ok {
					return errMsg, true
				}
				cfg := c.omi.GetConfig()
				cfg.Mode = domain.OldMaidMode(m)
				return c.omi.Reset(cfg), true
			case "sps", "setplacementstrategy":
				v, errMsg, ok := cuiutil.ParseIntArg(args, "CPU placement strategy flag is required (0=OFF, 1=ON).", "Invalid CPU placement strategy flag: %s. Please enter 0-1.", 0, 1)
				if !ok {
					return errMsg, true
				}
				cfg := c.omi.GetConfig()
				cfg.CpuPlacementStrategy = v == 1
				return c.omi.Reset(cfg), true
			case "smetaai", "smai":
				v, errMsg, ok := cuiutil.ParseIntArg(args, "Meta-AI flag is required (0=OFF, 1=ON).", "Invalid meta-AI flag: %s. Please enter 0-1.", 0, 1)
				if !ok {
					return errMsg, true
				}
				cfg := c.omi.GetConfig()
				cfg.CpuMetaAI = v == 1
				return c.omi.Reset(cfg), true
			case "rp", "resetprofile":
				return c.omi.ResetProfile(), true
			case "sma", "setmemoryai":
				v, errMsg, ok := cuiutil.ParseIntArg(args, "CPU memory AI flag is required (0=OFF, 1=ON).", "Invalid CPU memory AI flag: %s. Please enter 0-1.", 0, 1)
				if !ok {
					return errMsg, true
				}
				cfg := c.omi.GetConfig()
				cfg.CpuMemoryAI = v == 1
				return c.omi.Reset(cfg), true
			}
			return "", false
		},
	)
}
