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

// BlackJackCui ブラックジャックCUIクラス
type BlackJackCui struct {
	bjc *controller.BlackJackCuiController
}

// NewBlackJackCui コンストラクタ
func NewBlackJackCui() *BlackJackCui {
	return &BlackJackCui{
		bjc: controller.NewBlackJackCuiController(usecase.NewBlackJackInteractor(
			domain.NewDefaultBlackJack(),
			presenter.NewBlackJackCuiPresenter(),
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
		input, exit := readInput(scanner)
		if exit {
			break
		}
		res := cui.bjc.Exec(input)
		fmt.Println(res)
		if res == "bye." {
			break
		}
	}
}
