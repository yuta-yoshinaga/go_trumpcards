//go:build !js || !wasm || extra2

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// AuldLangSyneInteractorIF オールド・ラング・サインインタラクターインタフェース
type AuldLangSyneInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// Deal ストックから各ウェイストへ1枚ずつ配る
	Deal() string
	// PlayWasteToFoundation ウェイスト最上段をファンデーションに置く
	PlayWasteToFoundation(wasteIdx, fIdx int) string
	// GiveUp ギブアップ
	GiveUp() string
	// Undo アンドゥ
	Undo() string
	// UndoN n回連続アンドゥ
	UndoN(n int) string
	// AutoComplete オートコンプリート
	AutoComplete() string
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// AuldLangSyneInteractor オールド・ラング・サインインタラクタークラス
type AuldLangSyneInteractor struct {
	GameBase[interfaces.AuldLangSyneGame]
	solitaireActions[interfaces.AuldLangSyneGame]
	p presenter.AuldLangSynePresenter
}

// NewAuldLangSyneInteractor コンストラクタ
func NewAuldLangSyneInteractor(g interfaces.AuldLangSyneGame, p presenter.AuldLangSynePresenter) *AuldLangSyneInteractor {
	mustNotNil("AuldLangSyneInteractor", map[string]any{"g": g, "p": p})
	return &AuldLangSyneInteractor{
		GameBase:         GameBase[interfaces.AuldLangSyneGame]{Game: g},
		solitaireActions: newSolitaireActions[interfaces.AuldLangSyneGame](g, p),
		p:                p,
	}
}

// Reset ゲーム初期化
func (ai *AuldLangSyneInteractor) Reset() string {
	return runAndPresent(ai.Game, ai.p, ai.Game.Reset)
}

// Deal ストックから各ウェイストへ1枚ずつ配る
func (ai *AuldLangSyneInteractor) Deal() string {
	return execAndPresent(ai.Game, ai.p, ai.Game.Deal)
}

// PlayWasteToFoundation ウェイスト最上段をファンデーションに置く
func (ai *AuldLangSyneInteractor) PlayWasteToFoundation(wasteIdx, fIdx int) string {
	return execAndPresent(ai.Game, ai.p, func() error { return ai.Game.PlayWasteToFoundation(wasteIdx, fIdx) })
}

// Hint ヒント取得
func (ai *AuldLangSyneInteractor) Hint() string {
	return ai.p.HintOutput(ai.Game)
}

// ActionLog 棋譜を出力する
func (ai *AuldLangSyneInteractor) ActionLog() string {
	return ai.p.ActionLogOutput(ai.Game)
}

// RestoreAuldLangSyneInteractor deserialises JSON into an AuldLangSyneInteractor.
func RestoreAuldLangSyneInteractor(data []byte, p presenter.AuldLangSynePresenter) (*AuldLangSyneInteractor, error) {
	return restoreAndBuild[domain.AuldLangSyne](data, func(g *domain.AuldLangSyne) *AuldLangSyneInteractor {
		return &AuldLangSyneInteractor{
			GameBase:         GameBase[interfaces.AuldLangSyneGame]{Game: g},
			solitaireActions: newSolitaireActions[interfaces.AuldLangSyneGame](g, p),
			p:                p,
		}
	})
}
