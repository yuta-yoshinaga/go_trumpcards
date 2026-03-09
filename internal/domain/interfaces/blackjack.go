package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// BlackJackGame ブラックジャックゲームインタフェース
type BlackJackGame interface {
	// interactor が呼ぶメソッド
	Reset()
	PlayerBet(amount, ppBet, t3Bet, handCount int) error
	PlayerInsurance() error
	PlayerDeclineInsurance() error
	PlayerHit() error
	PlayerStand() error
	PlayerDoubleDown() error
	PlayerSplit() error
	PlayerSurrender() error
	PlayerEarlySurrender() error
	PlayerDeclineEarlySurrender() error
	SetDeckCount(count int) error
	ToggleHint()
	SetConfig(config domain.BlackJackConfig) error

	// presenter が呼ぶメソッド
	GetPlayer() *domain.BlackJackPlayer
	GetDealer() *domain.BlackJackPlayer
	GetPhase() int
	GetGameEndFlag() bool
	GetPlayerHands() []*domain.BlackJackHand
	GetCurrentHandIdx() int
	GetInsuranceBet() int
	IsInsuranceAvailable() bool
	GameJudgmentForHand(handIdx int) domain.GameResult
	GameJudgment() domain.GameResult
	GetDeckCount() int
	IsHintEnabled() bool
	GetBasicStrategySuggestion() domain.BJSuggestedAction
	GetConfig() domain.BlackJackConfig
	GetRunningCount() int
	GetTrueCount() float64
	IsCountingEnabled() bool
	GetCpuPlayers() []*domain.BlackJackCpuSeat
	GetSideBetResults() []*domain.BJSideBetResult
	GetPerfectPairsBet() int
	Get21Plus3Bet() int
	GetDeckPenetration() int
	GetMultiHandCount() int
	CanSurrenderHand(handIdx int) bool
	CanSurrenderCpuHand(cpuIdx, handIdx int) bool
	GetActionLog() []*domain.ActionLogEntry
}
