//go:build !js || !wasm || extra

package controller

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// GleekCuiController グリーク (Gleek) のCUIコントローラークラス
type GleekCuiController struct {
	di usecase.GleekInteractorIF
}

// NewGleekCuiController コンストラクタ
func NewGleekCuiController(di usecase.GleekInteractorIF) *GleekCuiController {
	return &GleekCuiController{di: di}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit                       → ゲーム終了 ("bye.")
//	r / reset                      → ゲームリセット (設定保持)
//	bid pass                       → 競りから降りる
//	bid <n>                        → n まで競り上げる (刻みは 2)
//	discard <i> <j> ...            → 落札者が 7 枚捨てる (捨て札フェーズ)
//	play <n>                       → カードをプレイ (プレイフェーズ)
//	n / next                       → 次のトリックへ
//	nr / nextround                 → 次のディールへ (スコアリング)
//	sd / setdifficulty <0-2>       → CPU難易度設定
//	h / hint                       → ヒント表示
//	log / l                        → 棋譜表示
func (c *GleekCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.di.GetConfig()
			return c.di.ResetWithConfig(cfg)
		},
		[]string{
			"bid", "discard", "d", "play",
			"n", "next", "nr", "nextround",
			"sd", "setdifficulty", "h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "bid":
				return c.execBid(args)
			case "discard", "d":
				return c.execDiscard(args)
			case "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.di.Play)
			case "n", "next":
				return c.di.NextTrick(), true
			case "nr", "nextround":
				return c.di.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.di.GetConfig()
					cfg.CpuDifficulty = domain.GleekCpuDifficulty(v)
					return c.di.ResetWithConfig(cfg)
				})
			default:
				return handleCuiHintAndLog(cmd, c.di.Hint, c.di.ActionLog)
			}
		},
	)
}

// execBid bid サブコマンドを解釈する。
func (c *GleekCuiController) execBid(args []string) (string, bool) {
	if len(args) == 0 {
		return invalidArg("bidAmountRequired"), true
	}
	if args[0] == "pass" || args[0] == "p" {
		return c.di.Bid(0), true
	}
	amount, err := strconv.Atoi(args[0])
	if err != nil {
		return invalidArg("invalidBidAmount", "val", args[0]), true
	}
	return c.di.Bid(amount), true
}

// execDiscard discard サブコマンドを解釈する。
//
// **捨て札フェーズを抜ける唯一の入力。** 落札の直後はこのフェーズで、7 枚捨てる
// まで play は「フェーズが違う」で弾かれ続ける。
func (c *GleekCuiController) execDiscard(args []string) (string, bool) {
	if len(args) == 0 {
		return invalidArg("discardIndicesRequired"), true
	}
	indices := make([]int, 0, len(args))
	for _, a := range args {
		for _, part := range strings.Split(a, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			v, err := strconv.Atoi(part)
			if err != nil {
				return invalidArg("invalidCardIndex", "val", part), true
			}
			indices = append(indices, v)
		}
	}
	return c.di.Discard(indices), true
}
