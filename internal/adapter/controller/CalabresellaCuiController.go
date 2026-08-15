//go:build !js || !wasm || extra

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// CalabresellaCuiController カラブレセッラ (Calabresella) のCUIコントローラークラス
type CalabresellaCuiController struct {
	di usecase.CalabresellaInteractorIF
}

// NewCalabresellaCuiController コンストラクタ
func NewCalabresellaCuiController(di usecase.CalabresellaInteractorIF) *CalabresellaCuiController {
	return &CalabresellaCuiController{di: di}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit                 → ゲーム終了 ("bye.")
//	r / reset                → ゲームリセット (設定保持)
//	bid <pass|chiamo|solo>   → ビッド (pass=降りる、chiamo=ステーク1、solo=ステーク2)
//	d / discard <n>          → monte 交換で1枚を捨てる
//	<n> / play <n>           → カードをプレイ (プレイフェーズ)
//	n / next                 → 次のトリックへ
//	nr / nextround           → 次のラウンドへ (スコアリング)
//	sd / setdifficulty <0-2> → CPU難易度設定
//	h / hint                 → ヒント表示
//	log / l                  → 棋譜表示
func (c *CalabresellaCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.di.GetConfig()
			return c.di.ResetWithConfig(cfg)
		},
		[]string{
			"bid", "d", "discard", "play",
			"n", "next", "nr", "nextround",
			"sd", "setdifficulty", "h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "bid":
				if len(args) == 0 {
					return invalidArg("bidActionRequiredChiamo"), true
				}
				switch args[0] {
				case "pass", "p":
					return c.di.Bid(domain.CalabresellaBidNone), true
				case "chiamo", "c", "call":
					return c.di.Bid(domain.CalabresellaBidChiamo), true
				case "solo", "s":
					return c.di.Bid(domain.CalabresellaBidSolo), true
				default:
					return "Invalid bid action: " + args[0] + ". Please enter pass, chiamo, or solo.", true
				}
			case "d", "discard":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.di.Discard)
			case "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.di.Play)
			case "n", "next":
				return c.di.NextTrick(), true
			case "nr", "nextround":
				return c.di.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.di.GetConfig()
					cfg.CpuDifficulty = domain.CalabresellaCpuDifficulty(v)
					return c.di.ResetWithConfig(cfg)
				})
			default:
				return handleCuiHintAndLog(cmd, c.di.Hint, c.di.ActionLog)
			}
		},
	)
}
