package controller

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuimsg"
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
		unknownCommandMessage,
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "h", "hit":
				return bcc.bji.Hit(), true
			case "s", "stand":
				return bcc.bji.Stand(), true
			case "b", "bet":
				if len(args) < 1 {
					return cuimsg.Required("Bet amount"), true
				}
				amount, err := strconv.Atoi(args[0])
				if err != nil || amount <= 0 {
					return cuimsg.InvalidNotANumber("bet amount"), true
				}
				ppBet := 0
				t3Bet := 0
				handCount := 0
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
				if len(args) >= 4 {
					if v, e := strconv.Atoi(args[3]); e == nil {
						handCount = v
					}
				}
				return bcc.bji.Bet(amount, ppBet, t3Bet, handCount), true
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
			case "es", "earlysurrender":
				return bcc.bji.EarlySurrender(), true
			case "des", "declineearlysurrender":
				return bcc.bji.DeclineEarlySurrender(), true
			case "ssr", "setsurrenderrule":
				if len(args) < 1 {
					return cuimsg.Required("Surrender rule"), true
				}
				rule, err := strconv.Atoi(args[0])
				if err != nil {
					return cuimsg.InvalidNotANumber("surrender rule"), true
				}
				return bcc.bji.SetSurrenderRule(rule), true
			case "hint", "togglehint":
				return bcc.bji.ToggleHint(), true
			case "soft17", "togglesoft17":
				return bcc.bji.ToggleSoft17(), true
			case "counting", "togglecounting":
				return bcc.bji.ToggleCounting(), true
			case "das", "toggledas":
				return bcc.bji.ToggleDAS(), true
			case "sd", "setdeckcount":
				if len(args) < 1 {
					return cuimsg.Required("Deck count"), true
				}
				count, err := strconv.Atoi(args[0])
				if err != nil || count <= 0 {
					return cuimsg.InvalidNotANumber("deck count"), true
				}
				return bcc.bji.SetDeckCount(count), true
			case "scc", "setcpucount":
				if len(args) < 1 {
					return cuimsg.Required("CPU player count"), true
				}
				count, err := strconv.Atoi(args[0])
				if err != nil {
					return cuimsg.InvalidNotANumber("CPU player count"), true
				}
				return bcc.bji.SetCpuPlayerCount(count), true
			case "scs", "setcountingsystem":
				if len(args) < 1 {
					return cuimsg.Required("Counting system"), true
				}
				system, err := strconv.Atoi(args[0])
				if err != nil {
					return cuimsg.InvalidNotANumber("counting system"), true
				}
				return bcc.bji.SetCountingSystem(system), true
			case "pen", "setpenetration":
				if len(args) < 1 {
					return cuimsg.Required("Penetration rate"), true
				}
				pen, err := strconv.Atoi(args[0])
				if err != nil {
					return cuimsg.InvalidNotANumber("penetration rate"), true
				}
				return bcc.bji.SetDeckPenetration(pen), true
			}
			return "", false
		},
	)
}
