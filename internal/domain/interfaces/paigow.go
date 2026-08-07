//go:build !js || !wasm || casino

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// PaiGowGame パイガオポーカーゲームインタフェース
type PaiGowGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// Bet ベットを行いカードを配る
	Bet(amount int) error
	// SetHands ローハンドの2枚を指定してハンドを分割する
	SetHands(lowIdx0, lowIdx1 int) error
	// AutoSetHands ハウスウェイの分割をそのまま適用する
	AutoSetHands() error
	// GetHint セットハンドフェーズでの推奨分割を取得する
	GetHint() *domain.PaiGowHint

	// GetPlayerCards プレイヤーの7枚を取得する
	GetPlayerCards() []*domain.Card
	// GetDealerCards ディーラーの7枚を取得する
	GetDealerCards() []*domain.Card
	// GetPlayerHighHand プレイヤーハイハンドを取得する
	GetPlayerHighHand() []*domain.Card
	// GetPlayerLowHand プレイヤーローハンドを取得する
	GetPlayerLowHand() []*domain.Card
	// GetDealerHighHand ディーラーハイハンドを取得する
	GetDealerHighHand() []*domain.Card
	// GetDealerLowHand ディーラーローハンドを取得する
	GetDealerLowHand() []*domain.Card
	// GetPhase 現在のフェーズを取得する
	GetPhase() int
	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetBet ベット額を取得する
	GetBet() int
	// GetResult ゲーム結果を取得する
	GetResult() domain.GameResult
	// GetHighHandResult ハイハンド結果を取得する
	GetHighHandResult() domain.GameResult
	// GetLowHandResult ローハンド結果を取得する
	GetLowHandResult() domain.GameResult
	// GetPayout 配当を取得する
	GetPayout() int
	// GetCommission コミッションを取得する
	GetCommission() int
	// GetPlayerHighRank プレイヤーハイハンドランクを取得する
	GetPlayerHighRank() int
	// GetPlayerLowRank プレイヤーローハンドランクを取得する
	GetPlayerLowRank() int
	// GetDealerHighRank ディーラーハイハンドランクを取得する
	GetDealerHighRank() int
	// GetDealerLowRank ディーラーローハンドランクを取得する
	GetDealerLowRank() int
	// GetChips チップを取得する
	GetChips() int
}
