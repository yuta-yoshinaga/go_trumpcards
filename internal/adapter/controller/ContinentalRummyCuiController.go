//go:build !js || !wasm || extra2

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// ContinentalRummyCuiController はコンチネンタル・ラミーの CUI コントローラー。
type ContinentalRummyCuiController struct {
	di usecase.ContinentalRummyInteractorIF
}

// NewContinentalRummyCuiController コンストラクタ。
func NewContinentalRummyCuiController(di usecase.ContinentalRummyInteractorIF) *ContinentalRummyCuiController {
	return &ContinentalRummyCuiController{di: di}
}

// Exec コマンド実行。
//
//	q / quit                 → ゲーム終了 ("bye.")
//	r / reset                → ゲームリセット (設定保持)
//	stock / ds               → 山札から 1 枚引く
//	take / dd                → 捨て札の一番上を取る
//	discard / d <idx>        → 手札の idx 番を捨てる
//	goout / g <idx>          → 15 枚を並べて上がる (idx は捨てる 1 枚)
//	gooutdeal / gd           → 引かずに、配られた 15 枚のまま上がる
//	next / n                 → 次のラウンドへ
//	sd / setdifficulty <0-2> → CPU 難易度設定
//	sr / setrounds <1-10>    → ラウンド数
//	h / hint                 → ヒント表示
//	log / l                  → 棋譜表示
func (c *ContinentalRummyCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.di.ResetWithConfig(c.di.GetConfig()) },
		[]string{
			"stock", "ds", "take", "dd", "discard", "d", "goout", "g",
			"gooutdeal", "gd", "next", "n", "sd", "setdifficulty", "sr", "setrounds",
			"h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			// **山と捨て札は別の命令。** 引数ひとつに畳むと、付け忘れた要求が
			// 黙ってどちらかに倒れる。
			case "stock", "ds":
				return c.di.DrawStock(), true
			case "take", "dd":
				return c.di.DrawDiscard(), true
			case "discard", "d":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex",
					0, domain.ContinentalRummyHandSize, c.di.Discard)
			case "goout", "g":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex",
					0, domain.ContinentalRummyHandSize, c.di.GoOut)
			// **引かずに上がるほうは札を捨てない。** 加点が違うので別命令。
			case "gooutdeal", "gd":
				return c.di.GoOut(-1), true
			case "next", "n":
				return c.di.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2,
					func(v int) string {
						cfg := c.di.GetConfig()
						cfg.CpuDifficulty = domain.ContinentalRummyCpuDifficulty(v)
						return c.di.ResetWithConfig(cfg)
					})
			case "sr", "setrounds":
				return cuiutil.WithParsedIntKeys(args, "totalRoundsRequired", "continentalrummy.invalidRounds",
					domain.ContinentalRummyMinRounds, domain.ContinentalRummyMaxRounds,
					func(v int) string {
						cfg := c.di.GetConfig()
						cfg.TotalRounds = v
						return c.di.ResetWithConfig(cfg)
					})
			default:
				return handleCuiHintAndLog(cmd, c.di.Hint, c.di.ActionLog)
			}
		},
	)
}
