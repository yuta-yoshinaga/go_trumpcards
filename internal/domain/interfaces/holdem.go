package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// HoldemGame テキサスホールデムゲームインタフェース
type HoldemGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset() error
	// PlayerAction プレイヤーのベッティングアクションを実行する
	// humanPlayMs: 迷い時間(ms, 0=計測なし)
	PlayerAction(action, amount, humanPlayMs int) error
	// GetPhase 現在のフェーズを取得する
	GetPhase() int
	// GetPlayers プレイヤー一覧を取得する
	GetPlayers() []*domain.HoldemPlayer
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.HoldemPlayer
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetCommunityCards コミュニティカードを取得する
	GetCommunityCards() []*domain.Card
	// GetPot ポット額を取得する
	GetPot() int
	// GetSidePots サイドポット一覧を取得する
	GetSidePots() []domain.HoldemSidePot
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
	// GetRoundResults ラウンド結果を取得する
	GetRoundResults() []domain.HoldemResult
	// GetCpuActions CPU行動記録を取得する
	GetCpuActions() []domain.HoldemCpuAction
	// GetConfig ゲーム設定を取得する
	GetConfig() domain.HoldemConfig
	// SetConfig ゲーム設定を変更する
	SetConfig(cfg domain.HoldemConfig)
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// GetActedFlags 各プレイヤーのアクション済みフラグを取得する
	GetActedFlags() []bool
	// GetHandCount ハンド数を取得する
	GetHandCount() int
	// Resize プレイヤー数を変更する
	Resize(players []*domain.HoldemPlayer)
	// Rebuy リバイを実行する
	Rebuy() error
	// SkipRebuy リバイをスキップする
	SkipRebuy() error
	// Addon アドオンを実行する
	Addon() error
	// SkipAddon アドオンをスキップする
	SkipAddon() error
	// IsRebuyAvailable リバイが可能かを返す
	IsRebuyAvailable() bool
	// IsAddonAvailable アドオンが可能かを返す
	IsAddonAvailable() bool
	// GetRebuyCounts 各プレイヤーのリバイ回数を取得する
	GetRebuyCounts() []int
	// GetAddonUsed 各プレイヤーのアドオン使用状態を取得する
	GetAddonUsed() []bool
	// GetRebuyPhaseType リバイフェーズ種別を取得する
	GetRebuyPhaseType() int
	// Muck ハンドをマックする
	Muck() error
	// ShowHand ハンドを公開する
	ShowHand() error
	// IsMuckAvailable マックが可能かを返す
	IsMuckAvailable() bool
	// GetHumanProfile メタAIプロファイルを取得する
	GetHumanProfile() *domain.BettingHumanProfile
	// ResetProfile メタAIプロファイルをリセットする
	ResetProfile()
	// ExportProfile メタAIプロファイルをエクスポートする
	ExportProfile() interface{}
	// ImportProfile JSONバイトからメタAIプロファイルをインポートする
	ImportProfile(data []byte) error
	// GetEquity エクイティ計算結果を取得する
	GetEquity() *domain.HoldemEquityResult
	// GetPotOdds ポットオッズを取得する
	GetPotOdds() float64
}
