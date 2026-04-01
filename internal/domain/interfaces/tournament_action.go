package interfaces

// TournamentActionGame トーナメントアクション共通インタフェース
// リバイ・アドオン・マック・ハンド公開など、ホールデム系ゲーム共通のトーナメント操作を定義する。
type TournamentActionGame interface {
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
}
