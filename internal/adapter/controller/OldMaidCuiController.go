package controller

import (
	"fmt"
	"strconv"

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
		unknownCommandMessage,
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "d", "draw":
				cardIdx := -1
				if len(args) >= 1 {
					if idx, err := strconv.Atoi(args[0]); err == nil {
						cardIdx = idx
					}
				}
				return c.omi.Draw(cardIdx), true
			case "s", "shuffle":
				return c.omi.Shuffle(), true
			case "ro", "reorder":
				indices := []int{}
				for _, p := range args {
					idx, err := strconv.Atoi(p)
					if err == nil {
						indices = append(indices, idx)
					}
				}
				return c.omi.Reorder(indices), true
			case "sm", "setmode":
				if len(args) < 1 {
					return "Game mode is required (0=Normal, 1=JijiNuki).", true
				}
				m, err := strconv.Atoi(args[0])
				if err != nil || m < 0 || m > 1 {
					return fmt.Sprintf("Invalid game mode: %s. Please enter 0-1.", args[0]), true
				}
				cfg := c.omi.GetConfig()
				cfg.Mode = domain.OldMaidMode(m)
				return c.omi.Reset(cfg), true
			case "sps", "setplacementstrategy":
				if len(args) < 1 {
					return "CPU placement strategy flag is required (0=OFF, 1=ON).", true
				}
				v, err := strconv.Atoi(args[0])
				if err != nil || v < 0 || v > 1 {
					return fmt.Sprintf("Invalid CPU placement strategy flag: %s. Please enter 0-1.", args[0]), true
				}
				cfg := c.omi.GetConfig()
				cfg.CpuPlacementStrategy = v == 1
				return c.omi.Reset(cfg), true
			case "sma", "setmemoryai":
				if len(args) < 1 {
					return "CPU memory AI flag is required (0=OFF, 1=ON).", true
				}
				v, err := strconv.Atoi(args[0])
				if err != nil || v < 0 || v > 1 {
					return fmt.Sprintf("Invalid CPU memory AI flag: %s. Please enter 0-1.", args[0]), true
				}
				cfg := c.omi.GetConfig()
				cfg.CpuMemoryAI = v == 1
				return c.omi.Reset(cfg), true
			}
			return "", false
		},
	)
}
