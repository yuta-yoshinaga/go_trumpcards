//go:build !js || !wasm || casino

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// TeenPattiCuiController ティーン・パティのCUIコントローラークラス
type TeenPattiCuiController struct {
	ti usecase.TeenPattiInteractorIF
}

// NewTeenPattiCuiController コンストラクタ
func NewTeenPattiCuiController(ti usecase.TeenPattiInteractorIF) *TeenPattiCuiController {
	return &TeenPattiCuiController{ti: ti}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit                  → ゲーム終了 ("bye.")
//	r / reset                 → ゲームリセット (設定保持)
//	s / see                   → 手札を見る (Seen に昇格)
//	b / bet                   → コール (現在の賭け単位に合わせる)
//	rs <n> / raise <n>        → <n> へレイズ
//	f / fold                  → 降りる
//	sh / show                 → 勝負を要求する (残り 2 人 & Seen)
//	ss / sideshow             → サイドショーを申請する (直前の Seen と)
//	ac / accept               → サイドショー申請を受諾する
//	dc / decline              → サイドショー申請を辞退する
//	n / next / nextround      → 次のディールへ (RoundEnd フェーズ)
//	sd [0-2] / setdifficulty  → CPU難易度設定
//	sa <n> / setante <n>      → アンティ額設定
//	sc <n> / setchips <n>     → 初期チップ設定
//	h / hint                  → ヒント表示
//	l / log                   → 棋譜表示
func (c *TeenPattiCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.ti.GetConfig()
			return c.ti.ResetWithConfig(cfg)
		},
		[]string{
			"s", "see",
			"b", "bet",
			"rs", "raise",
			"f", "fold",
			"sh", "show",
			"ss", "sideshow",
			"ac", "accept",
			"dc", "decline",
			"n", "next", "nextround",
			"sd", "setdifficulty",
			"sa", "setante",
			"sc", "setchips",
			"h", "hint", "l", "log",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "s", "see":
				return c.ti.See(), true
			case "b", "bet":
				return c.ti.Bet(), true
			case "rs", "raise":
				return cuiutil.WithParsedIntKeys(args, "stakeRequiredEGRs4", "invalidStake", 1, 100000, func(v int) string {
					return c.ti.Raise(v)
				})
			case "f", "fold":
				return c.ti.Fold(), true
			case "sh", "show":
				return c.ti.Show(), true
			case "ss", "sideshow":
				return c.ti.RequestSideShow(), true
			case "ac", "accept":
				return c.ti.RespondSideShow(true), true
			case "dc", "decline":
				return c.ti.RespondSideShow(false), true
			case "n", "next", "nextround":
				return c.ti.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.ti.GetConfig()
					cfg.CpuDifficulty = domain.TeenPattiCpuDifficulty(v)
					return c.ti.ResetWithConfig(cfg)
				})
			case "sa", "setante":
				return cuiutil.WithParsedIntKeys(args, "anteRequiredEGSa2", "invalidAntePlain", 1, 1000, func(v int) string {
					cfg := c.ti.GetConfig()
					cfg.Ante = v
					return c.ti.ResetWithConfig(cfg)
				})
			case "sc", "setchips":
				return cuiutil.WithParsedIntKeys(args, "startingChipsRequiredEGSc50", "invalidStartingChipsPlain", 2, domain.TeenPattiMaxStartingChips, func(v int) string {
					cfg := c.ti.GetConfig()
					cfg.StartingChips = v
					return c.ti.ResetWithConfig(cfg)
				})
			default:
				return handleCuiHintAndLog(cmd, c.ti.Hint, c.ti.ActionLog)
			}
		},
	)
}
