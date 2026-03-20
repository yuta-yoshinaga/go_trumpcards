package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// OmahaInteractorIF オマハホールデムインタラクターインタフェース
type OmahaInteractorIF interface {
	Reset() string
	ResetWithConfig(cfg domain.OmahaConfig) string
	Action(action int, amount int) string
	GetConfig() domain.OmahaConfig
	Rebuy() string
	SkipRebuy() string
	Addon() string
	SkipAddon() string
	Muck() string
	ShowHand() string
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
	err := oi.o.Reset()
	return oi.op.Output(oi.o, err)
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
	err := oi.o.PlayerAction(action, amount)
	return oi.op.Output(oi.o, err)
}

// GetConfig 現在の設定を取得
func (oi *OmahaInteractor) GetConfig() domain.OmahaConfig {
	return oi.o.GetConfig()
}

// Rebuy リバイ実行
func (oi *OmahaInteractor) Rebuy() string {
	err := oi.o.Rebuy()
	return oi.op.Output(oi.o, err)
}

// SkipRebuy リバイ辞退
func (oi *OmahaInteractor) SkipRebuy() string {
	err := oi.o.SkipRebuy()
	return oi.op.Output(oi.o, err)
}

// Addon アドオン実行
func (oi *OmahaInteractor) Addon() string {
	err := oi.o.Addon()
	return oi.op.Output(oi.o, err)
}

// SkipAddon アドオン辞退
func (oi *OmahaInteractor) SkipAddon() string {
	err := oi.o.SkipAddon()
	return oi.op.Output(oi.o, err)
}

// Muck マック
func (oi *OmahaInteractor) Muck() string {
	err := oi.o.Muck()
	return oi.op.Output(oi.o, err)
}

// ShowHand ハンドを公開する
func (oi *OmahaInteractor) ShowHand() string {
	err := oi.o.ShowHand()
	return oi.op.Output(oi.o, err)
}

// ActionLog 棋譜を出力する
func (oi *OmahaInteractor) ActionLog() string {
	return oi.op.ActionLogOutput(oi.o)
}
