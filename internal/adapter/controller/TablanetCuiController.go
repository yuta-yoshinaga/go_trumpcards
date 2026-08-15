//go:build !js || !wasm || extra3

package controller

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// TablanetCuiController はタブラネット (Tablanet) の CUI コントローラークラス。
type TablanetCuiController struct {
	bi usecase.TablanetInteractorIF
}

// NewTablanetCuiController コンストラクタ。
func NewTablanetCuiController(bi usecase.TablanetInteractorIF) *TablanetCuiController {
	return &TablanetCuiController{bi: bi}
}

// Exec コマンド実行。
//
//	q / quit                 → ゲーム終了 ("bye.")
//	r / reset                → ゲームリセット (設定保持)
//	p / play <h> [t...]      → 手札 h を出す。t... は捕獲する場札インデックス
//	                           (省略時: ジャックは一掃、それ以外はトレイル)
//	nr / nextround / n       → 次のゲームへ
//	sd / setdifficulty <0-2> → CPU難易度設定
//	h / hint                 → ヒント表示
//	log / l                  → 棋譜表示
func (c *TablanetCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.bi.GetConfig()
			return c.bi.ResetWithConfig(cfg)
		},
		[]string{
			"p", "play", "n", "next", "nr", "nextround",
			"sd", "setdifficulty", "h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "p", "play":
				return c.handlePlay(args), true
			case "n", "next", "nr", "nextround":
				return c.bi.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.bi.GetConfig()
					cfg.CpuDifficulty = domain.TablanetCpuDifficulty(v)
					return c.bi.ResetWithConfig(cfg)
				})
			default:
				return handleCuiHintAndLog(cmd, c.bi.Hint, c.bi.ActionLog)
			}
		},
	)
}

// handlePlay は "p <handIdx> [tableIdx...]" を解析して Play を呼ぶ。
func (c *TablanetCuiController) handlePlay(args []string) string {
	if len(args) == 0 {
		return i18n.T("cardIndexRequiredCapture")
	}
	handIdx, err := strconv.Atoi(args[0])
	if err != nil {
		return i18n.Tf("invalidCardIndex", "val", args[0])
	}
	tableIdxs, _ := cuiutil.ParseIntSlice(args[1:])
	return c.bi.Play(handIdx, tableIdxs)
}
