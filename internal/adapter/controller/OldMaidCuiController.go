package controller

import (
	"strconv"
	"strings"

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
		func(_ []string) string { return c.omi.Reset(c.omi.GetConfig(), nil) },
		[]string{
			"d", "draw", "s", "shuffle", "ro", "reorder", "sm", "setmode",
			"sps", "setplacementstrategy", "smetaai", "smai",
			"rp", "resetprofile", "sma", "setmemoryai",
			"log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "d", "draw":
				// No argument draws at random; a typed one picks which of the next
				// player's cards to take, so it must not be silently read as -1
				// (issue #5390) -- that drew a card the player did not choose in
				// 17 of 30 deals.
				idx := -1
				if len(args) > 0 {
					v, err := strconv.Atoi(args[0])
					if err != nil {
						return invalidArg("invalidCardIndex", "val", args[0]), true
					}
					idx = v
				}
				return c.omi.Draw(idx), true
			case "s", "shuffle":
				return c.omi.Shuffle(), true
			case "ro", "reorder":
				indices, skipped := cuiutil.ParseIntSlice(args)
				// Refuse before playing. PrependSkippedWarning ran the move first and
				// put the warning above the new board, so a mistyped index was dropped
				// and the remaining ones played as a different, legal move (issue #5390).
				if len(skipped) > 0 {
					return invalidArg("invalidCardIndex", "val", strings.Join(skipped, ", ")), true
				}
				return c.omi.Reorder(indices), true
			case "sm", "setmode":
				return cuiutil.WithParsedInt(args, "Game mode is required (0=Normal, 1=JijiNuki).", "Invalid game mode: %s. Please enter 0-1.", 0, 1, func(v int) string {
					cfg := c.omi.GetConfig()
					cfg.Mode = domain.OldMaidMode(v)
					return c.omi.Reset(cfg, nil)
				})
			case "sps", "setplacementstrategy":
				return cuiutil.WithParsedInt(args, "CPU placement strategy flag is required (0=OFF, 1=ON).", "Invalid CPU placement strategy flag: %s. Please enter 0-1.", 0, 1, func(v int) string {
					cfg := c.omi.GetConfig()
					cfg.CpuPlacementStrategy = v == 1
					return c.omi.Reset(cfg, nil)
				})
			case "smetaai", "smai":
				return cuiutil.WithParsedInt(args, "Meta-AI flag is required (0=OFF, 1=ON).", "Invalid meta-AI flag: %s. Please enter 0-1.", 0, 1, func(v int) string {
					cfg := c.omi.GetConfig()
					cfg.CpuMetaAI = v == 1
					return c.omi.Reset(cfg, nil)
				})
			case "rp", "resetprofile":
				return c.omi.ResetProfile(), true
			case "sma", "setmemoryai":
				return cuiutil.WithParsedInt(args, "CPU memory AI flag is required (0=OFF, 1=ON).", "Invalid CPU memory AI flag: %s. Please enter 0-1.", 0, 1, func(v int) string {
					cfg := c.omi.GetConfig()
					cfg.CpuMemoryAI = v == 1
					return c.omi.Reset(cfg, nil)
				})
			default:
				return handleCuiLog(cmd, c.omi.ActionLog)
			}
		},
	)
}
