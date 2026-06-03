//go:build !js || !wasm || casino

package domain

// Big O (5 Card Omaha) は通常のオマハ (ホールカード4枚) のホールカードを
// 5枚に増やしたバリアントである。役を作る際は通常のオマハと同じく
// 「手札から必ず2枚」「コミュニティカードから必ず3枚」を使う。手札が1枚
// 増えることでコンビネーションが C(4,2)=6 通りから C(5,2)=10 通りに増え、
// より強力なナッツが出やすくなる。
//
// 実装は既存の Omaha エンジンをそのまま流用し、配布枚数 (holeCards) のみを
// 5 にパラメータ化している (issue #1992)。ハンド評価 (EvalBestHand /
// EvalBestLowHand) は combinations() がホールカード枚数に依存しないため、
// 追加の評価ロジックは不要。

// NewDefaultBigO は Big O (5 Card Omaha) をデフォルトのテーブルサイズと
// DefaultOmahaConfig で生成する。CUI / Web / Worker の構築サイトで共有する
// 単一の生成元 (NewDefaultOmaha の 5 枚版)。
func NewDefaultBigO() *Omaha {
	cfg := DefaultOmahaConfig()
	o := NewOmaha(NewTrumpCards(0), NewOmahaPlayersForTable(cfg.TableSize), cfg)
	o.holeCards = bigOHoleCards
	return o
}

// NewDefaultBigOHiLo は Big O を 8 or Better (Hi-Lo スプリット) として
// デフォルト設定で生成する。CUI / Web / Worker 構築サイトの単一の生成元。
func NewDefaultBigOHiLo() *Omaha {
	o := NewDefaultBigO()
	o.hiLo = true
	return o
}

// bigOHoleCards は Big O のホールカード枚数。
const bigOHoleCards = 5
