//go:build !js || !wasm || casino

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// PokerGame ポーカーゲームインタフェース (マルチプレイヤー)
type PokerGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset() error
	// PlayerAction プレイヤーのベッティングアクションを実行する
	// humanPlayMs: 迷い時間(ms, 0=計測なし)
	PlayerAction(action, amount, humanPlayMs int) error
	// PlayerExchange プレイヤーのカード交換を実行する
	PlayerExchange(indices []int) error
	// PlayerStand カード交換なしで続行する
	PlayerStand() error
	// CalcDrawOdds 交換候補に基づくドローオッズを計算する
	CalcDrawOdds(indices []int) ([]domain.PokerDrawOdds, error)

	// GetPlayers プレイヤー一覧を取得する
	GetPlayers() []*domain.PokerPlayer
	// GetPhase 現在のフェーズを取得する
	GetPhase() int
	// GetPot ポット額を取得する
	GetPot() int
	// GetSidePots サイドポット一覧を取得する
	GetSidePots() []domain.SidePot
	// GetDealerIdx ディーラーインデックスを取得する
	GetDealerIdx() int
	// GetCurrentTurn 現在のターンを取得する
	GetCurrentTurn() int
	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetLastBet 最後のベット額を取得する
	GetLastBet() int
	// GetMinRaise 最小レイズ額を取得する
	GetMinRaise() int
	// GetRaiseCount 現在のレイズ回数を取得する
	GetRaiseCount() int
	// GetAnte アンティ額を取得する
	GetAnte() int
	// GetRoundResults ラウンド結果を取得する
	GetRoundResults() []domain.PokerResult
	// GetCpuActions CPU行動記録を取得する
	GetCpuActions() []domain.PokerCpuAction
	// GetCpuExchanges CPU交換記録を取得する
	GetCpuExchanges() []domain.PokerCpuExchange
	// GetConfig ゲーム設定を取得する
	GetConfig() domain.PokerConfig
	// SetConfig ゲーム設定を変更する
	SetConfig(cfg domain.PokerConfig)
	// GetHumanProfile メタAIプロファイルを取得する
	GetHumanProfile() *domain.BettingHumanProfile
	// ResetProfile メタAIプロファイルをリセットする
	ResetProfile()
	// ExportProfile メタAIプロファイルをエクスポートする
	ExportProfile() interface{}
	// ImportProfile JSONバイトからメタAIプロファイルをインポートする
	ImportProfile(data []byte) error
	// GetEquity ベッティングフェーズでの人間の勝率 (それ以外は nil)
	GetEquity() *domain.HoldemEquityResult
	// GetPotOdds コールに必要な額に対するポットオッズ (0-100)
	GetPotOdds() float64
}
