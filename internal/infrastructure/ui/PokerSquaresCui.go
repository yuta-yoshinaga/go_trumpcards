package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewPokerSquaresCui はコンストラクタ。
func NewPokerSquaresCui() *genericCuiGame {
	ps := domain.NewPokerSquares(domain.NewTrumpCards(0))
	pc := controller.NewPokerSquaresCuiController(usecase.NewPokerSquaresInteractor(ps, new(presenter.PokerSquaresCuiPresenter)))
	return newCuiGame(pc, []string{
		"Poker Squares (ポーカー・スクエアズ)",
		"",
		i18n.T("gameCommands"),
		"  p <row> <col>            カードを配置 (0-4)",
		"  u                        アンドゥ",
		"  g                        ギブアップ",
		"  l                        action log",
		"",
		i18n.T("session"),
		i18n.T("resetEntry"),
		i18n.T("quitEntry"),
		i18n.T("helpEntry"),
	})
}
