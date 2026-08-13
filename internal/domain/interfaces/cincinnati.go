//go:build !js || !wasm || casino

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// CincinnatiGame シンシナティゲームインタフェース
type CincinnatiGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// PlayerAction 人間の手を処理する
	PlayerAction(action, amount int) error
	// NextHand 次のハンドを始める
	NextHand() error
	// CpuPlay CPU の席を進める
	CpuPlay()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.CincinnatiConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.CincinnatiConfig)

	// GetPhase 現在のフェーズ
	GetPhase() domain.CincinnatiPhase
	// GetGameEndFlag ゲーム終了フラグ
	GetGameEndFlag() bool

	// GetPlayers 席の一覧
	GetPlayers() []*domain.CincinnatiPlayer
	// GetCommunityCards 表向きのコミュニティ
	GetCommunityCards() []*domain.Card
	// GetRevealedCount 公開済みのコミュニティ枚数
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
	// GetHandNumber ハンド数
	GetHandNumber() int
	// GetResults ハンドの結果
	GetResults() []domain.CincinnatiResult
	// GetRemainingCards 山の残り枚数
	GetRemainingCards() int
	// WinnerSeat チップがいちばん多い席
	WinnerSeat() int
	// GetHint 助言
	GetHint() *domain.CincinnatiHint
}
