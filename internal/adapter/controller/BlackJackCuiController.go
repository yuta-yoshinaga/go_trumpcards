package controller

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BlackJackCuiController ブラックジャックCUIコントローラークラス
type BlackJackCuiController struct {
	bji usecase.BlackJackInteractorIF
}

// NewBlackJackCuiController コンストラクタ
func NewBlackJackCuiController(bji usecase.BlackJackInteractorIF) *BlackJackCuiController {
	return &BlackJackCuiController{
		bji: bji,
	}
}

// Exec ゲーム実行
// コマンド例: "r", "h", "s", "b 100", "d", "sp", "i", "di", "q"
func (bcc *BlackJackCuiController) Exec(command string) string {
	res := ""
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return "Unsupported command."
	}
	cmd := parts[0]
	switch cmd {
	case "q", "quit":
		res = "bye."
	case "r", "reset":
		res = bcc.bji.Reset()
	case "h", "hit":
		res = bcc.bji.Hit()
	case "s", "stand":
		res = bcc.bji.Stand()
	case "b", "bet":
		if len(parts) < 2 {
			res = "Bet amount is required."
		} else {
			amount, err := strconv.Atoi(parts[1])
			if err == nil && amount > 0 {
				ppBet := 0
				t3Bet := 0
				if len(parts) >= 3 {
					if v, e := strconv.Atoi(parts[2]); e == nil {
						ppBet = v
					}
				}
				if len(parts) >= 4 {
					if v, e := strconv.Atoi(parts[3]); e == nil {
						t3Bet = v
					}
				}
				res = bcc.bji.Bet(amount, ppBet, t3Bet)
			} else {
				res = "Invalid bet amount. Please enter a number."
			}
		}
	case "d", "doubledown":
		res = bcc.bji.DoubleDown()
	case "sp", "split":
		res = bcc.bji.Split()
	case "i", "insurance":
		res = bcc.bji.Insurance()
	case "di", "declineinsurance":
		res = bcc.bji.DeclineInsurance()
	case "sur", "surrender":
		res = bcc.bji.Surrender()
	case "hint", "togglehint":
		res = bcc.bji.ToggleHint()
	case "soft17", "togglesoft17":
		res = bcc.bji.ToggleSoft17()
	case "counting", "togglecounting":
		res = bcc.bji.ToggleCounting()
	case "sd", "setdeckcount":
		if len(parts) < 2 {
			res = "Deck count is required."
		} else {
			count, err := strconv.Atoi(parts[1])
			if err == nil && count > 0 {
				res = bcc.bji.SetDeckCount(count)
			} else {
				res = "Invalid deck count. Please enter a number."
			}
		}
	default:
		res = "Unsupported command."
	}
	return res
}
