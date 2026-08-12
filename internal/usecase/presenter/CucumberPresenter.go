//go:build !js || !wasm || classic

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// CucumberPresenter キューカンバープレゼンターインタフェース
type CucumberPresenter interface {
	GamePresenter[interfaces.CucumberGame]
	// HintOutput ヒント情報を出力する
	HintOutput(s interfaces.CucumberGame) string
}
