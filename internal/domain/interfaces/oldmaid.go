package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// OldMaidGame ババ抜きゲームインタフェース
type OldMaidGame interface {
	// interactor が呼ぶメソッド
	Reset()
	GetGameEndFlag() bool
	IsHumanTurn() bool
	PlayerDraw(cardIdx int) error
	CpuDraw() error

	// presenter が呼ぶメソッド
	GetPlayerCnt() int
	GetPlayer(i int) *domain.OldMaidPlayer
	GetHasDrawn() bool
	GetLastDrawPlayerIdx() int
	GetLastDrawFromIdx() int
	GetLastDrawCard() *domain.Card
	GetLastDiscardedPairs() int
	GetLastDiscardedCards() []*domain.Card
	GetCpuActions() []*domain.OldMaidCpuAction
	GetHumanAction() *domain.OldMaidCpuAction
	GetLoserIdx() int
	GetCurrentTurn() int
	GetNextDrawTargetIdx() int
}
