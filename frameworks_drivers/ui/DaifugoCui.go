package ui

import (
	"bufio"
	"fmt"
	"os"

	"github.com/yuta-yoshinaga/go_trumpcards/interface_adapters/controllers"
	"github.com/yuta-yoshinaga/go_trumpcards/interface_adapters/presenters"
	"github.com/yuta-yoshinaga/go_trumpcards/usecases"
)

// DaifugoCui 大富豪CUIクラス
type DaifugoCui struct {
	dgc *controllers.DaifugoCuiController
}

// NewDaifugoCui コンストラクタ
func NewDaifugoCui() *DaifugoCui {
	return &DaifugoCui{
		dgc: controllers.NewDaifugoCuiController(
			usecases.NewDaifugoInteractor(presenters.NewDaifugoCuiPresenter()),
		),
	}
}

// Exec ゲーム実行
func (cui *DaifugoCui) Exec() {
	fmt.Println(cui.dgc.Exec("r"))
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Println("コマンドを入力してください。")
		fmt.Println("q・・・quit")
		fmt.Println("r・・・reset")
		fmt.Println("p [インデックス...]・・・カードを出す (インデックスなしでパス)")
		scanner.Scan()
		res := cui.dgc.Exec(scanner.Text())
		fmt.Println(res)
		if res == "bye." {
			break
		}
	}
}
