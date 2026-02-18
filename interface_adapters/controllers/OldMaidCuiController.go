package controllers

import "github.com/yuta-yoshinaga/go_trumpcards/usecases"

// OldMaidCuiController ババ抜きCUIコントローラークラス
type OldMaidCuiController struct {
	omi usecases.OldMaidInteractorIF
}

// NewOldMaidCuiController コンストラクタ
func NewOldMaidCuiController(omi usecases.OldMaidInteractorIF) *OldMaidCuiController {
	return &OldMaidCuiController{omi: omi}
}

// Exec コマンド実行
func (c *OldMaidCuiController) Exec(command string) string {
	switch command {
	case "q", "quit":
		return "bye."
	case "r", "reset":
		return c.omi.Reset()
	case "d", "draw":
		return c.omi.Draw()
	default:
		return "コマンドが不明です: " + command
	}
}
