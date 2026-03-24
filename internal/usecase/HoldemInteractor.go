package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// HoldemInteractorIF テキサスホールデムインタラクターインタフェース
type HoldemInteractorIF interface {
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化 (profileData: JSONプロファイル、nilなら無視)
	ResetWithConfig(cfg domain.HoldemConfig, profileData []byte) string
	// Action プレイヤーアクション実行
	Action(action int, amount int, humanPlayMs int) string
	// GetConfig 現在の設定を取得
	GetConfig() domain.HoldemConfig
	// Rebuy リバイ実行
	Rebuy() string
	// SkipRebuy リバイ辞退
	SkipRebuy() string
	// Addon アドオン実行
	Addon() string
	// SkipAddon アドオン辞退
	SkipAddon() string
	// Muck マック (ハンドを伏せる)
	Muck() string
	// ShowHand ハンドを公開する
	ShowHand() string
	// ActionLog 棋譜を出力する
	ActionLog() string
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
	return execAndPresent(hi.h, hi.hp, hi.h.Reset)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (hi *HoldemInteractor) ResetWithConfig(cfg domain.HoldemConfig, profileData []byte) string {
	if err := cfg.Validate(); err != nil {
		return hi.hp.Output(hi.h, err)
	}
	// テーブルサイズ変更時はプレイヤーを再構築
	if cfg.TableSize > 0 && cfg.TableSize != hi.h.GetPlayerCnt() {
		hi.h.Resize(domain.NewPlayersForTable(cfg.TableSize))
	}
	hi.h.SetConfig(cfg)
	err := hi.h.Reset()
	if len(profileData) > 0 {
		_ = hi.h.ImportProfile(profileData)
	}
	return hi.hp.Output(hi.h, err)
}

// Action プレイヤーアクション実行
func (hi *HoldemInteractor) Action(action int, amount int, humanPlayMs int) string {
	return execAndPresent(hi.h, hi.hp, func() error { return hi.h.PlayerAction(action, amount, humanPlayMs) })
}

// GetConfig 現在の設定を取得
func (hi *HoldemInteractor) GetConfig() domain.HoldemConfig {
	return hi.h.GetConfig()
}

// Rebuy リバイ実行
func (hi *HoldemInteractor) Rebuy() string {
	return execAndPresent(hi.h, hi.hp, hi.h.Rebuy)
}

// SkipRebuy リバイ辞退
func (hi *HoldemInteractor) SkipRebuy() string {
	return execAndPresent(hi.h, hi.hp, hi.h.SkipRebuy)
}

// Addon アドオン実行
func (hi *HoldemInteractor) Addon() string {
	return execAndPresent(hi.h, hi.hp, hi.h.Addon)
}

// SkipAddon アドオン辞退
func (hi *HoldemInteractor) SkipAddon() string {
	return execAndPresent(hi.h, hi.hp, hi.h.SkipAddon)
}

// Muck マック (ハンドを伏せる)
func (hi *HoldemInteractor) Muck() string {
	return execAndPresent(hi.h, hi.hp, hi.h.Muck)
}

// ShowHand ハンドを公開する
func (hi *HoldemInteractor) ShowHand() string {
	return execAndPresent(hi.h, hi.hp, hi.h.ShowHand)
}

// ActionLog 棋譜を出力する
func (hi *HoldemInteractor) ActionLog() string {
	return hi.hp.ActionLogOutput(hi.h)
}
