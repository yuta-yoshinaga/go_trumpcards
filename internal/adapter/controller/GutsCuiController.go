//go:build !js || !wasm || extra

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// GutsCuiController はガッツ (Guts) の CUI コントローラークラス。
type GutsCuiController struct {
	ti usecase.GutsInteractorIF
}

// NewGutsCuiController コンストラクタ。
func NewGutsCuiController(ti usecase.GutsInteractorIF) *GutsCuiController {
	return &GutsCuiController{ti: ti}
}

// Exec コマンド実行。
//
//	q / quit                 → ゲーム終了 ("bye.")
//	r / reset                → ゲームリセット (設定保持)
//	i / in                   → イン宣言 (勝負に残る)
//	o / out                  → アウト宣言 (降りる)
//	n / next / nr / nextround → 次のラウンドへ
//	sp <n> / setplayers <n>  → プレイヤー数設定 (2-7)
//	sa <n> / setante <n>     → アンティ額設定
//	sc <n> / setchips <n>    → 初期チップ設定
//	st <n> / setrounds <n>   → 実施ラウンド数設定
//	h / hint                 → ヒント表示
//	l / log                  → 棋譜表示
func (c *GutsCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.ti.GetConfig()
			return c.ti.ResetWithConfig(cfg)
		},
		[]string{
			"i", "in", "o", "out",
			"n", "next", "nr", "nextround",
			"sp", "setplayers", "sa", "setante",
			"sc", "setchips", "st", "setrounds",
			"h", "hint", "l", "log",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "i", "in":
				return c.ti.Declare(true), true
			case "o", "out":
				return c.ti.Declare(false), true
			case "n", "next", "nr", "nextround":
				return c.ti.NextRound(), true
			case "sp", "setplayers":
				return cuiutil.WithParsedInt(args, "Player count is required (e.g. sp 4).", "Invalid player count: %s. Please enter 2-7.", domain.GutsMinPlayerCount, domain.GutsMaxPlayerCount, func(v int) string {
					cfg := c.ti.GetConfig()
					cfg.PlayerCount = v
					return c.ti.ResetWithConfig(cfg)
				})
			case "sa", "setante":
				return cuiutil.WithParsedIntKeys(args, "anteRequired", "invalidAnte", domain.GutsMinAnte, domain.GutsMaxAnte, func(v int) string {
					cfg := c.ti.GetConfig()
					cfg.Ante = v
					return c.ti.ResetWithConfig(cfg)
				})
			case "sc", "setchips":
				return cuiutil.WithParsedIntKeys(args, "startingChipsRequired", "invalidStartingChips", domain.GutsMinStartingChips, domain.GutsMaxStartingChips, func(v int) string {
					cfg := c.ti.GetConfig()
					cfg.StartingChips = v
					return c.ti.ResetWithConfig(cfg)
				})
			case "st", "setrounds":
				return cuiutil.WithParsedIntKeys(args, "targetRoundsRequired", "invalidTargetRounds", domain.GutsMinTargetRounds, domain.GutsMaxTargetRounds, func(v int) string {
					cfg := c.ti.GetConfig()
					cfg.TargetRounds = v
					return c.ti.ResetWithConfig(cfg)
				})
			default:
				return handleCuiHintAndLog(cmd, c.ti.Hint, c.ti.ActionLog)
			}
		},
	)
}
