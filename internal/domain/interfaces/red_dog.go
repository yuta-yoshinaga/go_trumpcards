//go:build !js || !wasm || extra4

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// RedDogGame レッドドッグゲームインタフェース
type RedDogGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// Bet アンテをベットしカードを2枚配る
	Bet(amount int) error
	// ResolveInitial 初手2枚を評価しフェーズ遷移
	ResolveInitial()
	// Raise 3枚目に向けてレイズする
	Raise(amount int) error
	// Stay レイズせず3枚目を引く
	Stay() error
	// ResolveThird 3枚目を評価しペイアウト
	ResolveThird()

	// GetInitialCards 初手2枚を取得
	GetInitialCards() []*domain.Card
	// GetThirdCard 3枚目を取得
	GetThirdCard() *domain.Card
	// GetPhase 現在のフェーズ
	GetPhase() int
	// GetGameEndFlag ゲーム終了フラグ
	GetGameEndFlag() bool
	// GetAnte アンテ額
	GetAnte() int
	// GetRaise レイズ額
	GetRaise() int
	// GetSpread スプレッド
	GetSpread() int
	// GetResult 勝敗結果
	GetResult() domain.GameResult
	// GetTotalPayout 合計配当
	GetTotalPayout() int
	// GetChips チップ
	GetChips() int
}
