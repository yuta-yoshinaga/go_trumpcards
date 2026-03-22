package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// SpiderInteractorIF スパイダーソリティアインタラクターインタフェース
type SpiderInteractorIF interface {
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定付きリセット
	ResetWithConfig(cfg domain.SpiderConfig) string
	// Deal ストックからタブローに配る
	Deal() string
	// MoveTableauToTableau タブロー間でカードを移動
	MoveTableauToTableau(fromCol, cardIndex, toCol int) string
	// GiveUp ギブアップ
	GiveUp() string
	// Hint ヒント取得
	Hint() string
	// AutoComplete オートコンプリート
	AutoComplete() string
	// ActionLog 棋譜を出力する
	ActionLog() string
	// Undo アンドゥ
	Undo() string
}

// SpiderInteractor スパイダーソリティアインタラクタークラス
type SpiderInteractor struct {
	s  interfaces.SpiderGame
	sp presenter.SpiderPresenter
}

// NewSpiderInteractor コンストラクタ
func NewSpiderInteractor(s interfaces.SpiderGame, sp presenter.SpiderPresenter) *SpiderInteractor {
	mustNotNil("SpiderInteractor", map[string]any{"s": s, "sp": sp})
	return &SpiderInteractor{s: s, sp: sp}
}

// Reset ゲーム初期化
func (si *SpiderInteractor) Reset() string {
	return runAndPresent(si.s, si.sp, si.s.Reset)
}

// ResetWithConfig 設定付きリセット
func (si *SpiderInteractor) ResetWithConfig(cfg domain.SpiderConfig) string {
	return runAndPresent(si.s, si.sp, func() { si.s.ResetWithConfig(cfg) })
}

// Deal ストックからタブローに配る
func (si *SpiderInteractor) Deal() string {
	return execAndPresent(si.s, si.sp, si.s.Deal)
}

// MoveTableauToTableau タブロー間でカードを移動
func (si *SpiderInteractor) MoveTableauToTableau(fromCol, cardIndex, toCol int) string {
	return execAndPresent(si.s, si.sp, func() error { return si.s.MoveTableauToTableau(fromCol, cardIndex, toCol) })
}

// GiveUp ギブアップ
func (si *SpiderInteractor) GiveUp() string {
	return runAndPresent(si.s, si.sp, si.s.GiveUp)
}

// Hint ヒント取得
func (si *SpiderInteractor) Hint() string {
	return si.sp.HintOutput(si.s)
}

// AutoComplete オートコンプリート
func (si *SpiderInteractor) AutoComplete() string {
	return execAndPresent(si.s, si.sp, si.s.AutoComplete)
}

// ActionLog 棋譜を出力する
func (si *SpiderInteractor) ActionLog() string {
	return si.sp.ActionLogOutput(si.s)
}

// Undo アンドゥ
func (si *SpiderInteractor) Undo() string {
	return execAndPresent(si.s, si.sp, si.s.Undo)
}
