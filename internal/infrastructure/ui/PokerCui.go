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

// PokerCui ポーカーCUIクラス
type PokerCui struct {
	pc *controller.PokerCuiController
}

// NewPokerCui コンストラクタ
func NewPokerCui() *PokerCui {
	config := domain.DefaultPokerConfig()
	players := []*domain.PokerPlayer{
		domain.NewPokerPlayer(true, domain.PokerStyleBalanced),
		domain.NewPokerPlayer(false, domain.PokerStyleConservative),
		domain.NewPokerPlayer(false, domain.PokerStyleAggressive),
		domain.NewPokerPlayer(false, domain.PokerStyleBluffer),
	}
	poker := domain.NewPoker(domain.NewTrumpCards(config.JokerCount), players, config)
	return &PokerCui{
		pc: controller.NewPokerCuiController(usecase.NewPokerInteractor(poker, presenter.NewPokerCuiPresenter())),
	}
}

// Exec ゲーム実行
func (cui *PokerCui) Exec() {
	fmt.Println(cui.pc.Exec("r"))
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Println("Please enter a command.")
		fmt.Println("q・・・quit")
		fmt.Println("r・・・reset")
		fmt.Println("b [amount]・・・bet (e.g. 'b 20')")
		fmt.Println("c・・・call")
		fmt.Println("ra [amount]・・・raise (e.g. 'ra 30')")
		fmt.Println("ck・・・check")
		fmt.Println("f・・・fold")
		fmt.Println("a・・・all-in")
		fmt.Println("e [0-4]・・・exchange (e.g. 'e 0 2 4' to exchange cards at index 0, 2, 4)")
		fmt.Println("s・・・stand (no exchange)")
		input, exit := readInput(scanner)
		if exit {
			break
		}
		res := cui.pc.Exec(input)
		fmt.Println(res)
		if res == "bye." {
			break
		}
	}
}
