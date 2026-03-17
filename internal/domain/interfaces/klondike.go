package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// KlondikeGame クロンダイクゲームインタフェース
type KlondikeGame interface {
	// interactor が呼ぶメソッド
	Reset()
	ResetWithConfig(cfg domain.KlondikeConfig)
	Draw() error
	MoveWasteToTableau(col int) error
	MoveWasteToFoundation() error
	MoveTableauToTableau(fromCol, cardIndex, toCol int) error
	MoveTableauToFoundation(col int) error
	GiveUp()
	GetHint() *domain.KlondikeHint
	AutoComplete() error
	Undo() error

	// state readers
	CanUndo() bool
	GetPhase() domain.KlondikePhase
	GetMoveCount() int
	GetStockCount() int
	GetWaste() []*domain.Card
	GetTableau() [domain.KlondikeTableauCnt][]*domain.KlondikeTableauCard
	GetFoundation() [domain.KlondikeFoundationCnt][]*domain.Card
	AllFaceUp() bool
	GetActionLog() []*domain.ActionLogEntry
	GetDrawCount() int
	GetScore() int
	GetScoringMode() domain.KlondikeScoringMode
}
