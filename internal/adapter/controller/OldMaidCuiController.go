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
				return cuiutil.PrependSkippedWarning(c.omi.Reorder(indices), skipped), true
			case "sm", "setmode":
				return cuiutil.WithParsedInt(args, "Game mode is required (0=Normal, 1=JijiNuki).", "Invalid game mode: %s. Please enter 0-1.", 0, 1, func(v int) string {
					cfg := c.omi.GetConfig()
					cfg.Mode = domain.OldMaidMode(v)
					return c.omi.Reset(cfg)
				})
			case "sps", "setplacementstrategy":
				return cuiutil.WithParsedInt(args, "CPU placement strategy flag is required (0=OFF, 1=ON).", "Invalid CPU placement strategy flag: %s. Please enter 0-1.", 0, 1, func(v int) string {
					cfg := c.omi.GetConfig()
					cfg.CpuPlacementStrategy = v == 1
					return c.omi.Reset(cfg)
				})
			case "smetaai", "smai":
				return cuiutil.WithParsedInt(args, "Meta-AI flag is required (0=OFF, 1=ON).", "Invalid meta-AI flag: %s. Please enter 0-1.", 0, 1, func(v int) string {
					cfg := c.omi.GetConfig()
					cfg.CpuMetaAI = v == 1
					return c.omi.Reset(cfg)
				})
			case "rp", "resetprofile":
				return c.omi.ResetProfile(), true
			case "sma", "setmemoryai":
				return cuiutil.WithParsedInt(args, "CPU memory AI flag is required (0=OFF, 1=ON).", "Invalid CPU memory AI flag: %s. Please enter 0-1.", 0, 1, func(v int) string {
					cfg := c.omi.GetConfig()
					cfg.CpuMemoryAI = v == 1
					return c.omi.Reset(cfg)
				})
			}
			return "", false
		},
	)
}
