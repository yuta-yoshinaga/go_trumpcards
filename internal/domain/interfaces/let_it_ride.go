//go:build !js || !wasm || casino

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// LetItRideGame レット・イット・ライドゲームインタフェース
type LetItRideGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// Bet ベットを行いカードを配る
	Bet(amount int) error
	// Pull ベットを取り下げる
	Pull() error
	// LetItRideAction ベットをそのままにする
	LetItRideAction() error
	// GetPullPreview Pull を実行したときに戻る額とリスクの増減を取得する
	GetPullPreview() *domain.LetItRidePullPreview

	// GetPlayerHand プレイヤーハンドを取得する
	GetPlayerHand() []*domain.Card
	// GetCommunityCards コミュニティカードを取得する
	GetCommunityCards() []*domain.Card
	// GetPhase 現在のフェーズを取得する
	GetPhase() int
	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetBetAmount 1口あたりのベット額を取得する
	GetBetAmount() int
	// GetBet1Active ベット1アクティブ状態を取得する
	GetBet1Active() bool
	// GetBet2Active ベット2アクティブ状態を取得する
	GetBet2Active() bool
	// GetBet3Active ベット3アクティブ状態を取得する
	GetBet3Active() bool
	// GetResult ゲーム結果を取得する
	GetResult() domain.GameResult
	// GetHandRank ハンドランクを取得する
	GetHandRank() int
	// GetBet1Payout ベット1配当を取得する
	GetBet1Payout() int
	// GetBet2Payout ベット2配当を取得する
	GetBet2Payout() int
	// GetBet3Payout ベット3配当を取得する
	GetBet3Payout() int
	// GetTotalPayout 合計配当を取得する
	GetTotalPayout() int
	// GetChips チップを取得する
	GetChips() int
}
