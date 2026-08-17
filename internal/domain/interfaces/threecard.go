//go:build !js || !wasm || casino

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// ThreeCardGame スリーカードポーカーゲームインタフェース
type ThreeCardGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// Bet アンテベットを行いカードを配る
	Bet(ante, pairPlus int) error
	// Rebet 直前のラウンドと同じ額で賭け直す
	Rebet() error
	// Play プレイベットを置いて勝負する
	Play() error
	// Fold フォールドする
	Fold() error

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
	// GetPairPlusBet ペアプラスベット額を取得する
	GetPairPlusBet() int
	// GetPlayBet プレイベット額を取得する
	GetPlayBet() int
	// GetResult ゲーム結果を取得する
	GetResult() domain.GameResult
	// GetAntePayout アンテ配当を取得する
	GetAntePayout() int
	// GetPlayPayout プレイ配当を取得する
	GetPlayPayout() int
	// GetAnteBonusPayout アンテボーナス配当を取得する
	GetAnteBonusPayout() int
	// GetPairPlusPayout ペアプラス配当を取得する
	GetPairPlusPayout() int
	// GetTotalPayout 合計配当を取得する
	GetTotalPayout() int
	// GetDealerQualified ディーラークオリファイを取得する
	GetDealerQualified() bool
	// GetPlayerHandRank プレイヤーハンドランクを取得する
	GetPlayerHandRank() int
	// GetDealerHandRank ディーラーハンドランクを取得する
	GetDealerHandRank() int
	// GetChips チップを取得する
	GetChips() int
}
