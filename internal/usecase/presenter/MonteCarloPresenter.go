package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MonteCarloPresenter はモンテカルロ・ソリティアのプレゼンターインタフェース。
type MonteCarloPresenter interface {
	GamePresenter[interfaces.MonteCarloGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.MonteCarloGame) string
}
