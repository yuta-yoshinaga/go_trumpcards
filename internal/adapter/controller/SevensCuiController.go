package controller

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// SevensCuiController 7並べCUIコントローラークラス
type SevensCuiController struct {
	sgi usecase.SevensInteractorIF
}

// NewSevensCuiController コンストラクタ
func NewSevensCuiController(sgi usecase.SevensInteractorIF) *SevensCuiController {
	return &SevensCuiController{sgi: sgi}
}

// Exec コマンド実行
// play コマンドは "p [インデックス]" の形式でカードインデックスを指定。
// インデックスなし ("p") の場合はパス扱い (idx = -1)。
// joker コマンドは "j [カードインデックス] [スート] [値]" の形式。
// 例: "p"  → パス / "p 2" → 2番のカードを出す / "j 0 1 6" → ジョーカー(手札0)をスート1値6に配置
func (c *SevensCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(args []string) string {
			if len(args) > 0 {
				cfg := domain.DefaultSevensConfig()
				for _, f := range args {
					switch {
					case f == "tunnel":
						cfg.TunnelEnabled = true
					case strings.HasPrefix(f, "tunnelskip="):
						if parsed, err := strconv.Atoi(strings.TrimPrefix(f, "tunnelskip=")); err == nil {
							cfg.TunnelSkipWidth = parsed
						}
					case strings.HasPrefix(f, "joker="):
						if parsed, err := strconv.Atoi(strings.TrimPrefix(f, "joker=")); err == nil {
							cfg.JokerCount = parsed
						}
					case strings.HasPrefix(f, "passes="):
						if parsed, err := strconv.Atoi(strings.TrimPrefix(f, "passes=")); err == nil {
							cfg.MaxPasses = parsed
						}
					case f == "strategy":
						cfg.CpuStrategy = domain.SevensCpuStrategic
					case f == "harassment":
						cfg.CpuStrategy = domain.SevensCpuHarassment
					case f == "nojokerfinish":
						cfg.NoJokerFinish = true
					case f == "jokerreclaim":
						cfg.JokerReclaimEnabled = true
					case f == "endstop":
						cfg.EndStopEnabled = true
					case f == "jokerconsban":
						cfg.JokerConsecutiveBanned = true
					}
				}
				return c.sgi.ResetWithConfig(cfg)
			}
			return c.sgi.Reset()
		},
		[]string{"p", "play", "j", "joker", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "p", "play":
				idx, errMsg, ok := cuiutil.ParseOptionalIntKeys(args, 0, -1, "invalidCardIndex")
				if !ok {
					return errMsg, true
				}
				return c.sgi.Play(idx), true
			case "j", "joker":
				cardIdx, errMsg, ok := cuiutil.ParseOptionalIntKeys(args, 0, 0, "invalidCardIndex")
				if !ok {
					return errMsg, true
				}
				targetSuit, errMsg, ok := cuiutil.ParseOptionalIntKeys(args, 1, 0, "invalidSuit")
				if !ok {
					return errMsg, true
				}
				targetValue, errMsg, ok := cuiutil.ParseOptionalIntKeys(args, 2, 0, "invalidTargetValue")
				if !ok {
					return errMsg, true
				}
				return c.sgi.PlayJoker(cardIdx, targetSuit, targetValue), true
			default:
				return handleCuiLog(cmd, c.sgi.ActionLog)
			}
		},
	)
}
