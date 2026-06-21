//go:build !js || !wasm || casino

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// OpenFaceChineseInteractorIF オープンフェイス・チャイニーズポーカー (OFC) のインタラクターインタフェース
type OpenFaceChineseInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.OpenFaceChineseConfig) string
	// Place 保留カードを指定段に置く (0=front,1=middle,2=back)
	Place(row int) string
	// NextRound 次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.OpenFaceChineseConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// ofcCpuGuard caps the per-call CPU placement loop to avoid any infinite loop.
const ofcCpuGuard = 1000

// OpenFaceChineseInteractor オープンフェイス・チャイニーズポーカー (OFC) のインタラクタークラス
type OpenFaceChineseInteractor struct {
	GameBase[interfaces.OpenFaceChineseGame]
	sp presenter.OpenFaceChinesePresenter
}

// NewOpenFaceChineseInteractor コンストラクタ
func NewOpenFaceChineseInteractor(g interfaces.OpenFaceChineseGame, sp presenter.OpenFaceChinesePresenter) *OpenFaceChineseInteractor {
	mustNotNil("OpenFaceChineseInteractor", map[string]any{"g": g, "sp": sp})
	return &OpenFaceChineseInteractor{GameBase: GameBase[interfaces.OpenFaceChineseGame]{Game: g}, sp: sp}
}

// Reset ゲーム初期化
func (ti *OpenFaceChineseInteractor) Reset() string {
	ti.Game.Reset()
	ti.advanceCpu()
	return ti.sp.Output(ti.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ti *OpenFaceChineseInteractor) ResetWithConfig(cfg domain.OpenFaceChineseConfig) string {
	return resetWithValidatedConfig(ti.Game, ti.sp, cfg, ti.Game.SetConfig, ti.Reset)
}

// Place 保留カードを指定段に置く
func (ti *OpenFaceChineseInteractor) Place(row int) string {
	if out, blocked := guardNotPlayable(ti.Game, ti.sp); blocked {
		return out
	}
	if err := ti.Game.PlayerPlace(row); err != nil {
		return ti.sp.Output(ti.Game, err)
	}
	ti.advanceCpu()
	return ti.sp.Output(ti.Game, nil)
}

// NextRound 次のラウンドへ進む
func (ti *OpenFaceChineseInteractor) NextRound() string {
	if out, blocked := guardGameEnd(ti.Game, ti.sp); blocked {
		return out
	}
	ti.Game.NextRound()
	ti.advanceCpu()
	return ti.sp.Output(ti.Game, nil)
}

// GetConfig 現在の設定を取得
func (ti *OpenFaceChineseInteractor) GetConfig() domain.OpenFaceChineseConfig {
	return ti.Game.GetConfig()
}

// Hint ヒント取得
func (ti *OpenFaceChineseInteractor) Hint() string {
	return ti.sp.HintOutput(ti.Game)
}

// ActionLog 棋譜を出力する
func (ti *OpenFaceChineseInteractor) ActionLog() string {
	return ti.sp.ActionLogOutput(ti.Game)
}

// advanceCpu 配置フェーズで CPU の手番が続く限り自動配置する。人間の手番か、
// 配置フェーズ以外（採点済み／ゲーム終了）になったら止まる。
func (ti *OpenFaceChineseInteractor) advanceCpu() {
	for i := 0; i < ofcCpuGuard; i++ {
		if ti.Game.GetGameEndFlag() || ti.Game.GetPhase() != domain.OpenFaceChinesePhasePlacing {
			return
		}
		if ti.Game.IsHumanTurn() {
			return
		}
		ti.Game.CpuPlay()
	}
}

// RestoreOpenFaceChineseInteractor deserialises JSON into an OpenFaceChineseInteractor.
func RestoreOpenFaceChineseInteractor(data []byte, sp presenter.OpenFaceChinesePresenter) (*OpenFaceChineseInteractor, error) {
	return restoreAndBuild[domain.OpenFaceChinese](data, func(g *domain.OpenFaceChinese) *OpenFaceChineseInteractor {
		return &OpenFaceChineseInteractor{GameBase: GameBase[interfaces.OpenFaceChineseGame]{Game: g}, sp: sp}
	})
}
