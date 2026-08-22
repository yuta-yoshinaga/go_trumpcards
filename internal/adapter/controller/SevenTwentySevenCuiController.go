//go:build !js || !wasm || extra4

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// SevenTwentySevenCuiController はセブン・トゥエンティセブン (SevenTwentySeven) の CUI コントローラークラス。
type SevenTwentySevenCuiController struct {
	ti usecase.SevenTwentySevenInteractorIF
}

// NewSevenTwentySevenCuiController コンストラクタ。
func NewSevenTwentySevenCuiController(ti usecase.SevenTwentySevenInteractorIF) *SevenTwentySevenCuiController {
	return &SevenTwentySevenCuiController{ti: ti}
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
func (c *SevenTwentySevenCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.ti.GetConfig()
			return c.ti.ResetWithConfig(cfg)
		},
		[]string{
			"c", "card", "s", "stand",
			"n", "next", "nr", "nextround",
			"sp", "setplayers", "sa", "setante",
			"sc", "setchips", "st", "setrounds",
			"h", "hint", "l", "log",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "c", "card":
				return c.ti.TakeCard(true), true
			case "s", "stand":
				return c.ti.TakeCard(false), true
			case "n", "next", "nr", "nextround":
				return c.ti.NextRound(), true
			case "sp", "setplayers":
				return cuiutil.WithParsedIntKeys(args, "playerCountRequiredEGSp4", "invalidPlayerCount27", domain.SevenTwentySevenMinPlayerCount, domain.SevenTwentySevenMaxPlayerCount, func(v int) string {
					cfg := c.ti.GetConfig()
					cfg.PlayerCount = v
					return c.ti.ResetWithConfig(cfg)
				})
			case "sa", "setante":
				return cuiutil.WithParsedIntKeys(args, "anteRequired", "invalidAnte", domain.SevenTwentySevenMinAnte, domain.SevenTwentySevenMaxAnte, func(v int) string {
					cfg := c.ti.GetConfig()
					cfg.Ante = v
					return c.ti.ResetWithConfig(cfg)
				})
			case "sc", "setchips":
				return cuiutil.WithParsedIntKeys(args, "startingChipsRequired", "invalidStartingChips", domain.SevenTwentySevenMinStartingChips, domain.SevenTwentySevenMaxStartingChips, func(v int) string {
					cfg := c.ti.GetConfig()
					cfg.StartingChips = v
					return c.ti.ResetWithConfig(cfg)
				})
			case "st", "setrounds":
				return cuiutil.WithParsedIntKeys(args, "targetRoundsRequired", "invalidTargetRounds", domain.SevenTwentySevenMinTargetRounds, domain.SevenTwentySevenMaxTargetRounds, func(v int) string {
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
