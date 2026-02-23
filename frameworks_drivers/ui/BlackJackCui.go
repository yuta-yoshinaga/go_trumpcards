package ui

import (
	"bufio"
	"fmt"
	"os"

	"github.com/yuta-yoshinaga/go_trumpcards/entities"
	"github.com/yuta-yoshinaga/go_trumpcards/interface_adapters/controllers"
	"github.com/yuta-yoshinaga/go_trumpcards/interface_adapters/presenters"
	"github.com/yuta-yoshinaga/go_trumpcards/usecases"
)

// BlackJackCui ブラックジャックCUIクラス
type BlackJackCui struct {
	bjc *controllers.BlackJackCuiController
}

// NewBlackJackCui コンストラクタ
func NewBlackJackCui() *BlackJackCui {
	return &BlackJackCui{
		bjc: controllers.NewBlackJackCuiController(usecases.NewBlackJackInteractor(
			entities.NewDefaultBlackJack(),
			presenters.NewBlackJackCuiPresenter(),
		)),
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
		fmt.Println("b N・・・bet (e.g. b 100)")
		fmt.Println("h・・・hit")
		fmt.Println("s・・・stand")
		fmt.Println("d・・・doubledown")
		fmt.Println("sp・・・split")
		fmt.Println("i・・・insurance")
		fmt.Println("di・・・decline insurance")
		scanner.Scan()
		res := cui.bjc.Exec(scanner.Text())
		fmt.Println(res)
		if res == "bye." {
			break
		}
	}
}
