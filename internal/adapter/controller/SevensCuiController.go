package controller

import (
	"strconv"
	"strings"

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
	fields := strings.Fields(command)
	if len(fields) == 0 {
		return "コマンドが不明です: " + command
	}
	switch fields[0] {
	case "q", "quit":
		return "bye."
	case "r", "reset":
		if len(fields) > 1 {
			tunnelEnabled := false
			jokerCount := 0
			cpuStrategy := false
			maxPasses := domain.SevensMaxPasses
			noJokerFinish := false
			for _, f := range fields[1:] {
				switch {
				case f == "tunnel":
					tunnelEnabled = true
				case strings.HasPrefix(f, "joker="):
					if parsed, err := strconv.Atoi(strings.TrimPrefix(f, "joker=")); err == nil {
						jokerCount = parsed
					}
				case strings.HasPrefix(f, "passes="):
					if parsed, err := strconv.Atoi(strings.TrimPrefix(f, "passes=")); err == nil {
						maxPasses = parsed
					}
				case f == "strategy":
					cpuStrategy = true
				case f == "nojokerfinish":
					noJokerFinish = true
				}
			}
			return c.sgi.ResetWithConfig(tunnelEnabled, jokerCount, cpuStrategy, maxPasses, noJokerFinish)
		}
		return c.sgi.Reset()
	case "p", "play":
		idx := -1 // デフォルトはパス
		if len(fields) > 1 {
			if parsed, err := strconv.Atoi(fields[1]); err == nil {
				idx = parsed
			}
		}
		return c.sgi.Play(idx)
	case "j", "joker":
		cardIdx := 0
		targetSuit := 0
		targetValue := 0
		if len(fields) > 1 {
			if parsed, err := strconv.Atoi(fields[1]); err == nil {
				cardIdx = parsed
			}
		}
		if len(fields) > 2 {
			if parsed, err := strconv.Atoi(fields[2]); err == nil {
				targetSuit = parsed
			}
		}
		if len(fields) > 3 {
			if parsed, err := strconv.Atoi(fields[3]); err == nil {
				targetValue = parsed
			}
		}
		return c.sgi.PlayJoker(cardIdx, targetSuit, targetValue)
	default:
		return "コマンドが不明です: " + command
	}
}
