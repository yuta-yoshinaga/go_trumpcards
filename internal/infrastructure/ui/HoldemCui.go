package ui

import (
	"bufio"
	"fmt"
	"os"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// HoldemCui テキサスホールデムCUIクラス
type HoldemCui struct {
	hc *controller.HoldemCuiController
}

// NewHoldemCui コンストラクタ
func NewHoldemCui() *HoldemCui {
	tc := domain.NewTrumpCards(0)
	players := []*domain.HoldemPlayer{
		domain.NewHoldemPlayer(true, domain.HoldemStyleTAG),
		domain.NewHoldemPlayer(false, domain.HoldemStyleLAP),
		domain.NewHoldemPlayer(false, domain.HoldemStyleTAP),
		domain.NewHoldemPlayer(false, domain.HoldemStyleLAG),
	}
	holdem := domain.NewHoldem(tc, players, domain.DefaultHoldemConfig())
	return &HoldemCui{
		hc: controller.NewHoldemCuiController(usecase.NewHoldemInteractor(holdem, presenter.NewHoldemCuiPresenter())),
	}
}

// Exec ゲーム実行
func (cui *HoldemCui) Exec() {
	fmt.Println(cui.hc.Exec("r"))
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("--- Commands ---")
	fmt.Println("q・・・quit")
	fmt.Println("r・・・reset")
	fmt.Println("f・・・fold")
	fmt.Println("ck・・・check")
	fmt.Println("c・・・call")
	fmt.Println("b [amount]・・・bet (e.g. 'b 20')")
	fmt.Println("ra [amount]・・・raise (e.g. 'ra 30')")
	fmt.Println("a・・・allin")
	fmt.Println("bl [0-2]・・・betting limit (0=Fixed, 1=PotLimit, 2=NoLimit)")
	fmt.Println("----------------")

	for {
		fmt.Print("\nPlease enter a command > ")
		input, exit := readInput(scanner)
		if exit {
			break
		}
		res := cui.hc.Exec(input)
		fmt.Println(res)
		if res == "bye." {
			break
		}
	}
}
