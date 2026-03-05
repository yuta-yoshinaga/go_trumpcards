package controller

import (
	"strconv"

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
	return execCuiCommand(
		command,
		func(_ []string) string { return bcc.bji.Reset() },
		func(_ string) string { return "Unsupported command." },
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "h", "hit":
				return bcc.bji.Hit(), true
			case "s", "stand":
				return bcc.bji.Stand(), true
			case "b", "bet":
				if len(args) < 1 {
					return "Bet amount is required.", true
				}
				amount, err := strconv.Atoi(args[0])
				if err != nil || amount <= 0 {
					return "Invalid bet amount. Please enter a number.", true
				}
				ppBet := 0
				t3Bet := 0
				if len(args) >= 2 {
					if v, e := strconv.Atoi(args[1]); e == nil {
						ppBet = v
					}
				}
				if len(args) >= 3 {
					if v, e := strconv.Atoi(args[2]); e == nil {
						t3Bet = v
					}
				}
				return bcc.bji.Bet(amount, ppBet, t3Bet), true
			case "d", "doubledown":
				return bcc.bji.DoubleDown(), true
			case "sp", "split":
				return bcc.bji.Split(), true
			case "i", "insurance":
				return bcc.bji.Insurance(), true
			case "di", "declineinsurance":
				return bcc.bji.DeclineInsurance(), true
			case "sur", "surrender":
				return bcc.bji.Surrender(), true
			case "hint", "togglehint":
				return bcc.bji.ToggleHint(), true
			case "soft17", "togglesoft17":
				return bcc.bji.ToggleSoft17(), true
			case "counting", "togglecounting":
				return bcc.bji.ToggleCounting(), true
			case "sd", "setdeckcount":
				if len(args) < 1 {
					return "Deck count is required.", true
				}
				count, err := strconv.Atoi(args[0])
				if err != nil || count <= 0 {
					return "Invalid deck count. Please enter a number.", true
				}
				return bcc.bji.SetDeckCount(count), true
			}
			return "", false
		},
	)
}
