//go:build !js || !wasm || classic

package controller

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// DilotiCuiController はディロティの CUI コントローラー。
type DilotiCuiController struct {
	di usecase.DilotiInteractorIF
}

// NewDilotiCuiController コンストラクタ。
func NewDilotiCuiController(di usecase.DilotiInteractorIF) *DilotiCuiController {
	return &DilotiCuiController{di: di}
}

// Exec コマンド実行。
//
//	q / quit                    → ゲーム終了 ("bye.")
//	r / reset                   → ゲームリセット (設定保持)
//	t / take <idx> [t0 t1..] [dN..] → 取る (dN は宣言の番号)
//	d / declare <idx> <値> [t0..] [dN] → 宣言する
//	l2 / lay <idx>              → 場に置く
//	nr / nextround              → 次の局へ
//	sd / setdifficulty <0-2>    → CPU 難易度設定
//	st / settarget <21-101>     → 目標点
//	h / hint                    → ヒント表示
//	log / l                     → 棋譜表示
func (c *DilotiCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.di.ResetWithConfig(c.di.GetConfig()) },
		[]string{
			"t", "take", "d", "declare", "l2", "lay", "nr", "nextround",
			"sd", "setdifficulty", "st", "settarget", "h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "t", "take":
				return c.execTake(args)
			case "d", "declare":
				return c.execDeclare(args)
			case "l2", "lay":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex",
					cuiutil.NoMin, cuiutil.NoMax, func(handIdx int) string {
						return c.di.Play(handIdx, domain.DilotiActionTrail, nil, nil, 0)
					})
			case "nr", "nextround":
				return c.di.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2,
					func(v int) string {
						cfg := c.di.GetConfig()
						cfg.CpuDifficulty = domain.DilotiCpuDifficulty(v)
						return c.di.ResetWithConfig(cfg)
					})
			case "st", "settarget":
				return cuiutil.WithParsedIntKeys(args, "targetScoreRequired", "diloti.invalidTargetScore",
					domain.DilotiMinTarget, domain.DilotiMaxTarget,
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

// execTake は "take <idx> [場札の番号...] [d宣言の番号...]" を解釈する。
//
// **取る対象は同じ 1 行に書く。** 出してから別コマンドで取らせると、
// 「出したが取っていない」盤面が生まれる。宣言は `d0` のように d を前置する。
func (c *DilotiCuiController) execTake(args []string) (string, bool) {
	return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex",
		cuiutil.NoMin, cuiutil.NoMax, func(handIdx int) string {
			tableIdxs, declIdxs, bad := dilotiParseTargets(args[1:])
			// **読めない番号を「場に置く」に落とさない。** 落とすと `t 0 zz` が
			// 黙って場に足す手になり、取ったつもりの札が相手に残る。
			if bad != "" {
				return i18n.MarkError(i18n.Tf("invalidFieldIndex", "val", bad))
			}
			return c.di.Play(handIdx, domain.DilotiActionCapture, tableIdxs, declIdxs, 0)
		})
}

// execDeclare は "declare <idx> <宣言値> [場札の番号...] [d宣言の番号]" を解釈する。
func (c *DilotiCuiController) execDeclare(args []string) (string, bool) {
	return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex",
		cuiutil.NoMin, cuiutil.NoMax, func(handIdx int) string {
			if len(args) < 2 {
				return i18n.MarkError(i18n.T("diloti.declValueRequired"))
			}
			value, err := strconv.Atoi(strings.TrimSpace(args[1]))
			if err != nil || value < domain.DilotiMinDeclaration || value > domain.DilotiMaxDeclaration {
				return i18n.MarkError(i18n.Tf("diloti.invalidDeclValue", "val", args[1]))
			}
			tableIdxs, declIdxs, bad := dilotiParseTargets(args[2:])
			if bad != "" {
				return i18n.MarkError(i18n.Tf("invalidFieldIndex", "val", bad))
			}
			return c.di.Play(handIdx, domain.DilotiActionDeclare, tableIdxs, declIdxs, value)
		})
}

// dilotiParseTargets は引数を場札と宣言のインデックスに振り分ける。
// `d` を前置したものが宣言。読めない引数があれば、その文字列を返す。
func dilotiParseTargets(args []string) (tableIdxs, declIdxs []int, bad string) {
	tableIdxs = make([]int, 0, len(args))
	declIdxs = make([]int, 0, len(args))
	for _, a := range args {
		t := strings.TrimSpace(a)
		if t == "" {
			continue
		}
		target := &tableIdxs
		if t[0] == 'd' || t[0] == 'D' {
			target = &declIdxs
			t = t[1:]
		}
		n, err := strconv.Atoi(t)
		if err != nil || n < 0 {
			return nil, nil, strings.TrimSpace(a)
		}
		*target = append(*target, n)
	}
	if len(tableIdxs) == 0 {
		tableIdxs = nil
	}
	if len(declIdxs) == 0 {
		declIdxs = nil
	}
	return tableIdxs, declIdxs, ""
}
