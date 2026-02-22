package ui

import (
	"bufio"
	"fmt"
	"os"

	"github.com/yuta-yoshinaga/go_trumpcards/interface_adapters/controllers"
	"github.com/yuta-yoshinaga/go_trumpcards/interface_adapters/presenters"
	"github.com/yuta-yoshinaga/go_trumpcards/usecases"
)

// SevensCui 7並べCUIクラス
type SevensCui struct {
	sgc *controllers.SevensCuiController
}

// NewSevensCui コンストラクタ
func NewSevensCui() *SevensCui {
	return &SevensCui{
		sgc: controllers.NewSevensCuiController(
			usecases.NewSevensInteractor(presenters.NewSevensCuiPresenter()),
		),
	}
}

// Exec ゲーム実行
func (cui *SevensCui) Exec() {
	fmt.Println(cui.sgc.Exec("r"))
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Println("コマンドを入力してください。")
		fmt.Println("q・・・quit")
		fmt.Println("r [tunnel] [joker=N] [strategy]・・・reset (オプションルール設定)")
		fmt.Println("p [インデックス]・・・カードを出す (インデックスなしでパス)")
		scanner.Scan()
		res := cui.sgc.Exec(scanner.Text())
		fmt.Println(res)
		if res == "bye." {
			break
		}
	}
}
