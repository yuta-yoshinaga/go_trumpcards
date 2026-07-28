//go:build !js || !wasm || extra3

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BouillotteCuiController はブイヨット (Bouillotte) の CUI コントローラークラス。
type BouillotteCuiController struct {
	ti usecase.BouillotteInteractorIF
}

// NewBouillotteCuiController コンストラクタ。
func NewBouillotteCuiController(ti usecase.BouillotteInteractorIF) *BouillotteCuiController {
	return &BouillotteCuiController{ti: ti}
}

// Exec コマンド実行。
//
//	q / quit                  → ゲーム終了 ("bye.")
//	r / reset                 → ゲームリセット (設定保持)
//	c / call                  → コール (現在の賭けにマッチ)
//	ra / raise / v            → レイズ (ヴィ: 賭けをアンティ分だけ上げる)
//	f / fold                  → フォールド (降りる)
//	n / next / nr / nextround → 次のラウンドへ
//	sp <n> / setplayers <n>   → プレイヤー数設定 (3-4)
//	sa <n> / setante <n>      → アンティ額設定
//	sc <n> / setchips <n>     → 初期チップ設定
//	st <n> / setrounds <n>    → 実施ラウンド数設定
//	h / hint                  → ヒント表示
//	l / log                   → 棋譜表示
func (c *BouillotteCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.ti.GetConfig()
			return c.ti.ResetWithConfig(cfg)
		},
		[]string{
			"c", "call", "ra", "raise", "v", "f", "fold",
			"n", "next", "nr", "nextround",
			"sp", "setplayers", "sa", "setante",
			"sc", "setchips", "st", "setrounds",
			"h", "hint", "l", "log",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "c", "call":
				return c.ti.Bet("call"), true
			case "ra", "raise", "v":
				return c.ti.Bet("raise"), true
			case "f", "fold":
				return c.ti.Bet("fold"), true
			case "n", "next", "nr", "nextround":
				return c.ti.NextRound(), true
			case "sp", "setplayers":
				return cuiutil.WithParsedInt(args, "Player count is required (e.g. sp 4).", "Invalid player count: %s. Please enter 3-4.", domain.BouillotteMinPlayerCount, domain.BouillotteMaxPlayerCount, func(v int) string {
					cfg := c.ti.GetConfig()
					cfg.PlayerCount = v
					return c.ti.ResetWithConfig(cfg)
				})
			case "sa", "setante":
				return cuiutil.WithParsedInt(args, "Ante is required (e.g. sa 10).", "Invalid ante: %s.", domain.BouillotteMinAnte, domain.BouillotteMaxAnte, func(v int) string {
					cfg := c.ti.GetConfig()
					cfg.Ante = v
					return c.ti.ResetWithConfig(cfg)
				})
			case "sc", "setchips":
				return cuiutil.WithParsedInt(args, "Starting chips is required (e.g. sc 200).", "Invalid starting chips: %s.", domain.BouillotteMinStartingChips, domain.BouillotteMaxStartingChips, func(v int) string {
					cfg := c.ti.GetConfig()
					cfg.StartingChips = v
					return c.ti.ResetWithConfig(cfg)
				})
			case "st", "setrounds":
				return cuiutil.WithParsedInt(args, "Target rounds is required (e.g. st 10).", "Invalid target rounds: %s.", domain.BouillotteMinTargetRounds, domain.BouillotteMaxTargetRounds, func(v int) string {
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
