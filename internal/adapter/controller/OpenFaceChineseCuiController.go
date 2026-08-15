//go:build !js || !wasm || casino

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// OpenFaceChineseCuiController オープンフェイス・チャイニーズポーカー (OFC) のCUIコントローラークラス
type OpenFaceChineseCuiController struct {
	di usecase.OpenFaceChineseInteractorIF
}

// NewOpenFaceChineseCuiController コンストラクタ
func NewOpenFaceChineseCuiController(di usecase.OpenFaceChineseInteractorIF) *OpenFaceChineseCuiController {
	return &OpenFaceChineseCuiController{di: di}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit                    → ゲーム終了 ("bye.")
//	r / reset                   → ゲームリセット (設定保持)
//	p / place <0/1/2>           → 保留カードを段に置く (0=front,1=middle,2=back)
//	front / f                   → 上段に置く (place 0)
//	middle / m                  → 中段に置く (place 1)
//	back / b                    → 下段に置く (place 2)
//	nr / nextround              → 次のラウンドへ
//	sd / setdifficulty <0-2>    → CPU難易度設定
//	sp / setplayers <2-4>       → プレイヤー数設定
//	h / hint                    → ヒント表示
//	log / l                     → 棋譜表示
func (c *OpenFaceChineseCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.di.GetConfig()
			return c.di.ResetWithConfig(cfg)
		},
		[]string{
			"p", "place", "front", "f", "middle", "m", "back", "b",
			"nr", "nextround", "sd", "setdifficulty", "sp", "setplayers",
			"h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "p", "place":
				return cuiutil.WithParsedInt(args, "Row is required (0=front, 1=middle, 2=back).",
					"Invalid row: %s. Please enter 0, 1 or 2.",
					domain.OpenFaceChineseRowFront, domain.OpenFaceChineseRowBack, c.di.Place)
			case "front", "f":
				return c.di.Place(domain.OpenFaceChineseRowFront), true
			case "middle", "m":
				return c.di.Place(domain.OpenFaceChineseRowMiddle), true
			case "back", "b":
				return c.di.Place(domain.OpenFaceChineseRowBack), true
			case "nr", "nextround":
				return c.di.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.di.GetConfig()
					cfg.CpuDifficulty = domain.OpenFaceChineseCpuDifficulty(v)
					return c.di.ResetWithConfig(cfg)
				})
			case "sp", "setplayers":
				return cuiutil.WithParsedIntKeys(args, "playerCountRequired", "invalidPlayerCount", domain.OpenFaceChinesePlayerMin, domain.OpenFaceChinesePlayerMax, func(v int) string {
					cfg := c.di.GetConfig()
					cfg.PlayerCount = v
					return c.di.ResetWithConfig(cfg)
				})
			default:
				return handleCuiHintAndLog(cmd, c.di.Hint, c.di.ActionLog)
			}
		},
	)
}
