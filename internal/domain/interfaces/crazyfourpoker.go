//go:build !js || !wasm || extra4

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// CrazyFourPokerGame クレイジー 4 ポーカーゲームインタフェース
type CrazyFourPokerGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// PlaceBet アンティ (と同額の Super Bonus)、任意の Queens Up を置いて配る
	PlaceBet(ante, queensUp int) error
	// Play プレイベットを置いて決着させる
	Play(multiplier int) error
	// Fold 降りる
	Fold() error
	// NextRound 次のラウンドを始める
	NextRound() error

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.CrazyFourPokerConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.CrazyFourPokerConfig)

	// GetPhase 現在のフェーズ
	GetPhase() domain.CrazyFourPokerPhase
	// GetGameEndFlag ゲーム終了フラグ
	GetGameEndFlag() bool

	// GetPlayerHand 配られた 5 枚
	GetPlayerHand() []*domain.Card
	// GetDealerHand ディーラーの 5 枚
	GetDealerHand() []*domain.Card
	// GetPlayerBest プレイヤーの最良 4 枚
	GetPlayerBest() []*domain.Card
	// GetDealerBest ディーラーの最良 4 枚
	GetDealerBest() []*domain.Card
	// GetPlayerHandRank プレイヤーの役
	GetPlayerHandRank() int
	// GetDealerHandRank ディーラーの役
	GetDealerHandRank() int

	// PlayerHasAcesOrBetter エースのペア以上か (3 倍の条件)
	PlayerHasAcesOrBetter() bool
	// MaxPlayMultiplier いま置ける上限倍率
	MaxPlayMultiplier() int
	// PlayerQualifies プレイヤーの手がキング以上か
	PlayerQualifies() bool
	// DealerQualifies ディーラーが成立しているか
	DealerQualifies() bool

	// GetAnteBet アンティ額
	GetAnteBet() int
	// GetSuperBet Super Bonus 額 (常にアンティと同額)
	GetSuperBet() int
	// GetQueensUpBet Queens Up 額
	GetQueensUpBet() int
	// GetPlayBet プレイベット額
	GetPlayBet() int
	// GetPlayMultiplier プレイベットの倍率
	GetPlayMultiplier() int
	// GetResult ラウンドの決着
	GetResult() domain.CrazyFourPokerResult
	// GetPayout このラウンドで戻ってきた総額
	GetPayout() int
	// GetChips 保有チップ数
	GetChips() int
	// GetMinTotalWager 1 ラウンドに最低限必要なチップ
	GetMinTotalWager() int
	// GetRoundNumber ラウンド数
	GetRoundNumber() int
	// GetRemainingCards デッキの残り枚数
	GetRemainingCards() int
	// GetHint 助言
	GetHint() *domain.CrazyFourPokerHint
}
