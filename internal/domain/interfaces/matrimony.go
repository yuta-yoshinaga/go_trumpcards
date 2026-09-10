//go:build !js || !wasm || extra

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// MatrimonyGame is the domain contract used by Matrimony adapters and use cases.
type MatrimonyGame interface {
	SolitaireGame
	GetGameEndFlag() bool
	Reset()
	Draw() error
	MoveTableauToFoundation(int) error
	MoveWasteToFoundation() error
	MoveWasteToTableau(int) error
	MoveStockToTableau(int) error
	GetHint() *domain.MatrimonyHint
	GetPhase() domain.MatrimonyPhase
	GetMoveCount() int
	GetStockCount() int
	GetRedealCount() int
	GetWaste() []*domain.Card
	GetTableau() [domain.MatrimonyTableauCnt]*domain.Card
	GetFoundation() [domain.MatrimonyFoundationCnt][]*domain.Card
	AllFaceUp() bool
	IsStalemate() bool
	UndoToEscape() int
}
