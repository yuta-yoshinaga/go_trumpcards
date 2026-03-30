package usecase

import (
	"encoding/json"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// ThreeCardInteractorIF スリーカードポーカーインタラクターインタフェース
type ThreeCardInteractorIF interface {
	// Reset ゲーム初期化
	Reset() string
	// Bet アンテベット
	Bet(ante, pairPlus int) string
	// Play プレイ
	Play() string
	// Fold フォールド
	Fold() string
	// ActionLog 棋��を出力する
	ActionLog() string
}

// ThreeCardInteractor スリーカードポーカーインタラクタークラス
type ThreeCardInteractor struct {
	tc interfaces.ThreeCardGame
	tp presenter.ThreeCardPresenter
}

// NewThreeCardInteractor コンストラクタ
func NewThreeCardInteractor(tc interfaces.ThreeCardGame, tp presenter.ThreeCardPresenter) *ThreeCardInteractor {
	mustNotNil("ThreeCardInteractor", map[string]any{"tc": tc, "tp": tp})
	return &ThreeCardInteractor{
		tc: tc,
		tp: tp,
	}
}

// Reset ゲーム初期化
func (ti *ThreeCardInteractor) Reset() string {
	return runAndPresent(ti.tc, ti.tp, ti.tc.Reset)
}

// Bet アンテベット
func (ti *ThreeCardInteractor) Bet(ante, pairPlus int) string {
	return execAndPresent(ti.tc, ti.tp, func() error { return ti.tc.Bet(ante, pairPlus) })
}

// Play プレイ
func (ti *ThreeCardInteractor) Play() string {
	return execAndPresent(ti.tc, ti.tp, ti.tc.Play)
}

// Fold フォールド
func (ti *ThreeCardInteractor) Fold() string {
	return execAndPresent(ti.tc, ti.tp, ti.tc.Fold)
}

// ActionLog 棋譜を出力する
func (ti *ThreeCardInteractor) ActionLog() string {
	return ti.tp.ActionLogOutput(ti.tc)
}

// Snapshot serialises the game state to JSON for KV persistence.
func (ti *ThreeCardInteractor) Snapshot() ([]byte, error) {
	return json.Marshal(ti.tc)
}

// RestoreThreeCardInteractor deserialises JSON into a ThreeCardInteractor.
func RestoreThreeCardInteractor(data []byte, tp presenter.ThreeCardPresenter) (*ThreeCardInteractor, error) {
	tc, err := restoreGame[domain.ThreeCard](data)
	if err != nil {
		return nil, err
	}
	return &ThreeCardInteractor{tc: tc, tp: tp}, nil
}
