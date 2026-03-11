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
	cfg := domain.DefaultHoldemConfig()
	holdem := domain.NewHoldem(domain.NewTrumpCards(0), domain.NewPlayersForTable(cfg.TableSize), cfg)
	return &HoldemCui{
		hc: controller.NewHoldemCuiController(usecase.NewHoldemInteractor(holdem, new(presenter.HoldemCuiPresenter))),
	}
}

// Controller returns the game controller.
func (cui *HoldemCui) Controller() CuiExecer { return cui.hc }

// HelpLines returns the game's help lines.
func (cui *HoldemCui) HelpLines() []string {
	return []string{
		"--- Commands ---",
		"q・・・quit",
		"r・・・reset",
		"f・・・fold",
		"ck・・・check",
		"c・・・call",
		"b [amount]・・・bet (e.g. 'b 20')",
		"ra [amount]・・・raise (e.g. 'ra 30')",
		"a・・・allin",
		"bl [0-2]・・・betting limit (0=Fixed, 1=PotLimit, 2=NoLimit)",
		"tm [0-1]・・・tournament mode (0=OFF, 1=ON)",
		"sb [amount]・・・small blind (>=1)",
		"bb [amount]・・・big blind (>=2)",
		"lh [hands]・・・blind level-up hands (>=1)",
		"ts [4|6|9]・・・table size (4-max, 6-max, 9-max)",
		"rb・・・rebuy (accept rebuy)",
		"sr・・・skip rebuy (decline rebuy)",
		"ad・・・addon (accept addon)",
		"sa・・・skip addon (decline addon)",
		"----------------",
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
	fmt.Println("tm [0-1]・・・tournament mode (0=OFF, 1=ON)")
	fmt.Println("sb [amount]・・・small blind (>=1)")
	fmt.Println("bb [amount]・・・big blind (>=2)")
	fmt.Println("lh [hands]・・・blind level-up hands (>=1)")
	fmt.Println("ts [4|6|9]・・・table size (4-max, 6-max, 9-max)")
	fmt.Println("rb・・・rebuy (accept rebuy)")
	fmt.Println("sr・・・skip rebuy (decline rebuy)")
	fmt.Println("ad・・・addon (accept addon)")
	fmt.Println("sa・・・skip addon (decline addon)")
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
