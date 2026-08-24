//go:build !js || !wasm || classic

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// UnsunKarutaCuiController はうんすんカルタの CUI コントローラー。
type UnsunKarutaCuiController struct {
	di usecase.UnsunKarutaInteractorIF
}

// NewUnsunKarutaCuiController コンストラクタ。
func NewUnsunKarutaCuiController(di usecase.UnsunKarutaInteractorIF) *UnsunKarutaCuiController {
	return &UnsunKarutaCuiController{di: di}
}

// Exec コマンド実行。
//
//	q / quit                 → ゲーム終了 ("bye.")
//	r / reset                → ゲームリセット (設定保持)
//	p / play <idx>           → カードを出す
//	m / meri / monchi <idx>  → 宣言して出す (リードのみ。メリ / モンチ)
//	n / next                 → 次のトリックへ
//	nr / nextround           → 次のディールへ
//	sd / setdifficulty <0-2> → CPU 難易度設定
//	st / setdeals <1-8>      → マッチのディール数
//	h / hint                 → ヒント表示
//	log / l                  → 棋譜表示
func (c *UnsunKarutaCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.di.ResetWithConfig(c.di.GetConfig()) },
		[]string{
			"p", "play", "m", "meri", "monchi",
			"n", "next", "nr", "nextround",
			"sd", "setdifficulty", "st", "setdeals", "h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "p", "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex",
					cuiutil.NoMin, cuiutil.NoMax, func(v int) string { return c.di.Play(v, false) })
			case "m", "meri", "monchi":
				// **宣言はリードの札と一緒に出す。** 宣言だけを先に送る面を作ると、
				// 「宣言したが札を出していない」状態が盤面に生まれる。
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex",
					cuiutil.NoMin, cuiutil.NoMax, func(v int) string { return c.di.Play(v, true) })
			case "n", "next":
				return c.di.NextTrick(), true
			case "nr", "nextround":
				return c.di.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2,
					func(v int) string {
						cfg := c.di.GetConfig()
						cfg.CpuDifficulty = domain.UnsunKarutaCpuDifficulty(v)
						return c.di.ResetWithConfig(cfg)
					})
			case "st", "setdeals":
				return cuiutil.WithParsedIntKeys(args, "numberOfDealsRequired18", "invalidNumberOfDeals18",
					domain.UnsunKarutaMinDeals, domain.UnsunKarutaMaxDeals, func(v int) string {
						cfg := c.di.GetConfig()
						cfg.TargetDeals = v
						return c.di.ResetWithConfig(cfg)
					})
			default:
				return handleCuiHintAndLog(cmd, c.di.Hint, c.di.ActionLog)
			}
		},
	)
}
