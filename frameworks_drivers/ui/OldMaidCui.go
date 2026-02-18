package ui

import (
	"bufio"
	"fmt"
	"os"

	"github.com/yuta-yoshinaga/go_trumpcards/interface_adapters/controllers"
	"github.com/yuta-yoshinaga/go_trumpcards/interface_adapters/presenters"
	"github.com/yuta-yoshinaga/go_trumpcards/usecases"
)

// OldMaidCui ババ抜きCUIクラス
type OldMaidCui struct {
	omc *controllers.OldMaidCuiController
}

// NewOldMaidCui コンストラクタ
func NewOldMaidCui() *OldMaidCui {
	return &OldMaidCui{
		omc: controllers.NewOldMaidCuiController(
			usecases.NewOldMaidInteractor(presenters.NewOldMaidCuiPresenter()),
		),
	}
}

// Exec ゲーム実行
func (cui *OldMaidCui) Exec() {
	fmt.Println(cui.omc.Exec("r"))
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Println("コマンドを入力してください。")
		fmt.Println("q・・・quit")
		fmt.Println("r・・・reset")
		fmt.Println("d・・・draw (カードを引く)")
		scanner.Scan()
		res := cui.omc.Exec(scanner.Text())
		fmt.Println(res)
		if res == "bye." {
			break
		}
	}
}
