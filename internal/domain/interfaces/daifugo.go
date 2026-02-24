package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// DaifugoGame 大富豪ゲームインタフェース
type DaifugoGame interface {
	// interactor が呼ぶメソッド
	Reset()
	GetGameEndFlag() bool
	IsHumanTurn() bool
	PlayerPlay(indices []int) error
	CpuPlay()

	// presenter が呼ぶメソッド
	GetPlayerCnt() int
	GetPlayer(i int) *domain.DaifugoPlayer
	GetRevolutionActive() bool
	GetElevenBackActive() bool
	GetSuitLocked() bool
	GetLockedSuit() int
	GetTableIsSequence() bool
	GetExchangeActions() []*domain.DaifugoExchangeAction
	GetTableCards() []*domain.Card
	GetLastPlayPlayerIdx() int
	GetHumanAction() *domain.DaifugoCpuAction
	GetCpuActions() []*domain.DaifugoCpuAction
	GetCurrentTurn() int
	GetConfig() domain.DaifugoConfig
	GetPassCount() int
}
