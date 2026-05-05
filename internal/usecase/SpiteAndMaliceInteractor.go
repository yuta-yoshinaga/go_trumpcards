package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// SpiteAndMaliceInteractorIF Spite & Malice インタラクターインタフェース
type SpiteAndMaliceInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// PlayFromHand 手札からファウンデーションに出す
	PlayFromHand(handIdx, foundationIdx int) string
	// PlayFromGoal ゴールパイルのトップをファウンデーションに出す
	PlayFromGoal(foundationIdx int) string
	// PlayFromSide サイドパイルのトップをファウンデーションに出す
	PlayFromSide(sideIdx, foundationIdx int) string
	// Discard 手札 1 枚をサイドパイルに捨ててターンを終了する
	Discard(handIdx, sideIdx int) string
	// CpuStep CPU の手番を 1 ステップ進める
	CpuStep() string
	// AutoComplete 自明な手をまとめて適用する
	AutoComplete() string
	// Hint 推奨手を出力する
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// SpiteAndMaliceInteractor Spite & Malice インタラクタークラス
type SpiteAndMaliceInteractor struct {
	GameBase[interfaces.SpiteAndMaliceGame]
	sp presenter.SpiteAndMalicePresenter
}

// NewSpiteAndMaliceInteractor コンストラクタ
func NewSpiteAndMaliceInteractor(g interfaces.SpiteAndMaliceGame, sp presenter.SpiteAndMalicePresenter) *SpiteAndMaliceInteractor {
	mustNotNil("SpiteAndMaliceInteractor", map[string]any{"g": g, "sp": sp})
	return &SpiteAndMaliceInteractor{GameBase: GameBase[interfaces.SpiteAndMaliceGame]{Game: g}, sp: sp}
}

// Reset ゲーム初期化
func (si *SpiteAndMaliceInteractor) Reset() string {
	return runAndPresent(si.Game, si.sp, si.Game.Reset)
}

// PlayFromHand 手札からファウンデーションに出す
func (si *SpiteAndMaliceInteractor) PlayFromHand(handIdx, foundationIdx int) string {
	return execAndPresent(si.Game, si.sp, func() error { return si.Game.PlayFromHand(handIdx, foundationIdx) })
}

// PlayFromGoal ゴールパイルのトップをファウンデーションに出す
func (si *SpiteAndMaliceInteractor) PlayFromGoal(foundationIdx int) string {
	return execAndPresent(si.Game, si.sp, func() error { return si.Game.PlayFromGoal(foundationIdx) })
}

// PlayFromSide サイドパイルのトップをファウンデーションに出す
func (si *SpiteAndMaliceInteractor) PlayFromSide(sideIdx, foundationIdx int) string {
	return execAndPresent(si.Game, si.sp, func() error { return si.Game.PlayFromSide(sideIdx, foundationIdx) })
}

// Discard 手札 1 枚をサイドパイルに捨ててターンを終了する
func (si *SpiteAndMaliceInteractor) Discard(handIdx, sideIdx int) string {
	return execAndPresent(si.Game, si.sp, func() error { return si.Game.Discard(handIdx, sideIdx) })
}

// CpuStep CPU の手番を 1 ステップ進める
func (si *SpiteAndMaliceInteractor) CpuStep() string {
	return execAndPresent(si.Game, si.sp, si.Game.CpuStep)
}

// AutoComplete 自明な手をまとめて適用する
func (si *SpiteAndMaliceInteractor) AutoComplete() string {
	return execAndPresent(si.Game, si.sp, si.Game.AutoComplete)
}

// Hint 推奨手を出力する
func (si *SpiteAndMaliceInteractor) Hint() string {
	return si.sp.HintOutput(si.Game)
}

// ActionLog 棋譜を出力する
func (si *SpiteAndMaliceInteractor) ActionLog() string {
	return si.sp.ActionLogOutput(si.Game)
}

// RestoreSpiteAndMaliceInteractor deserialises JSON into a SpiteAndMaliceInteractor.
func RestoreSpiteAndMaliceInteractor(data []byte, sp presenter.SpiteAndMalicePresenter) (*SpiteAndMaliceInteractor, error) {
	return restoreAndBuild[domain.SpiteAndMalice](data, func(g *domain.SpiteAndMalice) *SpiteAndMaliceInteractor {
		return &SpiteAndMaliceInteractor{GameBase: GameBase[interfaces.SpiteAndMaliceGame]{Game: g}, sp: sp}
	})
}
