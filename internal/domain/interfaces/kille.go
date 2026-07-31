//go:build !js || !wasm || extra3

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// KilleGame キッレ (Kille) ゲームインタフェース
type KilleGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のラウンドを配る
	NextRound() error
	// Exchange 指定席が隣と (ディーラーなら山と) 交換を仕掛ける
	Exchange(player int) error
	// Satisfied 指定席が交換しないと宣言する
	Satisfied(player int) error
	// Reenter 脱落した席が買い戻す
	Reenter(seat int) error
	// CpuPlay CPUプレイヤーが1アクション実行する
	CpuPlay()
	// KilleCpuReenterDecide 脱落した CPU が買い戻すかを返す
	KilleCpuReenterDecide(seat int) bool

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.KilleConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.KilleConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.KillePhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// GetRoundNumber 現在のラウンド番号を取得する
	GetRoundNumber() int
	// GetCurrentPlayerIdx 現在のプレイヤーインデックスを取得する
	GetCurrentPlayerIdx() int
	// GetDealerIdx ディーラーインデックスを取得する
	GetDealerIdx() int
	// GetStockCount 残り山札の枚数を取得する
	GetStockCount() int
	// GetPot 現在のポットを取得する
	GetPot() int
	// GetWinnerIdx 勝者インデックスを取得する
	GetWinnerIdx() int
	// GetPlayers 全プレイヤーを取得する
	GetPlayers() []*domain.KillePlayer
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(idx int) *domain.KillePlayer
	// GetEvents このラウンドの交換記録を取得する
	GetEvents() []*domain.KilleEvent
	// GetLoserIdxs 直近ラウンドで脱落した席を取得する
	GetLoserIdxs() []int
	// KilleStrength 指定席の実効的な強さを取得する
	KilleStrength(seat int) int
	// KilleReentryCost 指定席が次に買い戻すのに要る額を取得する
	KilleReentryCost(seat int) int
}
