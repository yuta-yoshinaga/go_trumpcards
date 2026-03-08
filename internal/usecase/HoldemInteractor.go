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
	GetConfig() domain.HoldemConfig
	Rebuy() string
	SkipRebuy() string
	Addon() string
	SkipAddon() string
}

// HoldemInteractor テキサスホールデムインタラクタークラス
type HoldemInteractor struct {
	h  interfaces.HoldemGame
	hp presenter.HoldemPresenter
}

// NewHoldemInteractor コンストラクタ
func NewHoldemInteractor(h interfaces.HoldemGame, hp presenter.HoldemPresenter) *HoldemInteractor {
	mustNotNil("HoldemInteractor", map[string]any{"h": h, "hp": hp})
	return &HoldemInteractor{h: h, hp: hp}
}

// Reset ゲーム初期化
func (hi *HoldemInteractor) Reset() string {
	err := hi.h.Reset()
	return hi.hp.Output(hi.h, err)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (hi *HoldemInteractor) ResetWithConfig(cfg domain.HoldemConfig) string {
	// テーブルサイズ変更時はプレイヤーを再構築
	if cfg.TableSize > 0 && cfg.TableSize != hi.h.GetPlayerCnt() {
		hi.h.Resize(domain.NewPlayersForTable(cfg.TableSize))
	}
	hi.h.SetConfig(cfg)
	err := hi.h.Reset()
	return hi.hp.Output(hi.h, err)
}

// Action プレイヤーアクション実行
func (hi *HoldemInteractor) Action(action int, amount int) string {
	err := hi.h.PlayerAction(action, amount)
	return hi.hp.Output(hi.h, err)
}

// GetConfig 現在の設定を取得
func (hi *HoldemInteractor) GetConfig() domain.HoldemConfig {
	return hi.h.GetConfig()
}

// Rebuy リバイ実行
func (hi *HoldemInteractor) Rebuy() string {
	err := hi.h.Rebuy()
	return hi.hp.Output(hi.h, err)
}

// SkipRebuy リバイ辞退
func (hi *HoldemInteractor) SkipRebuy() string {
	err := hi.h.SkipRebuy()
	return hi.hp.Output(hi.h, err)
}

// Addon アドオン実行
func (hi *HoldemInteractor) Addon() string {
	err := hi.h.Addon()
	return hi.hp.Output(hi.h, err)
}

// SkipAddon アドオン辞退
func (hi *HoldemInteractor) SkipAddon() string {
	err := hi.h.SkipAddon()
	return hi.hp.Output(hi.h, err)
}
