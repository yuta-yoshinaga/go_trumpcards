package usecase

import (
	"encoding/json"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// ClockSolitaireInteractorIF クロックソリティアインタラクターインタフェース
type ClockSolitaireInteractorIF interface {
	// Reset ゲーム初期化
	Reset() string
	// Step 1ステップ実行
	Step() string
	// AutoPlay 自動プレイ
	AutoPlay() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// ClockSolitaireInteractor クロックソリティアインタラクタークラス
type ClockSolitaireInteractor struct {
	g interfaces.ClockSolitaireGame
	p presenter.ClockSolitairePresenter
}

// NewClockSolitaireInteractor コンストラクタ
func NewClockSolitaireInteractor(g interfaces.ClockSolitaireGame, p presenter.ClockSolitairePresenter) *ClockSolitaireInteractor {
	mustNotNil("ClockSolitaireInteractor", map[string]any{"g": g, "p": p})
	return &ClockSolitaireInteractor{g: g, p: p}
}

// Reset ゲーム初期化
func (ci *ClockSolitaireInteractor) Reset() string {
	return runAndPresent(ci.g, ci.p, ci.g.Reset)
}

// Step 1ステップ実行
func (ci *ClockSolitaireInteractor) Step() string {
	return execAndPresent(ci.g, ci.p, ci.g.Step)
}

// AutoPlay 自動プレイ
func (ci *ClockSolitaireInteractor) AutoPlay() string {
	return execAndPresent(ci.g, ci.p, ci.g.AutoPlay)
}

// ActionLog 棋譜を出力する
func (ci *ClockSolitaireInteractor) ActionLog() string {
	return ci.p.ActionLogOutput(ci.g)
}

// Snapshot serialises the game state to JSON for KV persistence.
func (ci *ClockSolitaireInteractor) Snapshot() ([]byte, error) {
	return json.Marshal(ci.g)
}

// RestoreClockSolitaireInteractor deserialises JSON into a ClockSolitaireInteractor.
func RestoreClockSolitaireInteractor(data []byte, p presenter.ClockSolitairePresenter) (*ClockSolitaireInteractor, error) {
	cs, err := restoreGame[domain.ClockSolitaire](data)
	if err != nil {
		return nil, err
	}
	return &ClockSolitaireInteractor{g: cs, p: p}, nil
}
