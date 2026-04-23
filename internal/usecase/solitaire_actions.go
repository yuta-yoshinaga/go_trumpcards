package usecase

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// solitaireActions はソリティア系インタラクターの GiveUp/AutoComplete/Undo/UndoN 委譲を共通化する。
// 各ゲームのインタラクター構造体に埋め込んで使用する (tournamentActions と同パターン)。
type solitaireActions[G interfaces.SolitaireGame] struct {
	game G
	pres outputPresenter[G]
}

// newSolitaireActions コンストラクタ
func newSolitaireActions[G interfaces.SolitaireGame](game G, pres outputPresenter[G]) solitaireActions[G] {
	return solitaireActions[G]{game: game, pres: pres}
}

// GiveUp ギブアップ
func (s solitaireActions[G]) GiveUp() string {
	return runAndPresent(s.game, s.pres, s.game.GiveUp)
}

// AutoComplete オートコンプリート
func (s solitaireActions[G]) AutoComplete() string {
	return execAndPresent(s.game, s.pres, s.game.AutoComplete)
}

// Undo アンドゥ
func (s solitaireActions[G]) Undo() string {
	return execAndPresent(s.game, s.pres, s.game.Undo)
}

// UndoN n回連続アンドゥ
func (s solitaireActions[G]) UndoN(n int) string {
	return execAndPresent(s.game, s.pres, func() error { return s.game.UndoN(n) })
}
