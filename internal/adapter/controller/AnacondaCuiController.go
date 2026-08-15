//go:build !js || !wasm || extra

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// AnacondaCuiController はアナコンダ (Anaconda) の CUI コントローラークラス。
type AnacondaCuiController struct {
	ti usecase.AnacondaInteractorIF
}

// NewAnacondaCuiController コンストラクタ。
func NewAnacondaCuiController(ti usecase.AnacondaInteractorIF) *AnacondaCuiController {
	return &AnacondaCuiController{ti: ti}
}

// Exec コマンド実行。
//
//	q / quit                    → ゲーム終了 ("bye.")
//	r / reset                   → ゲームリセット (設定保持)
//	p <i...> / pass <i...>      → 選んだ札を左隣へ渡す
//	k <i...> / keep <i...>      → 残す 5 枚を選ぶ
//	c / call                    → コール (チェック含む)
//	ra / raise                  → レイズ
//	f / fold                    → フォールド
//	n / next / nr / nextround   → 次のラウンドへ
//	sp <n> / setplayers <n>     → プレイヤー数設定 (3-7)
//	sa <n> / setante <n>        → アンティ額設定
//	sc <n> / setchips <n>       → 初期チップ設定
//	st <n> / setrounds <n>      → 実施ラウンド数設定
//	h / hint                    → ヒント表示
//	l / log                     → 棋譜表示
func (c *AnacondaCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.ti.GetConfig()
			return c.ti.ResetWithConfig(cfg)
		},
		[]string{
			"p", "pass", "k", "keep",
			"c", "call", "ra", "raise", "f", "fold",
			"n", "next", "nr", "nextround",
			"sp", "setplayers", "sa", "setante",
			"sc", "setchips", "st", "setrounds",
			"h", "hint", "l", "log",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "p", "pass":
				indices, skipped := cuiutil.ParseIntSlice(args)
				return cuiutil.PrependSkippedWarning(c.ti.Pass(indices), skipped), true
			case "k", "keep":
				indices, skipped := cuiutil.ParseIntSlice(args)
				return cuiutil.PrependSkippedWarning(c.ti.Keep(indices), skipped), true
			case "c", "call":
				return c.ti.Call(), true
			case "ra", "raise":
				return c.ti.Raise(), true
			case "f", "fold":
				return c.ti.Fold(), true
			case "n", "next", "nr", "nextround":
				return c.ti.NextRound(), true
			case "sp", "setplayers":
				return cuiutil.WithParsedInt(args, "Player count is required (e.g. sp 4).", "Invalid player count: %s. Please enter 3-7.", domain.AnacondaMinPlayerCount, domain.AnacondaMaxPlayerCount, func(v int) string {
					cfg := c.ti.GetConfig()
					cfg.PlayerCount = v
					return c.ti.ResetWithConfig(cfg)
				})
			case "sa", "setante":
				return cuiutil.WithParsedIntKeys(args, "anteRequired", "invalidAnte", domain.AnacondaMinAnte, domain.AnacondaMaxAnte, func(v int) string {
					cfg := c.ti.GetConfig()
					cfg.Ante = v
					return c.ti.ResetWithConfig(cfg)
				})
			case "sc", "setchips":
				return cuiutil.WithParsedIntKeys(args, "startingChipsRequired", "invalidStartingChips", domain.AnacondaMinStartingChips, domain.AnacondaMaxStartingChips, func(v int) string {
					cfg := c.ti.GetConfig()
					cfg.StartingChips = v
					return c.ti.ResetWithConfig(cfg)
				})
			case "st", "setrounds":
				return cuiutil.WithParsedIntKeys(args, "targetRoundsRequired", "invalidTargetRounds", domain.AnacondaMinTargetRounds, domain.AnacondaMaxTargetRounds, func(v int) string {
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
