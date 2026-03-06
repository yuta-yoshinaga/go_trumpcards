package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// PokerGame ポーカーゲームインタフェース (マルチプレイヤー)
type PokerGame interface {
	// interactor が呼ぶメソッド
	Reset() error
	PlayerAction(action, amount int) error
	PlayerExchange(indices []int) error
	PlayerStand() error
	CalcDrawOdds(indices []int) ([]domain.PokerDrawOdds, error)

	// presenter が呼ぶメソッド
	GetPlayers() []*domain.PokerPlayer
	GetPhase() int
	GetPot() int
	GetSidePots() []domain.PokerSidePot
	GetDealerIdx() int
	GetCurrentTurn() int
	GetGameEndFlag() bool
	GetLastBet() int
	GetMinRaise() int
	GetRaiseCount() int
	GetAnte() int
	GetRoundResults() []domain.PokerResult
	GetCpuActions() []domain.PokerCpuAction
	GetCpuExchanges() []domain.PokerCpuExchange
	GetConfig() domain.PokerConfig
	SetConfig(cfg domain.PokerConfig)
}
