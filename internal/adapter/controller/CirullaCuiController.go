//go:build !js || !wasm || extra3

package controller

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// CirullaCuiController はチルッラの CUI コントローラー。
type CirullaCuiController struct {
	di usecase.CirullaInteractorIF
}

// NewCirullaCuiController コンストラクタ。
func NewCirullaCuiController(di usecase.CirullaInteractorIF) *CirullaCuiController {
	return &CirullaCuiController{di: di}
}

// Exec コマンド実行。
//
//	q / quit                 → ゲーム終了 ("bye.")
//	r / reset                → ゲームリセット (設定保持)
//	p / play <idx> [t0 t1..] → 手札を出す (場札の番号を並べると取る)
//	nr / nextround           → 次のラウンドへ
//	sd / setdifficulty <0-2> → CPU 難易度設定
//	st / settarget <11-51>   → 目標点
//	h / hint                 → ヒント表示
//	log / l                  → 棋譜表示
func (c *CirullaCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.di.ResetWithConfig(c.di.GetConfig()) },
		[]string{
			"p", "play", "nr", "nextround",
			"sd", "setdifficulty", "st", "settarget", "h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "p", "play":
				return c.execPlay(args)
			case "nr", "nextround":
				return c.di.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2,
					func(v int) string {
						cfg := c.di.GetConfig()
						cfg.CpuDifficulty = domain.CirullaCpuDifficulty(v)
						return c.di.ResetWithConfig(cfg)
					})
			case "st", "settarget":
				return cuiutil.WithParsedIntKeys(args, "targetScoreRequired", "cirulla.invalidTargetScore",
					domain.CirullaMinTarget, domain.CirullaMaxTarget,
					func(v int) string {
						cfg := c.di.GetConfig()
						cfg.TargetScore = v
						return c.di.ResetWithConfig(cfg)
					})
			default:
				return handleCuiHintAndLog(cmd, c.di.Hint, c.di.ActionLog)
			}
		},
	)
}

// execPlay は "play <idx> [場札の番号...]" を解釈する。
//
// **取る札は同じ 1 行に書く。** 出してから別コマンドで取らせると、
// 「出したが取っていない」盤面が生まれる。
func (c *CirullaCuiController) execPlay(args []string) (string, bool) {
	return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex",
		cuiutil.NoMin, cuiutil.NoMax, func(handIdx int) string {
			captures, bad := cirullaParseCaptureArgs(args[1:])
			// **読めない番号を「場に置く」に落とさない。** 落とすと `p 0 zz` が
			// 黙って場に足す手になり、取ったつもりの札が相手に残る。
			if bad != "" {
				return i18n.MarkError(i18n.Tf("invalidFieldIndex", "val", bad))
			}
			return c.di.Play(handIdx, captures)
		})
}

// cirullaParseCaptureArgs は 2 つ目以降の引数を場札インデックスとして読む。
// 読めない引数があれば、その文字列を 2 つ目の戻り値で返す。
func cirullaParseCaptureArgs(args []string) ([]int, string) {
	out := make([]int, 0, len(args))
	for _, a := range args {
		t := strings.TrimSpace(a)
		n, err := strconv.Atoi(t)
		if err != nil || n < 0 {
			return nil, t
		}
		out = append(out, n)
	}
	if len(out) == 0 {
		return nil, ""
	}
	return out, ""
}
