package usecase

import (
	"encoding/json"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// PineappleInteractorIF パイナップルポーカーインタラクターインタフェース
type PineappleInteractorIF interface {
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化 (profileData: JSONプロファイル、nilなら無視)
	ResetWithConfig(cfg domain.PineappleConfig, profileData []byte) string
	// Action プレイヤーアクション実行
	Action(action int, amount int, humanPlayMs int) string
	// Discard ディスカード実行
	Discard(cardIdx int) string
	// GetConfig 現在の設定を取得
	GetConfig() domain.PineappleConfig
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

// PineappleInteractor パイナップルポーカーインタラクタークラス
type PineappleInteractor struct {
	p  interfaces.PineappleGame
	pp presenter.PineapplePresenter
}

// NewPineappleInteractor コンストラクタ
func NewPineappleInteractor(p interfaces.PineappleGame, pp presenter.PineapplePresenter) *PineappleInteractor {
	mustNotNil("PineappleInteractor", map[string]any{"p": p, "pp": pp})
	return &PineappleInteractor{p: p, pp: pp}
}

// Reset ゲーム初期化
func (pi *PineappleInteractor) Reset() string {
	return execAndPresent(pi.p, pi.pp, pi.p.Reset)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (pi *PineappleInteractor) ResetWithConfig(cfg domain.PineappleConfig, profileData []byte) string {
	if err := cfg.Validate(); err != nil {
		return pi.pp.Output(pi.p, err)
	}
	// テーブルサイズ変更時はプレイヤーを再構築
	if cfg.TableSize > 0 && cfg.TableSize != pi.p.GetPlayerCnt() {
		pi.p.Resize(domain.NewPineapplePlayersForTable(cfg.TableSize))
	}
	pi.p.SetConfig(cfg)
	err := pi.p.Reset()
	if len(profileData) > 0 {
		_ = pi.p.ImportProfile(profileData)
	}
	return pi.pp.Output(pi.p, err)
}

// Action プレイヤーアクション実行
func (pi *PineappleInteractor) Action(action int, amount int, humanPlayMs int) string {
	return execAndPresent(pi.p, pi.pp, func() error { return pi.p.PlayerAction(action, amount, humanPlayMs) })
}

// Discard ディスカード実行
func (pi *PineappleInteractor) Discard(cardIdx int) string {
	return execAndPresent(pi.p, pi.pp, func() error { return pi.p.DiscardCard(cardIdx) })
}

// GetConfig 現在の設定を取得
func (pi *PineappleInteractor) GetConfig() domain.PineappleConfig {
	return pi.p.GetConfig()
}

// Rebuy リバイ実行
func (pi *PineappleInteractor) Rebuy() string {
	return execAndPresent(pi.p, pi.pp, pi.p.Rebuy)
}

// SkipRebuy リバイ辞退
func (pi *PineappleInteractor) SkipRebuy() string {
	return execAndPresent(pi.p, pi.pp, pi.p.SkipRebuy)
}

// Addon アドオン実行
func (pi *PineappleInteractor) Addon() string {
	return execAndPresent(pi.p, pi.pp, pi.p.Addon)
}

// SkipAddon アドオン辞退
func (pi *PineappleInteractor) SkipAddon() string {
	return execAndPresent(pi.p, pi.pp, pi.p.SkipAddon)
}

// Muck マック (ハンドを伏せる)
func (pi *PineappleInteractor) Muck() string {
	return execAndPresent(pi.p, pi.pp, pi.p.Muck)
}

// ShowHand ハンドを公開する
func (pi *PineappleInteractor) ShowHand() string {
	return execAndPresent(pi.p, pi.pp, pi.p.ShowHand)
}

// ActionLog 棋譜を出力する
func (pi *PineappleInteractor) ActionLog() string {
	return pi.pp.ActionLogOutput(pi.p)
}

// Snapshot serialises the game state to JSON for KV persistence.
func (pi *PineappleInteractor) Snapshot() ([]byte, error) {
	return json.Marshal(pi.p)
}

// RestorePineappleInteractor deserialises JSON into a PineappleInteractor.
func RestorePineappleInteractor(data []byte, pp presenter.PineapplePresenter) (*PineappleInteractor, error) {
	p, err := restoreGame[domain.Pineapple](data)
	if err != nil {
		return nil, err
	}
	return &PineappleInteractor{p: p, pp: pp}, nil
}
