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

// SevensCui 7並べCUIクラス
type SevensCui struct {
	sgc *controller.SevensCuiController
}

// NewSevensCui コンストラクタ
func NewSevensCui() *SevensCui {
	config := domain.DefaultSevensConfig()
	players := []*domain.SevensPlayer{
		domain.NewSevensPlayer(true),
		domain.NewSevensPlayer(false),
		domain.NewSevensPlayer(false),
		domain.NewSevensPlayer(false),
	}
	sevens := domain.NewSevens(domain.NewTrumpCards(config.JokerCount), players, config)
	return &SevensCui{
		sgc: controller.NewSevensCuiController(
			usecase.NewSevensInteractor(sevens, presenter.NewSevensCuiPresenter()),
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
		fmt.Println("r [tunnel] [joker=N] [strategy] [passes=N]・・・reset (オプションルール設定)")
		fmt.Println("p [インデックス]・・・カードを出す (インデックスなしでパス)")
		scanner.Scan()
		res := cui.sgc.Exec(scanner.Text())
		fmt.Println(res)
		if res == "bye." {
			break
		}
	}
}
