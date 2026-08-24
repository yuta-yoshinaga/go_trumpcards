//go:build !js || !wasm || extra4

package controller

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// PiedmonteseTarotCuiController はピエモンテ・タロッコの CUI コントローラー。
type PiedmonteseTarotCuiController struct {
	di usecase.PiedmonteseTarotInteractorIF
}

// NewPiedmonteseTarotCuiController コンストラクタ。
func NewPiedmonteseTarotCuiController(di usecase.PiedmonteseTarotInteractorIF) *PiedmonteseTarotCuiController {
	return &PiedmonteseTarotCuiController{di: di}
}

// Exec コマンド実行。
//
//	q / quit                   → ゲーム終了 ("bye.")
//	r / reset                  → ゲームリセット (設定保持)
//	scarto <i0> <i1> [i2]      → タロンぶんを捨てる (親のみ)
//	discard <i0> <i1> [i2]     → scarto のエイリアス
//	play <n>                   → カードをプレイ
//	n / next                   → 次のトリックへ
//	nr / nextround             → 次のディールへ (精算)
//	ss / setseats <3|4>        → 席数設定
//	sd / setdifficulty <0-2>   → CPU 難易度設定
//	h / hint                   → ヒント表示
//	log / l                    → 棋譜表示
func (c *PiedmonteseTarotCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.di.ResetWithConfig(c.di.GetConfig()) },
		[]string{
			"scarto", "discard", "play",
			"n", "next", "nr", "nextround",
			"ss", "setseats", "sd", "setdifficulty", "h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "scarto", "discard":
				return c.execScarto(args)
			case "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex",
					cuiutil.NoMin, cuiutil.NoMax, c.di.Play)
			case "n", "next":
				return c.di.NextTrick(), true
			case "nr", "nextround":
				return c.di.NextRound(), true
			case "ss", "setseats":
				return c.execSetSeats(args)
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.di.GetConfig()
					cfg.CpuDifficulty = domain.PiedmonteseTarotCpuDifficulty(v)
					return c.di.ResetWithConfig(cfg)
				})
			default:
				return handleCuiHintAndLog(cmd, c.di.Hint, c.di.ActionLog)
			}
		},
	)
}

// execScarto は scarto / discard を解釈する。
//
// **枚数はタロンで決まる。** 4 人卓は 2 枚、3 人卓は 3 枚 ── 席数を見ずに固定で
// 検査すると、片方の卓で必ず弾かれる。
func (c *PiedmonteseTarotCuiController) execScarto(args []string) (string, bool) {
	want := domain.PiedmonteseTarotTalonSize(c.di.GetConfig().Seats)
	if len(args) < want {
		return invalidArg("cardIndicesRequiredScartoN", "n", strconv.Itoa(want)), true
	}
	indices, skipped := cuiutil.ParseIntSlice(args)
	if len(skipped) > 0 {
		return invalidArg("invalidCardIndex", "val", skipped[0]), true
	}
	return c.di.Discard(indices), true
}

// execSetSeats は席数を変える。選べるのは 3 か 4 だけ。
func (c *PiedmonteseTarotCuiController) execSetSeats(args []string) (string, bool) {
	if len(args) == 0 {
		return invalidArg("numberOfSeatsRequired34"), true
	}
	return cuiutil.WithParsedIntKeys(args, "numberOfSeatsRequired34", "invalidNumberOfSeats34",
		cuiutil.NoMin, cuiutil.NoMax, func(v int) string {
			cfg := c.di.GetConfig()
			cfg.Seats = v
			if err := cfg.Validate(); err != nil {
				return invalidArg("invalidNumberOfSeats34")
			}
			return c.di.ResetWithConfig(cfg)
		})
}
