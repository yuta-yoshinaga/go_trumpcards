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

// DaifugoCui 大富豪CUIクラス
type DaifugoCui struct {
	dgc *controller.DaifugoCuiController
}

// NewDaifugoCui コンストラクタ
func NewDaifugoCui() *DaifugoCui {
	config := domain.DefaultDaifugoConfig()
	players := []*domain.DaifugoPlayer{
		domain.NewDaifugoPlayer(true),
		domain.NewDaifugoPlayer(false),
		domain.NewDaifugoPlayer(false),
		domain.NewDaifugoPlayer(false),
	}
	daifugo := domain.NewDaifugo(domain.NewTrumpCards(config.JokerCount), players, config)
	return &DaifugoCui{
		dgc: controller.NewDaifugoCuiController(
			usecase.NewDaifugoInteractor(daifugo, presenter.NewDaifugoCuiPresenter()),
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
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				fmt.Fprintf(os.Stderr, "入力の読み取り中にエラーが発生しました: %v\n", err)
			}
			break
		}
		res := cui.dgc.Exec(scanner.Text())
		fmt.Println(res)
		if res == "bye." {
			break
		}
	}
}
