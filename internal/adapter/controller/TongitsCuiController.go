//go:build !js || !wasm || extra5

package controller

import (
	"math"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// TongitsCuiController Tongits CUIコントローラークラス
type TongitsCuiController struct {
	ci usecase.TongitsInteractorIF
}

// NewTongitsCuiController コンストラクタ
func NewTongitsCuiController(ci usecase.TongitsInteractorIF) *TongitsCuiController {
	return &TongitsCuiController{ci: ci}
}

// Exec コマンド実行
func (c *TongitsCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.ci.GetConfig()
			return c.ci.ResetWithConfig(cfg)
		},
		[]string{
			"ds", "drawstock", "dd", "drawdiscard", "d", "discard",
			"m", "meld",
			"sp", "sapaw",
			"c", "challenge",
			"nr", "nextround",
			"sd", "setdifficulty", "sl", "setlimit", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "ds", "drawstock":
				return c.ci.DrawFromStock(), true
			case "dd", "drawdiscard":
				return c.ci.DrawFromDiscard(), true
			case "d", "discard":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.ci.Discard)
			case "m", "meld":
				// メルドは複数枚を一度に公開するので、可変長の位置引数を取る。
				idx, errMsg := tongitsParseIndices(args)
				if errMsg != "" {
					return errMsg, true
				}
				return c.ci.Meld(idx), true
			case "sp", "sapaw":
				// **他家のメルドにも付け足せる**ので、宛先はプレイヤー番号 +
				// そのプレイヤーの何番目のメルドか、の2つが要る。
				nums, errMsg := tongitsParseIndices(args)
				if errMsg != "" {
					return errMsg, true
				}
				if len(nums) != 3 {
					return i18n.T("tongits.sapawArgsRequired"), true
				}
				return c.ci.Sapaw(nums[0], nums[1], nums[2]), true
			case "c", "challenge":
				// 他家全員が応じたものとして宣言する。CUI には合意を尋ねる面が
				// 無いので、ここで合意を組み立てる (Web は個別に送れる)。
				agreed := make([]bool, domain.TongitsPlayerCnt-1)
				for i := range agreed {
					agreed[i] = true
				}
				return c.ci.Challenge(agreed), true
			case "nr", "nextround":
				return c.ci.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.ci.GetConfig()
					cfg.CpuDifficulty = domain.TongitsCpuDifficulty(v)
					return c.ci.ResetWithConfig(cfg)
				})
			case "sl", "setlimit":
				return cuiutil.WithParsedIntKeys(args, "pointLimitRequired", "invalidPointLimit", 1, math.MaxInt, func(v int) string {
					cfg := c.ci.GetConfig()
					cfg.PointLimit = v
					return c.ci.ResetWithConfig(cfg)
				})
			default:
				return handleCuiLog(cmd, c.ci.ActionLog)
			}
		},
	)
}

// tongitsParseIndices は位置引数を整数の並びに直す。meld と sapaw が同じ形の
// 引数を取るので、両方から使う。1 つでも整数でなければ、どれが悪いかを返す。
func tongitsParseIndices(args []string) ([]int, string) {
	if len(args) == 0 {
		return nil, i18n.T("tongits.cardIndexRequired")
	}
	out := make([]int, 0, len(args))
	for _, a := range args {
		v, err := strconv.Atoi(a)
		if err != nil {
			return nil, i18n.Tf("tongits.invalidCardIndex", "val", a)
		}
		out = append(out, v)
	}
	return out, ""
}
