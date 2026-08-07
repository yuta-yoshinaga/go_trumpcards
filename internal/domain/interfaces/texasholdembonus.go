//go:build !js || !wasm || casino

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// TexasHoldemBonusGame テキサスホールデムボーナスポーカーゲームインタフェース
type TexasHoldemBonusGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// Bet アンテベットとオプションのボーナスサイドベットを行う
	Bet(ante, bonus int) error
	// Play プリフロップでフロップベット（2×アンテ）を置く
	Play() error
	// Fold プリフロップでフォールドする
	Fold() error
	// Check フロップ後またはターン後にチェックする
	Check() error
	// Raise フロップ後またはターン後に1×アンテをベットする
	Raise() error

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
	// GetNextBetCost 現在のフェーズで Play / Raise に必要なチップ額を取得する
	GetNextBetCost() int
	// GetBonusBet ボーナスサイドベット額を取得する
	GetBonusBet() int
	// GetFlopBet フロップベット額を取得する
	GetFlopBet() int
	// GetTurnBet ターンベット額を取得する
	GetTurnBet() int
	// GetRiverBet リバーベット額を取得する
	GetRiverBet() int
	// GetTotalPlayBet フロップ＋ターン＋リバーの合計ベット額を取得する
	GetTotalPlayBet() int
	// GetResult ゲーム結果を取得する
	GetResult() domain.GameResult
	// GetAntePayout アンテ配当を取得する
	GetAntePayout() int
	// GetPlayPayout プレイベット配当合計を取得する
	GetPlayPayout() int
	// GetBonusPayout ボーナスサイドベット配当を取得する
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
}
