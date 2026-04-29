package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// TexasHoldemBonusInteractorIF テキサスホールデムボーナスポーカーインタラクターインタフェース
type TexasHoldemBonusInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// Bet アンテベットとオプションのボーナスサイドベット
	Bet(ante, bonus int) string
	// Play プリフロップでフロップベット
	Play() string
	// Fold プリフロップでフォールド
	Fold() string
	// Check フロップ後またはターン後のチェック
	Check() string
	// Raise フロップ後またはターン後の1×アンテベット
	Raise() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// TexasHoldemBonusInteractor テキサスホールデムボーナスポーカーインタラクタークラス
type TexasHoldemBonusInteractor struct {
	GameBase[interfaces.TexasHoldemBonusGame]
	tp presenter.TexasHoldemBonusPresenter
}

// NewTexasHoldemBonusInteractor コンストラクタ
func NewTexasHoldemBonusInteractor(g interfaces.TexasHoldemBonusGame, tp presenter.TexasHoldemBonusPresenter) *TexasHoldemBonusInteractor {
	mustNotNil("TexasHoldemBonusInteractor", map[string]any{"g": g, "tp": tp})
	return &TexasHoldemBonusInteractor{
		GameBase: GameBase[interfaces.TexasHoldemBonusGame]{Game: g},
		tp:       tp,
	}
}

// Reset ゲーム初期化
func (ti *TexasHoldemBonusInteractor) Reset() string {
	return runAndPresent(ti.Game, ti.tp, ti.Game.Reset)
}

// Bet アンテベット
func (ti *TexasHoldemBonusInteractor) Bet(ante, bonus int) string {
	return execAndPresent(ti.Game, ti.tp, func() error { return ti.Game.Bet(ante, bonus) })
}

// Play プリフロップでフロップベット
func (ti *TexasHoldemBonusInteractor) Play() string {
	return execAndPresent(ti.Game, ti.tp, ti.Game.Play)
}

// Fold プリフロップでフォールド
func (ti *TexasHoldemBonusInteractor) Fold() string {
	return execAndPresent(ti.Game, ti.tp, ti.Game.Fold)
}

// Check フロップ後またはターン後のチェック
func (ti *TexasHoldemBonusInteractor) Check() string {
	return execAndPresent(ti.Game, ti.tp, ti.Game.Check)
}

// Raise フロップ後またはターン後の1×アンテベット
func (ti *TexasHoldemBonusInteractor) Raise() string {
	return execAndPresent(ti.Game, ti.tp, ti.Game.Raise)
}

// ActionLog 棋譜を出力する
func (ti *TexasHoldemBonusInteractor) ActionLog() string {
	return ti.tp.ActionLogOutput(ti.Game)
}

// RestoreTexasHoldemBonusInteractor deserialises JSON into a TexasHoldemBonusInteractor.
func RestoreTexasHoldemBonusInteractor(data []byte, tp presenter.TexasHoldemBonusPresenter) (*TexasHoldemBonusInteractor, error) {
	return restoreAndBuild[domain.TexasHoldemBonus](data, func(g *domain.TexasHoldemBonus) *TexasHoldemBonusInteractor {
		return &TexasHoldemBonusInteractor{GameBase: GameBase[interfaces.TexasHoldemBonusGame]{Game: g}, tp: tp}
	})
}
