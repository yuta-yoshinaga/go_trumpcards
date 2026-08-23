//go:build !js || !wasm || extra

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BauernschnapsenCuiController バウエルンシュナプセンCUIコントローラークラス
type BauernschnapsenCuiController struct {
	gi usecase.BauernschnapsenInteractorIF
}

// NewBauernschnapsenCuiController コンストラクタ
func NewBauernschnapsenCuiController(gi usecase.BauernschnapsenInteractorIF) *BauernschnapsenCuiController {
	return &BauernschnapsenCuiController{gi: gi}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit           → ゲーム終了 ("bye.")
//	r / reset          → ゲームリセット (設定保持)
//	c / contract <0-3> [suit 0-3] → 契約を宣言 (0=パス 1=通常 2=同スート縛り 3=ベテル)
//	p / play <i>       → カードをプレイ
//	m / marriage <i>   → マリアージュを宣言 (リード番のみ)
//	n / next           → 次のトリックへ
//	nr / nextround     → 次のラウンドへ
//	sd / setdifficulty <0-2> → CPU難易度設定
//	st / settarget <n> → ターゲットスコア設定 (デフォルト24)
//	h / hint           → ヒント表示
//	log / l            → 棋譜表示
func (c *BauernschnapsenCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.gi.GetConfig()
			return c.gi.ResetWithConfig(cfg)
		},
		[]string{
			"c", "contract",
			"p", "play",
			"m", "marriage",
			"n", "next", "nr", "nextround",
			"sd", "setdifficulty", "st", "settarget",
			"h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "c", "contract":
				// **契約フェーズを抜ける唯一の入力。** 第 2 引数は切り札スート
				// (省略時は 0)。ベテルは切り札を取らないので無視される。
				return cuiutil.WithParsedIntKeys(args, "contractRequired", "invalidContract",
					int(domain.BauernschnapsenContractNone), int(domain.BauernschnapsenContractBettel),
					func(v int) string {
						suit := cuiutil.ParseOptionalInt(args, 1, domain.CardDesignSpade)
						return c.gi.DeclareContract(domain.BauernschnapsenContract(v), suit)
					})
			case "p", "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.gi.Play)
			case "m", "marriage":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.gi.DeclareMarriage)
			case "n", "next":
				return c.gi.NextTrick(), true
			case "nr", "nextround":
				return c.gi.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.gi.GetConfig()
					cfg.CpuDifficulty = domain.BauernschnapsenCpuDifficulty(v)
					return c.gi.ResetWithConfig(cfg)
				})
			case "st", "settarget":
				return cuiutil.WithParsedIntKeys(args, "targetScoreRequired", "invalidTargetScore", 1, math.MaxInt, func(v int) string {
					cfg := c.gi.GetConfig()
					cfg.TargetScore = v
					return c.gi.ResetWithConfig(cfg)
				})
			default:
				return handleCuiHintAndLog(cmd, c.gi.Hint, c.gi.ActionLog)
			}
		},
	)
}
