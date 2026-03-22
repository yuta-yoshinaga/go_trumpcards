package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// OmahaInteractorIF オマハホールデムインタラクターインタフェース
type OmahaInteractorIF interface {
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.OmahaConfig) string
	// Action プレイヤーアクション実行
	Action(action int, amount int) string
	// GetConfig 現在の設定を取得
	GetConfig() domain.OmahaConfig
	// Rebuy リバイ実行
	Rebuy() string
	// SkipRebuy リバイ辞退
	SkipRebuy() string
	// Addon アドオン実行
	Addon() string
	// SkipAddon アドオン辞退
	SkipAddon() string
	// Muck マック
	Muck() string
	// ShowHand ハンドを公開する
	ShowHand() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// OmahaInteractor オマハホールデムインタラクタークラス
type OmahaInteractor struct {
	o  interfaces.OmahaGame
	op presenter.OmahaPresenter
}

// NewOmahaInteractor コンストラクタ
func NewOmahaInteractor(o interfaces.OmahaGame, op presenter.OmahaPresenter) *OmahaInteractor {
	mustNotNil("OmahaInteractor", map[string]any{"o": o, "op": op})
	return &OmahaInteractor{o: o, op: op}
}

// Reset ゲーム初期化
func (oi *OmahaInteractor) Reset() string {
	return execAndPresent(oi.o, oi.op, oi.o.Reset)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (oi *OmahaInteractor) ResetWithConfig(cfg domain.OmahaConfig) string {
	if err := cfg.Validate(); err != nil {
		return oi.op.Output(oi.o, err)
	}
	if cfg.TableSize > 0 && cfg.TableSize != oi.o.GetPlayerCnt() {
		oi.o.Resize(domain.NewOmahaPlayersForTable(cfg.TableSize))
	}
	oi.o.SetConfig(cfg)
	err := oi.o.Reset()
	return oi.op.Output(oi.o, err)
}

// Action プレイヤーアクション実行
func (oi *OmahaInteractor) Action(action int, amount int) string {
	return execAndPresent(oi.o, oi.op, func() error { return oi.o.PlayerAction(action, amount) })
}

// GetConfig 現在の設定を取得
func (oi *OmahaInteractor) GetConfig() domain.OmahaConfig {
	return oi.o.GetConfig()
}

// Rebuy リバイ実行
func (oi *OmahaInteractor) Rebuy() string {
	return execAndPresent(oi.o, oi.op, oi.o.Rebuy)
}

// SkipRebuy リバイ辞退
func (oi *OmahaInteractor) SkipRebuy() string {
	return execAndPresent(oi.o, oi.op, oi.o.SkipRebuy)
}

// Addon アドオン実行
func (oi *OmahaInteractor) Addon() string {
	return execAndPresent(oi.o, oi.op, oi.o.Addon)
}

// SkipAddon アドオン辞退
func (oi *OmahaInteractor) SkipAddon() string {
	return execAndPresent(oi.o, oi.op, oi.o.SkipAddon)
}

// Muck マック
func (oi *OmahaInteractor) Muck() string {
	return execAndPresent(oi.o, oi.op, oi.o.Muck)
}

// ShowHand ハンドを公開する
func (oi *OmahaInteractor) ShowHand() string {
	return execAndPresent(oi.o, oi.op, oi.o.ShowHand)
}

// ActionLog 棋譜を出力する
func (oi *OmahaInteractor) ActionLog() string {
	return oi.op.ActionLogOutput(oi.o)
}
