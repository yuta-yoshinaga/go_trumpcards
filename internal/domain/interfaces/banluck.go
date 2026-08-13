//go:build !js || !wasm || casino

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// BanLuckGame バンラックゲームインタフェース
type BanLuckGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// PlaceBet 賭け金を置いて配る
	PlaceBet(bet int) error
	// Hit 1 枚引く
	Hit() error
	// Stand 打ち止めにする
	Stand() error
	// NextRound 次のラウンドを始める
	NextRound() error
	// CpuPlay CPU の席を進める
	CpuPlay()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.BanLuckConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.BanLuckConfig)

	// GetPhase 現在のフェーズ
	GetPhase() domain.BanLuckPhase
	// GetGameEndFlag ゲーム終了フラグ
	GetGameEndFlag() bool

	// GetPlayers 席の一覧
	GetPlayers() []*domain.BanLuckPlayer
	// GetHands 席ごとの手札
	GetHands() []*domain.BlackJackHand
	// GetResults 席ごとのラウンド結果
	GetResults() []domain.BanLuckSeatResult
	// GetBankerSeat 親の席
	GetBankerSeat() int
	// GetTurnSeat いま操作している席
	GetTurnSeat() int
	// GetHumanSeat 人間の席
	GetHumanSeat() int
	// IsHumanTurn 人間の操作待ちか
	IsHumanTurn() bool
	// MustHit 人間の席がいま引く義務を負っているか
	MustHit() bool

	// GetRoundNumber ラウンド数
	GetRoundNumber() int
	// GetRemainingCards 山の残り枚数
	GetRemainingCards() int
	// WinnerSeat チップがいちばん多い席
	WinnerSeat() int
	// GetHint 助言
	GetHint() *domain.BanLuckHint
}
