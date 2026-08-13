//go:build !js || !wasm || casino

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// IronCrossGame アイアンクロスゲームインタフェース
type IronCrossGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// PlayerAction 人間の手を処理する
	PlayerAction(action, amount int) error
	// ChooseLine 使う列を決める
	ChooseLine(l domain.IronCrossLine) error
	// NextHand 次のハンドを始める
	NextHand() error
	// CpuPlay CPU の席を進める
	CpuPlay()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.IronCrossConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.IronCrossConfig)

	// GetPhase 現在のフェーズ
	GetPhase() domain.IronCrossPhase
	// GetGameEndFlag ゲーム終了フラグ
	GetGameEndFlag() bool

	// GetPlayers 席の一覧
	GetPlayers() []*domain.IronCrossPlayer
	// GetCross 十字の 5 枚 (伏せている位置は nil)
	GetCross() []*domain.Card
	// GetRevealedCount 開いた枚数
	GetRevealedCount() int
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
	// HumanSeat 人間の席
	HumanSeat() int
	// IsHumanTurn 人間の操作待ちか
	IsHumanTurn() bool
	// IsChoosing 列を選ぶ場面か
	IsChoosing() bool
	// GetHandNumber ハンド数
	GetHandNumber() int
	// GetResults ハンドの結果
	GetResults() []domain.IronCrossResult
	// GetRemainingCards 山の残り枚数
	GetRemainingCards() int
	// WinnerSeat チップがいちばん多い席
	WinnerSeat() int
	// GetHint 助言
	GetHint() *domain.IronCrossHint
}
