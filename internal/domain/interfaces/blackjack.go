package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// BlackJackGame ブラックジャックゲームインタフェース
type BlackJackGame interface {
	// interactor が呼ぶメソッド
	Reset()
	PlayerBet(amount int) error
	PlayerInsurance() error
	PlayerDeclineInsurance() error
	PlayerHit() error
	PlayerStand() error
	PlayerDoubleDown() error
	PlayerSplit() error

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
}
