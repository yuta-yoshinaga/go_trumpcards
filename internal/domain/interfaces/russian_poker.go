package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// RussianPokerGame ロシアンポーカーゲームインタフェース
type RussianPokerGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// Bet アンテベットを行いカードを配る
	Bet(ante int) error
	// Exchange 指定インデックスのカードを交換する（手数料はante×枚数）
	Exchange(indices []int) error
	// Buy6th 6枚目のカードを購入する（手数料はante同額）
	Buy6th() error
	// Select 6枚の手札から1枚を捨てて5枚にする
	Select(discardIndex int) error
	// Play プレイベットを置いて勝負する
	Play() error
	// Fold フォールドする
	Fold() error
	// ForceExchange ディーラーの最高カードを交換させる
	ForceExchange() error
	// Decline 強制クオリファイを辞退する
	Decline() error

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
	// GetPlayBet プレイベット額を取得する
	GetPlayBet() int
	// GetExchangeCount 交換した枚数を取得する
	GetExchangeCount() int
	// GetExchangeFee 徴収した交換手数料を取得する
	GetExchangeFee() int
	// GetBought6th 6枚目を購入したかどうかを取得する
	GetBought6th() bool
	// GetBuy6thFee 6枚目購入手数料を取得する
	GetBuy6thFee() int
	// GetForceExchanged 強制クオリファイを実行したかどうかを取得する
	GetForceExchanged() bool
	// GetForceExchangeFee 強制クオリファイ手数料を取得する
	GetForceExchangeFee() int
	// GetResult ゲーム結果を取得する
	GetResult() domain.GameResult
	// GetAntePayout アンテ配当を取得する
	GetAntePayout() int
	// GetPlayPayout プレイ配当を取得する
	GetPlayPayout() int
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
