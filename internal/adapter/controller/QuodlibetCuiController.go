//go:build !js || !wasm || solo

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// QuodlibetCuiController はクオドリベットの CUI コントローラー。
type QuodlibetCuiController struct {
	di usecase.QuodlibetInteractorIF
}

// NewQuodlibetCuiController コンストラクタ。
func NewQuodlibetCuiController(di usecase.QuodlibetInteractorIF) *QuodlibetCuiController {
	return &QuodlibetCuiController{di: di}
}

// Exec コマンド実行。
//
//	q / quit                 → ゲーム終了 ("bye.")
//	r / reset                → ゲームリセット (設定保持)
//	c / contract <0-11>      → コントラクトを選ぶ (ディーラーのみ)
//	p / play <idx>           → カードを出す
//	pass                     → パス (四分 / 小食いで出せる札が無いときだけ)
//	nd / nextdeal            → 次のディールへ
//	sd / setdifficulty <0-2> → CPU 難易度設定
//	auto                     → コントラクトの自動選択を切り替えてリセット
//	h / hint                 → ヒント表示
//	log / l                  → 棋譜表示
func (c *QuodlibetCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.di.ResetWithConfig(c.di.GetConfig()) },
		[]string{
			"c", "contract", "p", "play", "pass", "nd", "nextdeal",
			"sd", "setdifficulty", "auto", "h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "c", "contract":
				return cuiutil.WithParsedIntKeys(args, "contractRequired", "invalidContract",
					0, domain.QuodlibetContractCnt-1, func(v int) string { return c.di.SelectContract(v) })
			case "p", "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex",
					cuiutil.NoMin, cuiutil.NoMax, func(v int) string { return c.di.Play(v) })
			case "pass":
				// **パスは -1 のプレイ。** 出せる札があるかの判定はドメインに 1 つだけ置く。
				return c.di.Play(-1), true
			case "nd", "nextdeal":
				return c.di.NextDeal(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2,
					func(v int) string {
						cfg := c.di.GetConfig()
						cfg.CpuDifficulty = domain.QuodlibetCpuDifficulty(v)
						return c.di.ResetWithConfig(cfg)
					})
			case "auto":
				cfg := c.di.GetConfig()
				cfg.AutoSelectContract = !cfg.AutoSelectContract
				return c.di.ResetWithConfig(cfg), true
			default:
				return handleCuiHintAndLog(cmd, c.di.Hint, c.di.ActionLog)
			}
		},
	)
}
