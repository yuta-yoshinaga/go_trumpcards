//go:build !js || !wasm || casino

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// DoubleAttackBlackjackGame 追加ベット・ブラックジャックゲームインタフェース
type DoubleAttackBlackjackGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// PlaceBet アンティと任意の Bust It を置いて配る
	PlaceBet(ante, bustIt int) error
	// Attack アップカードを見てから追加ベットを置く (0 で見送り)
	Attack(amount int) error
	// Hit 1 枚引く
	Hit() error
	// Stand その手札を打ち止めにする
	Stand() error
	// Double 賭け金を倍にして 1 枚だけ引く
	Double() error
	// Split 同じ数字 2 枚を分ける
	Split() error
	// NextRound 次のラウンドを始める
	NextRound() error

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.DoubleAttackBlackjackConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.DoubleAttackBlackjackConfig)

	// GetPhase 現在のフェーズ
	GetPhase() domain.DoubleAttackPhase
	// GetGameEndFlag ゲーム終了フラグ
	GetGameEndFlag() bool

	// GetHands プレイヤーの手札 (スプリットで増える)
	GetHands() []*domain.BlackJackHand
	// GetHandCount 手札の数
	GetHandCount() int
	// GetActiveHandIdx いま操作している手札の位置
	GetActiveHandIdx() int
	// GetDealerCards ディーラーの札 (追加ベット前はアップカードのみ)
	GetDealerCards() []*domain.Card
	// GetDealerScore ディーラーの点数
	GetDealerScore() int
	// IsDealerHoleDealt ディーラーの 2 枚目が配られたか
	IsDealerHoleDealt() bool

	// MaxAttackBet 追加ベットの上限 (アンティまで)
	MaxAttackBet() int
	// CanDouble いまダブルできるか
	CanDouble() bool
	// CanSplit いまスプリットできるか
	CanSplit() bool

	// GetAnteBet アンティ額
	GetAnteBet() int
	// GetAttackBet 追加ベット額
	GetAttackBet() int
	// GetBustItBet Bust It 額
	GetBustItBet() int
	// GetResults 手札ごとの決着
	GetResults() []domain.DoubleAttackResult
	// GetPayout このラウンドで戻ってきた総額
	GetPayout() int
	// GetBustItPayout Bust It からの払い戻し
	GetBustItPayout() int
	// GetChips 保有チップ数
	GetChips() int
	// GetRoundNumber ラウンド数
	GetRoundNumber() int
	// GetRemainingCards シューの残り枚数
	GetRemainingCards() int
	// GetHint 助言
	GetHint() *domain.DoubleAttackHint
}
