package controller

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// OldMaidCuiController ババ抜きCUIコントローラークラス
type OldMaidCuiController struct {
	omi usecase.OldMaidInteractorIF
}

// NewOldMaidCuiController コンストラクタ
func NewOldMaidCuiController(omi usecase.OldMaidInteractorIF) *OldMaidCuiController {
	return &OldMaidCuiController{omi: omi}
}

// Exec コマンド実行
// draw コマンドは "d N" または "draw N" の形式でカードインデックスを指定可能。
// 例: "d 2" → インデックス2のカードを引く / "d" → ランダムに引く
func (c *OldMaidCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.omi.Reset(domain.DefaultOldMaidConfig()) },
		func(cmd string) string { return "コマンドが不明です: " + cmd },
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "d", "draw":
				cardIdx := -1
				if len(args) >= 1 {
					if idx, err := strconv.Atoi(args[0]); err == nil {
						cardIdx = idx
					}
				}
				return c.omi.Draw(cardIdx), true
			case "s", "shuffle":
				return c.omi.Shuffle(), true
			}
			return "", false
		},
	)
}
