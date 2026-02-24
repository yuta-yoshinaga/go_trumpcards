package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// PokerGame ポーカーゲームインタフェース
type PokerGame interface {
	// interactor が呼ぶメソッド
	Reset()
	PlayerExchange(indices []int) error
	PlayerStand() error
	PlayerBet(amount int) error
	PlayerCall() error
	PlayerRaise(amount int) error
	PlayerFold() error
	PlayerCheck() error

	// presenter が呼ぶメソッド
	GetPlayer() *domain.PokerPlayer
	GetDealer() *domain.PokerPlayer
	GetPot() int
	GetPlayerBet() int
	GetDealerBet() int
	GetPhase() int
	GetFolded() int
	GetAnte() int
	GameJudgment() int
}
