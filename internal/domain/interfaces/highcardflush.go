//go:build !js || !wasm || extra4

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// HighCardFlushGame ハイカードフラッシュゲームインタフェース
type HighCardFlushGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// Bet アンテ＋オプションのサイドベットを置きカードを配る
	Bet(ante, flushBonus, straightFlush int) error
	// Raise レイズベットを置いて勝負する
	Raise(multiplier int) error
	// Fold フォールドする
	Fold() error
	// MaxRaiseMultiplier 現在のプレイヤーのフラッシュ長に応じた最大レイズ倍率を取得する
	MaxRaiseMultiplier() int

	// GetPlayerHand プレイヤーハンドを取得する
	GetPlayerHand() []*domain.Card
	// GetDealerHand ディーラーハンドを取得する
	GetDealerHand() []*domain.Card
	// GetPhase 現在のフェーズを取得する
	GetPhase() int
	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetAnteBet アンテベット額を取得する
	GetAnteBet() int
	// GetFlushBonusBet Flush Bonus ベット額を取得する
	GetFlushBonusBet() int
	// GetStraightFlushBet Straight Flush Bonus ベット額を取得する
	GetStraightFlushBet() int
	// GetRaiseBet レイズベット額を取得する
	GetRaiseBet() int
	// GetResult ゲーム結果を取得する
	GetResult() domain.GameResult
	// GetAntePayout アンテ配当を取得する
	GetAntePayout() int
	// GetRaisePayout レイズ配当を取得する
	GetRaisePayout() int
	// GetFlushBonusPayout Flush Bonus 配当を取得する
	GetFlushBonusPayout() int
	// GetStraightFlushPayout Straight Flush Bonus 配当を取得する
	GetStraightFlushPayout() int
	// GetTotalPayout 合計配当を取得する
	GetTotalPayout() int
	// GetDealerQualified ディーラークオリファイを取得する
	GetDealerQualified() bool
	// GetPlayerFlushLen プレイヤーの最長フラッシュ長を取得する
	GetPlayerFlushLen() int
	// GetPlayerFlushSuit 上の長さを数えたスートを取得する
	GetPlayerFlushSuit() int
	// GetDealerFlushLen ディーラーの最長フラッシュ長を取得する
	GetDealerFlushLen() int
	// GetPlayerStraightFlushLen プレイヤーの最長ストレートフラッシュ長を取得する
	GetPlayerStraightFlushLen() int
	// GetChips チップを取得する
	GetChips() int
}
