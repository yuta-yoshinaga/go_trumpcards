package ui

import (
	"bufio"
	"fmt"
	"os"
	"go_trumpcards/interface_adapters/controllers"
	"go_trumpcards/interface_adapters/presenters"
	"go_trumpcards/usecases"
)

// BlackJackCui ブラックジャックCUIクラス
type BlackJackCui struct {
	bjc *controllers.BlackJackCuiController
}

// NewBlackJackCui コンストラクタ
func NewBlackJackCui() *BlackJackCui {
	return &BlackJackCui{
		bjc: controllers.NewBlackJackCuiController(usecases.NewBlackJackInteractor(presenters.NewBlackJackCuiPresenter())),
	}
}

// Exec ゲーム実行
func (cui *BlackJackCui) Exec() {
	fmt.Println(cui.bjc.Exec("r"))
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Println("Please enter a command.")
		fmt.Println("q・・・quit")
		fmt.Println("r・・・reset")
		fmt.Println("h・・・hit")
		fmt.Println("s・・・stand")
		scanner.Scan()
		res := cui.bjc.Exec(scanner.Text())
		fmt.Println(res)
		if res == "bye." {
			break
		}
	}
}
