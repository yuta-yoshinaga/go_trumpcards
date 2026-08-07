//go:build !js || !wasm || casino

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// UltimateTexasHoldemGame アルティメット・テキサスホールデムゲームインタフェース
type UltimateTexasHoldemGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// Bet アンテ＋同額のブラインド＋オプションのトリップスをベットする
	Bet(ante, trips int) error
	// Play プレイベットを置く（プリフロップ 3 or 4、フロップ 2、リバー 1）
	Play(multiplier int) error
	// Check プリフロップ・フロップでチェックする
	Check() error
	// Fold リバーでフォールドする
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
	// GetBlindBet ブラインドベット額を取得する
	GetBlindBet() int
	// GetTripsBet トリップス（サイドベット）額を取得する
	GetTripsBet() int
	// GetPlayBet プレイベット額を取得する
	GetPlayBet() int
	// GetFolded リバーでフォールドしたかを取得する
	GetFolded() bool
	// GetResult ゲーム結果を取得する
	GetResult() domain.GameResult
	// GetDealerQualified ディーラークオリファイ状態を取得する
	GetDealerQualified() bool
	// GetAntePayout アンテ配当を取得する
	GetAntePayout() int
	// GetBlindPayout ブラインド配当を取得する
	GetBlindPayout() int
	// GetPlayPayout プレイベット配当を取得する
	GetPlayPayout() int
	// GetTripsPayout トリップス配当を取得する
	GetTripsPayout() int
	// GetTotalPayout 合計配当を取得する
	GetTotalPayout() int
	// GetPlayerHandRank プレイヤーハンドランクを取得する
	GetPlayerHandRank() int
	// RecommendPlay 現在のフェーズでの推奨アクションを取得する
	RecommendPlay() string
	// GetDealerHandRank ディーラーハンドランクを取得する
	GetDealerHandRank() int
	// GetPlayerBest プレイヤー最良5枚を取得する
	GetPlayerBest() []*domain.Card
	// GetDealerBest ディーラー最良5枚を取得する
	GetDealerBest() []*domain.Card
	// GetChips チップを取得する
	GetChips() int
}
