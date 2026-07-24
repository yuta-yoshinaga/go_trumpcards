//go:build !js || !wasm || casino

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// CasinoHoldemGame カジノホールデムゲームインタフェース
type CasinoHoldemGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// Bet アンテベットとオプションの AA ボーナスサイドベットを行う
	Bet(ante, bonus int) error
	// Call フロップ後にコール（2×アンテ）してターン／リバーを公開する
	Call() error
	// Fold フロップ後にフォールドする
	Fold() error

	// GetPlayerHand プレイヤーホールカードを取得する
	GetPlayerHand() []*domain.Card
	// GetDealerHand ディーラーホールカードを取得する
	GetDealerHand() []*domain.Card
	// GetCommunity コミュニティカードを取得する
	GetCommunity() []*domain.Card
	// GetPhase 現在のフェーズを取得する
	GetPhase() int
	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetAnteBet アンテベット額を取得する
	GetAnteBet() int
	// GetBonusBet AA ボーナスベット額を取得する
	GetBonusBet() int
	// GetCallBet コールベット額を取得する
	GetCallBet() int
	// GetResult ゲーム結果を取得する
	GetResult() domain.GameResult
	// GetDealerQualify ディーラークオリファイフラグを取得する
	GetDealerQualify() bool
	// GetAntePayout アンテ配当を取得する
	GetAntePayout() int
	// GetCallPayout コール配当を取得する
	GetCallPayout() int
	// GetBonusPayout AA ボーナス配当を取得する
	GetBonusPayout() int
	// GetTotalPayout 合計配当を取得する
	GetTotalPayout() int
	// GetPlayerHandRank プレイヤーハンドランクを取得する
	GetPlayerHandRank() int
	// GetDealerHandRank ディーラーハンドランクを取得する
	GetDealerHandRank() int
	// GetPlayerBest プレイヤー最良5枚を取得する
	GetPlayerBest() []*domain.Card
	// GetDealerBest ディーラー最良5枚を取得する
	GetDealerBest() []*domain.Card
	// GetChips チップを取得する
	GetChips() int
	// RecommendCall はフロップ後にコールを推奨するかを返す
	RecommendCall() bool
}
