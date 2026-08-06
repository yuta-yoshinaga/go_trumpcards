//go:build !js || !wasm || casino

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// BlackJackGame ブラックジャックゲームインタフェース
type BlackJackGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// PlayerBet プレイヤーがベットしてゲームを開始する
	PlayerBet(amount, ppBet, t3Bet, handCount int) error
	// PlayerInsurance プレイヤーがインシュランスを購入する
	PlayerInsurance() error
	// PlayerDeclineInsurance プレイヤーがインシュランスを辞退する
	PlayerDeclineInsurance() error
	// PlayerHit プレイヤーがヒットする
	PlayerHit() error
	// PlayerStand プレイヤーがスタンドする
	PlayerStand() error
	// PlayerDoubleDown プレイヤーがダブルダウンする
	PlayerDoubleDown() error
	// PlayerSplit プレイヤーがスプリットする
	PlayerSplit() error
	// PlayerSurrender プレイヤーがサレンダーする
	PlayerSurrender() error
	// PlayerEarlySurrender プレイヤーがアーリーサレンダーする
	PlayerEarlySurrender() error
	// PlayerDeclineEarlySurrender プレイヤーがアーリーサレンダーを辞退する
	PlayerDeclineEarlySurrender() error
	// SetDeckCount デッキ数を設定する
	SetDeckCount(count int) error
	// ToggleHint ヒント表示を切り替える
	ToggleHint()
	// SetConfig ゲーム設定を変更する
	SetConfig(config domain.BlackJackConfig) error

	// GetPlayer プレイヤーを取得する
	GetPlayer() *domain.BlackJackPlayer
	// GetDealer ディーラーを取得する
	GetDealer() *domain.BlackJackPlayer
	// GetPhase 現在のフェーズを取得する
	GetPhase() int
	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPlayerHands プレイヤーハンド一覧を取得する
	GetPlayerHands() []*domain.BlackJackHand
	// GetCurrentHandIdx 現在操作中のハンドインデックスを取得する
	GetCurrentHandIdx() int
	// GetInsuranceBet インシュランスベット額を取得する
	GetInsuranceBet() int
	// IsInsuranceAvailable インシュランスが可能かを返す
	IsInsuranceAvailable() bool
	// GameJudgmentForHand 指定ハンドの勝敗を判定する
	GameJudgmentForHand(handIdx int) domain.GameResult
	// GameJudgment ゲーム勝敗を判定する
	GameJudgment() domain.GameResult
	// GetDeckCount デッキ数を取得する
	GetDeckCount() int
	// IsHintEnabled ヒントが有効かを返す
	IsHintEnabled() bool
	// GetBasicStrategySuggestion ベーシックストラテジー推奨アクションを取得する
	GetBasicStrategySuggestion() domain.BJSuggestedAction
	// GetConfig ゲーム設定を取得する
	GetConfig() domain.BlackJackConfig
	// GetRunningCount ランニングカウントを取得する
	GetRunningCount() int
	// GetTrueCount トゥルーカウントを取得する
	GetTrueCount() float64
	// IsCountingEnabled カウンティングが有効かを返す
	IsCountingEnabled() bool
	// GetCpuPlayers CPUプレイヤー一覧を取得する
	GetCpuPlayers() []*domain.BlackJackCpuSeat
	// GetSideBetResults サイドベット結果を取得する
	GetSideBetResults() []*domain.BJSideBetResult
	// GetPerfectPairsBet Perfect Pairsベット額を取得する
	GetPerfectPairsBet() int
	// Get21Plus3Bet 21+3ベット額を取得する
	Get21Plus3Bet() int
	// GetDeckPenetration デッキペネトレーション率を取得する
	GetDeckPenetration() int
	// GetMultiHandCount マルチハンド数を取得する
	GetMultiHandCount() int
	// GetVariant バリアント設定を取得する (nil = 標準ブラックジャック)
	GetVariant() *domain.BlackJackVariantConfig
	// GetBonusKeys 当ラウンドで成立したバリアントボーナスのi18nキー一覧を取得する
	GetBonusKeys() []string
	// CanSurrenderHand 指定ハンドのサレンダー可否を返す
	CanSurrenderHand(handIdx int) bool
	// CanSurrenderCpuHand CPUハンドのサレンダー可否を返す
	CanSurrenderCpuHand(cpuIdx, handIdx int) bool
}
