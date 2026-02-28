package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// HoldemInteractorIF テキサスホールデムインタラクターインタフェース
type HoldemInteractorIF interface {
	Reset() string
	ResetWithConfig(cfg domain.HoldemConfig) string
	Action(action int, amount int) string
}

// HoldemInteractor テキサスホールデムインタラクタークラス
type HoldemInteractor struct {
	h  interfaces.HoldemGame
	hp presenter.HoldemPresenter
}

// NewHoldemInteractor コンストラクタ
func NewHoldemInteractor(h interfaces.HoldemGame, hp presenter.HoldemPresenter) *HoldemInteractor {
	if h == nil {
		panic("HoldemInteractor: h must not be nil")
	}
	if hp == nil {
		panic("HoldemInteractor: hp must not be nil")
	}
	return &HoldemInteractor{h: h, hp: hp}
}

// Reset ゲーム初期化
func (hi *HoldemInteractor) Reset() string {
	hi.h.Reset()
	return hi.hp.Output(hi.h, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (hi *HoldemInteractor) ResetWithConfig(cfg domain.HoldemConfig) string {
	hi.h.SetConfig(cfg)
	hi.h.Reset()
	return hi.hp.Output(hi.h, nil)
}

// Action プレイヤーアクション実行
func (hi *HoldemInteractor) Action(action int, amount int) string {
	err := hi.h.PlayerAction(action, amount)
	return hi.hp.Output(hi.h, err)
}
