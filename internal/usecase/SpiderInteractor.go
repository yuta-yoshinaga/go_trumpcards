package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// SpiderInteractorIF スパイダーソリティアインタラクターインタフェース
type SpiderInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
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
	// UndoN n回連続アンドゥ
	UndoN(n int) string
}

// SpiderInteractor スパイダーソリティアインタラクタークラス
type SpiderInteractor struct {
	GameBase[interfaces.SpiderGame]
	sp presenter.SpiderPresenter
}

// NewSpiderInteractor コンストラクタ
func NewSpiderInteractor(s interfaces.SpiderGame, sp presenter.SpiderPresenter) *SpiderInteractor {
	mustNotNil("SpiderInteractor", map[string]any{"s": s, "sp": sp})
	return &SpiderInteractor{GameBase: GameBase[interfaces.SpiderGame]{Game: s}, sp: sp}
}

// Reset ゲーム初期化
func (si *SpiderInteractor) Reset() string {
	return runAndPresent(si.Game, si.sp, si.Game.Reset)
}

// ResetWithConfig 設定付きリセット
func (si *SpiderInteractor) ResetWithConfig(cfg domain.SpiderConfig) string {
	return runAndPresent(si.Game, si.sp, func() { si.Game.ResetWithConfig(cfg) })
}

// Deal ストックからタブローに配る
func (si *SpiderInteractor) Deal() string {
	return execAndPresent(si.Game, si.sp, si.Game.Deal)
}

// MoveTableauToTableau タブロー間でカードを移動
func (si *SpiderInteractor) MoveTableauToTableau(fromCol, cardIndex, toCol int) string {
	return execAndPresent(si.Game, si.sp, func() error { return si.Game.MoveTableauToTableau(fromCol, cardIndex, toCol) })
}

// GiveUp ギブアップ
func (si *SpiderInteractor) GiveUp() string {
	return runAndPresent(si.Game, si.sp, si.Game.GiveUp)
}

// Hint ヒント取得
func (si *SpiderInteractor) Hint() string {
	return si.sp.HintOutput(si.Game)
}

// AutoComplete オートコンプリート
func (si *SpiderInteractor) AutoComplete() string {
	return execAndPresent(si.Game, si.sp, si.Game.AutoComplete)
}

// ActionLog 棋譜を出力する
func (si *SpiderInteractor) ActionLog() string {
	return si.sp.ActionLogOutput(si.Game)
}

// Undo アンドゥ
func (si *SpiderInteractor) Undo() string {
	return execAndPresent(si.Game, si.sp, si.Game.Undo)
}

// UndoN n回連続アンドゥ
func (si *SpiderInteractor) UndoN(n int) string {
	return execAndPresent(si.Game, si.sp, func() error { return si.Game.UndoN(n) })
}

// RestoreSpiderInteractor deserialises JSON into a SpiderInteractor.
func RestoreSpiderInteractor(data []byte, sp presenter.SpiderPresenter) (*SpiderInteractor, error) {
	return restoreAndBuild[domain.Spider](data, func(g *domain.Spider) *SpiderInteractor {
		return &SpiderInteractor{GameBase: GameBase[interfaces.SpiderGame]{Game: g}, sp: sp}
	})
}
