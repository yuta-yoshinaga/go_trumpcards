package controllers

import "go_trumpcards/usecases"

// BlackJackCuiController ブラックジャックCUIコントローラークラス
type BlackJackCuiController struct {
	bji usecases.BlackJackInteractorIF
}

// NewBlackJackCuiController コンストラクタ
func NewBlackJackCuiController(bji usecases.BlackJackInteractorIF) *BlackJackCuiController {
	return &BlackJackCuiController{
		bji: bji,
	}
}

// Exec ゲーム実行
func (bcc *BlackJackCuiController) Exec(command string) string {
	res := ""
	switch command {
	case "q", "quit":
		res = "bye."
	case "r", "reset":
		res = bcc.bji.Reset()
	case "h", "hit":
		res = bcc.bji.Hit()
	case "s", "stand":
		res = bcc.bji.Stand()
	default:
		res = "Unsupported command."
	}
	return res
}
