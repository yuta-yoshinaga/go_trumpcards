//go:build !js || !wasm || solo

package controller

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// RussianBankCuiController ロシアンバンク (クラペット) のCUIコントローラークラス。
type RussianBankCuiController struct {
	di usecase.RussianBankInteractorIF
}

// NewRussianBankCuiController コンストラクタ。
func NewRussianBankCuiController(di usecase.RussianBankInteractorIF) *RussianBankCuiController {
	return &RussianBankCuiController{di: di}
}

// rbParseSource 移動元トークンを解析する。
//
//	r  → 自リザーブ, w  → 自廃札
//	or → 相手リザーブ, ow → 相手廃札
//	t0..t3 → 共有タブロー列
func rbParseSource(tok string) (zone int, fromOpp bool, col int, ok bool) {
	switch strings.ToLower(tok) {
	case "r":
		return int(domain.RussianBankZoneReserve), false, 0, true
	case "w":
		return int(domain.RussianBankZoneWaste), false, 0, true
	case "or":
		return int(domain.RussianBankZoneReserve), true, 0, true
	case "ow":
		return int(domain.RussianBankZoneWaste), true, 0, true
	}
	if len(tok) == 2 && (tok[0] == 't' || tok[0] == 'T') {
		if c, err := strconv.Atoi(tok[1:]); err == nil && c >= 0 && c < domain.RussianBankTableauCnt {
			return int(domain.RussianBankZoneTableau), false, c, true
		}
	}
	return 0, false, 0, false
}

// Exec コマンド実行。
// コマンド一覧:
//
//	q / quit                 → ゲーム終了
//	r / reset                → ゲームリセット (設定保持)
//	pf <src>                 → 移動元をファウンデーションへ (src: r/w/or/ow/t0..t3)
//	mt <src> <col>           → 移動元を共有タブロー col へ
//	d / discard              → 手札を1枚捨てて手番終了
//	s / stop                 → CPUの取りこぼしを咎める
//	u / undo                 → 直近の手を取り消す
//	sd / setdifficulty <0-2> → CPU難易度設定
//	h / hint                 → ヒント表示
//	log / l                  → 棋譜表示
func (c *RussianBankCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			return c.di.ResetWithConfig(c.di.GetConfig())
		},
		[]string{
			"pf", "mt", "d", "discard", "s", "stop", "u", "undo",
			"sd", "setdifficulty", "h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "pf":
				if len(args) < 1 {
					return "Source is required (r/w/or/ow/t0..t3).", true
				}
				zone, opp, col, ok := rbParseSource(args[0])
				if !ok {
					return "Invalid source: " + args[0] + ".", true
				}
				return c.di.MoveToFoundation(zone, opp, col), true
			case "mt":
				if len(args) < 2 {
					return "Usage: mt <src> <col>.", true
				}
				zone, opp, col, ok := rbParseSource(args[0])
				if !ok {
					return "Invalid source: " + args[0] + ".", true
				}
				toCol, err := strconv.Atoi(args[1])
				if err != nil || toCol < 0 || toCol >= domain.RussianBankTableauCnt {
					return "Invalid column: " + args[1] + ".", true
				}
				return c.di.MoveToTableau(zone, opp, col, toCol), true
			case "d", "discard":
				return c.di.Discard(), true
			case "s", "stop":
				return c.di.CallStop(), true
			case "u", "undo":
				return c.di.Undo(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired",
					"invalidCpuDifficulty", 0, 2, func(v int) string {
						cfg := c.di.GetConfig()
						cfg.CpuDifficulty = domain.RussianBankCpuDifficulty(v)
						return c.di.ResetWithConfig(cfg)
					})
			default:
				return handleCuiHintAndLog(cmd, c.di.Hint, c.di.ActionLog)
			}
		},
	)
}
