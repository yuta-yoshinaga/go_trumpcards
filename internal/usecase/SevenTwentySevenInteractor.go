//go:build !js || !wasm || extra4

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// SevenTwentySevenInteractorIF はセブン・トゥエンティセブン (SevenTwentySeven) のインタラクターインタフェース。
type SevenTwentySevenInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化 (新規ゲーム)
	Reset() string
	// ResetWithConfig 設定を変更してゲームを初期化
	ResetWithConfig(cfg domain.SevenTwentySevenConfig) string
	// TakeCard 人間の「引く / 止まる」を受け付ける。全員が止まるまで繰り返す
	TakeCard(draw bool) string
	// NextRound 次のラウンドを配る
	NextRound() string
	// GetConfig 現在の設定を返す
	GetConfig() domain.SevenTwentySevenConfig
	// Hint ヒントを出力する
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// SevenTwentySevenInteractor はセブン・トゥエンティセブンのインタラクター。
type SevenTwentySevenInteractor struct {
	GameBase[interfaces.SevenTwentySevenGame]
	sp presenter.SevenTwentySevenPresenter
}

// NewSevenTwentySevenInteractor コンストラクタ。
func NewSevenTwentySevenInteractor(g interfaces.SevenTwentySevenGame, sp presenter.SevenTwentySevenPresenter) *SevenTwentySevenInteractor {
	mustNotNil("SevenTwentySevenInteractor", map[string]any{"g": g, "sp": sp})
	return &SevenTwentySevenInteractor{GameBase: GameBase[interfaces.SevenTwentySevenGame]{Game: g}, sp: sp}
}

// Reset ゲーム初期化 (新規ゲーム)。
func (ti *SevenTwentySevenInteractor) Reset() string {
	return runAndPresent(ti.Game, ti.sp, ti.Game.Reset)
}

// ResetWithConfig 設定を変更してゲームを初期化。
func (ti *SevenTwentySevenInteractor) ResetWithConfig(cfg domain.SevenTwentySevenConfig) string {
	return resetWithValidatedConfig(ti.Game, ti.sp, cfg, ti.Game.SetConfig, ti.Reset)
}

// TakeCard 人間の「引く / 止まる」を受け付ける。
func (ti *SevenTwentySevenInteractor) TakeCard(draw bool) string {
	if out, blocked := guardGameEnd(ti.Game, ti.sp); blocked {
		return out
	}
	return execAndPresent(ti.Game, ti.sp, func() error {
		return ti.Game.TakeCard(draw)
	})
}

// NextRound 次のラウンドを配る。
func (ti *SevenTwentySevenInteractor) NextRound() string {
	return runAndPresent(ti.Game, ti.sp, ti.Game.NextRound)
}

// GetConfig 現在の設定を返す。
func (ti *SevenTwentySevenInteractor) GetConfig() domain.SevenTwentySevenConfig {
	return ti.Game.GetConfig()
}

// Hint ヒントを出力する。
func (ti *SevenTwentySevenInteractor) Hint() string { return ti.sp.HintOutput(ti.Game) }

// ActionLog 棋譜を出力する。
func (ti *SevenTwentySevenInteractor) ActionLog() string { return ti.sp.ActionLogOutput(ti.Game) }

// RestoreSevenTwentySevenInteractor deserialises JSON into a SevenTwentySevenInteractor.
func RestoreSevenTwentySevenInteractor(data []byte, sp presenter.SevenTwentySevenPresenter) (*SevenTwentySevenInteractor, error) {
	return restoreAndBuild[domain.SevenTwentySeven](data, func(g *domain.SevenTwentySeven) *SevenTwentySevenInteractor {
		return &SevenTwentySevenInteractor{GameBase: GameBase[interfaces.SevenTwentySevenGame]{Game: g}, sp: sp}
	})
}
