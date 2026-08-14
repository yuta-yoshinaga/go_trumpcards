//go:build !js || !wasm || casino

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// FreeBetBlackjackGame フリーベット・ブラックジャックゲームインタフェース
type FreeBetBlackjackGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// PlaceBet アンティを置いて配る
	PlaceBet(ante int) error
	// Hit 1 枚引く
	Hit() error
	// Stand その手札を打ち止めにする
	Stand() error
	// FreeDouble ハウス持ちで賭け金を倍にする
	FreeDouble() error
	// FreeSplit ハウス持ちで手札を分ける
	FreeSplit() error
	// NextRound 次のラウンドを始める
	NextRound() error

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.FreeBetBlackjackConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.FreeBetBlackjackConfig)

	// GetPhase 現在のフェーズ
	GetPhase() domain.FreeBetPhase
	// GetGameEndFlag ゲーム終了フラグ
	GetGameEndFlag() bool

	// GetHands プレイヤーの手札
	GetHands() []*domain.BlackJackHand
	// GetHandCount 手札の数
	GetHandCount() int
	// GetFreeBets 手札ごとのハウス出資額
	GetFreeBets() []int
	// GetFreeBet 手札 idx のハウス出資額
	GetFreeBet(idx int) int
	// GetActiveHandIdx いま操作している手札の位置
	GetActiveHandIdx() int
	// GetDealerCards ディーラーの札
	GetDealerCards() []*domain.Card
	// GetDealerScore ディーラーの点数
	GetDealerScore() int
	// IsDealerPushed22 ディーラーが 22 でバストしたか
	IsDealerPushed22() bool

	// CanFreeDouble いま無料ダブルできるか
	CanFreeDouble() bool
	// CanFreeSplit いま無料スプリットできるか
	CanFreeSplit() bool

	// GetAnteBet アンティ額
	GetAnteBet() int
	// GetResults 手札ごとの決着
	GetResults() []domain.FreeBetResult
	// GetPayout このラウンドで戻ってきた総額
	GetPayout() int
	// GetChips 保有チップ数
	GetChips() int
	// GetRoundNumber ラウンド数
	GetRoundNumber() int
	// GetRemainingCards シューの残り枚数
	GetRemainingCards() int
	// GetHint 助言
	GetHint() *domain.FreeBetHint
}
