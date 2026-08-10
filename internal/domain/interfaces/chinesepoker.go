//go:build !js || !wasm || casino

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// ChinesePokerGame チャイニーズポーカーゲームインタフェース
type ChinesePokerGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// Bet ベットを行いカードを配る
	Bet(amount int) error
	// SetHands フロント3枚・ミドル5枚を指定してハンドを分割する
	SetHands(frontIndices []int, middleIndices []int) error

	// GetSuggestedArrangement 推奨する13枚の分け方を取得する
	GetSuggestedArrangement() *domain.ChinesePokerSuggestedArrangement
	// GetPlayerCards プレイヤーの13枚を取得する
	GetPlayerCards() []*domain.Card
	// GetDealerCards ディーラーの13枚を取得する
	GetDealerCards() []*domain.Card
	// GetPlayerFront プレイヤーフロントハンドを取得する
	GetPlayerFront() []*domain.Card
	// GetPlayerMiddle プレイヤーミドルハンドを取得する
	GetPlayerMiddle() []*domain.Card
	// GetPlayerBack プレイヤーバックハンドを取得する
	GetPlayerBack() []*domain.Card
	// GetDealerFront ディーラーフロントハンドを取得する
	GetDealerFront() []*domain.Card
	// GetDealerMiddle ディーラーミドルハンドを取得する
	GetDealerMiddle() []*domain.Card
	// GetDealerBack ディーラーバックハンドを取得する
	GetDealerBack() []*domain.Card
	// GetPhase 現在のフェーズを取得する
	GetPhase() int
	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetBet ベット額を取得する
	GetBet() int
	// GetResult ゲーム結果を取得する
	GetResult() domain.GameResult
	// GetFrontResult フロントハンド結果を取得する
	GetFrontResult() domain.GameResult
	// GetMiddleResult ミドルハンド結果を取得する
	GetMiddleResult() domain.GameResult
	// GetBackResult バックハンド結果を取得する
	GetBackResult() domain.GameResult
	// GetPayout 配当を取得する
	GetPayout() int
	// GetPlayerFrontRank プレイヤーフロントハンドランクを取得する
	GetPlayerFrontRank() int
	// GetPlayerMiddleRank プレイヤーミドルハンドランクを取得する
	GetPlayerMiddleRank() int
	// GetPlayerBackRank プレイヤーバックハンドランクを取得する
	GetPlayerBackRank() int
	// GetDealerFrontRank ディーラーフロントハンドランクを取得する
	GetDealerFrontRank() int
	// GetDealerMiddleRank ディーラーミドルハンドランクを取得する
	GetDealerMiddleRank() int
	// GetDealerBackRank ディーラーバックハンドランクを取得する
	GetDealerBackRank() int
	// GetPlayerRoyalty プレイヤーロイヤリティを取得する
	GetPlayerRoyalty() int
	// GetDealerRoyalty ディーラーロイヤリティを取得する
	GetDealerRoyalty() int
	// GetScoop スクープフラグを取得する
	GetScoop() bool
	// GetChips チップを取得する
	GetChips() int
}
