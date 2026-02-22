package controllers

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/usecases"
)

// SevensCuiController 7並べCUIコントローラークラス
type SevensCuiController struct {
	sgi usecases.SevensInteractorIF
}

// NewSevensCuiController コンストラクタ
func NewSevensCuiController(sgi usecases.SevensInteractorIF) *SevensCuiController {
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
