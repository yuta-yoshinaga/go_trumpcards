//go:build !js || !wasm || casino

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// BaseballPokerGame ベースボールポーカーゲームインタフェース
type BaseballPokerGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// PlayerAction 人間の手を処理する
	PlayerAction(action, amount int) error
	// AnswerBuyIn 買い増しの返事を処理する
	AnswerBuyIn(answer int) error
	// NextHand 次のハンドを始める
	NextHand() error
	// CpuPlay CPU の席を進める
	CpuPlay()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.BaseballPokerConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.BaseballPokerConfig)

	// GetPhase 現在のフェーズ
	GetPhase() domain.BaseballPhase
	// GetGameEndFlag ゲーム終了フラグ
	GetGameEndFlag() bool

	// GetPlayers 席の一覧
	GetPlayers() []*domain.BaseballPokerPlayer
	// GetStreet 配り終えた表札の数
	GetStreet() int
	// GetPot ポット
	GetPot() int
	// GetCurrentBet このラウンドの現在の賭け額
	GetCurrentBet() int
	// GetToCall 人間がコールに要する額
	GetToCall() int
	// GetRaiseCount このラウンドのレイズ回数
	GetRaiseCount() int
	// CanRaise いまレイズできるか
	CanRaise() bool
	// GetTurnSeat いまの手番
	GetTurnSeat() int
	// GetBuyerSeat 買い増しを迫られている席 (-1 なら誰もいない)
	GetBuyerSeat() int
	// GetBuyCost 買い増しの額
	GetBuyCost() int
	// HumanSeat 人間の席
	HumanSeat() int
	// IsHumanTurn 人間の操作待ちか
	IsHumanTurn() bool
	// IsHumanBuying 人間が買い増しを迫られているか
	IsHumanBuying() bool
	// GetHandNumber ハンド数
	GetHandNumber() int
	// GetResults ハンドの結果
	GetResults() []domain.BaseballResult
	// GetRemainingCards 山の残り枚数
	GetRemainingCards() int
	// WinnerSeat チップがいちばん多い席
	WinnerSeat() int
	// GetHint 助言
	GetHint() *domain.BaseballHint
}
