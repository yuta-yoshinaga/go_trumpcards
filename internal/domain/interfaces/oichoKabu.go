//go:build !js || !wasm || extra

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// OichoKabuGame おいちょかぶゲームインタフェース
type OichoKabuGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// Bet 掛け金を置き、子・親に 2 枚ずつ配る
	Bet(amount int) error
	// Draw 子が 3 枚目を引き、勝負を確定する
	Draw() error
	// Stand 子が引かずに勝負を確定する
	Stand() error

	// GetPlayerHand 子（human）の手
	GetPlayerHand() []*domain.Card
	// GetBankerHand 親（banker/胴元）の手
	GetBankerHand() []*domain.Card
	// GetPlayerRank 子の目（点数合計 % 10）
	GetPlayerRank() int
	// GetBankerRank 親の目（点数合計 % 10）
	GetBankerRank() int
	// GetPhase 現在のフェーズ
	GetPhase() int
	// GetGameEndFlag ゲーム終了フラグ
	GetGameEndFlag() bool
	// GetBet 掛け金
	GetBet() int
	// GetResult 勝敗結果
	GetResult() domain.GameResult
	// GetTotalPayout 合計配当
	GetTotalPayout() int
	// GetChips チップ
	GetChips() int
}
