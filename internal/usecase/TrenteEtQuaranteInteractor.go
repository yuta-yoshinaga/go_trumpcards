//go:build !js || !wasm || extra4

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// TrenteEtQuaranteInteractorIF はトラント・エ・カラント (Trente et Quarante) の
// インタラクターインタフェース。
type TrenteEtQuaranteInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化 (新規ゲーム)
	Reset() string
	// ResetWithConfig 設定を変更してゲームを初期化
	ResetWithConfig(cfg domain.TrenteEtQuaranteConfig) string
	// Bet ベット種別とステークを賭けてラウンドを解決する
	Bet(bet domain.TrenteEtQuaranteBet, stake int) string
	// NextRound 次のラウンドを開始する
	NextRound() string
	// GetConfig 現在の設定を返す
	GetConfig() domain.TrenteEtQuaranteConfig
	// Hint ヒントを出力する
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// TrenteEtQuaranteInteractor はトラント・エ・カラントのインタラクター。
type TrenteEtQuaranteInteractor struct {
	GameBase[interfaces.TrenteEtQuaranteGame]
	cp presenter.TrenteEtQuarantePresenter
}

// NewTrenteEtQuaranteInteractor コンストラクタ。
func NewTrenteEtQuaranteInteractor(bg interfaces.TrenteEtQuaranteGame, cp presenter.TrenteEtQuarantePresenter) *TrenteEtQuaranteInteractor {
	mustNotNil("TrenteEtQuaranteInteractor", map[string]any{"bg": bg, "cp": cp})
	return &TrenteEtQuaranteInteractor{GameBase: GameBase[interfaces.TrenteEtQuaranteGame]{Game: bg}, cp: cp}
}

// Reset ゲーム初期化 (新規ゲーム)。
func (bi *TrenteEtQuaranteInteractor) Reset() string {
	return runAndPresent(bi.Game, bi.cp, bi.Game.Reset)
}

// ResetWithConfig 設定を変更してゲームを初期化。
func (bi *TrenteEtQuaranteInteractor) ResetWithConfig(cfg domain.TrenteEtQuaranteConfig) string {
	return resetWithValidatedConfig(bi.Game, bi.cp, cfg, bi.Game.SetConfig, bi.Reset)
}

// Bet ベット種別とステークを賭けてラウンドを解決する。
func (bi *TrenteEtQuaranteInteractor) Bet(bet domain.TrenteEtQuaranteBet, stake int) string {
	return execAndPresent(bi.Game, bi.cp, func() error {
		return bi.Game.PlaceBet(bet, stake)
	})
}

// NextRound 次のラウンドを開始する。
func (bi *TrenteEtQuaranteInteractor) NextRound() string {
	return runAndPresent(bi.Game, bi.cp, bi.Game.NextRound)
}

// GetConfig 現在の設定を返す。
func (bi *TrenteEtQuaranteInteractor) GetConfig() domain.TrenteEtQuaranteConfig {
	return bi.Game.GetConfig()
}

// Hint ヒントを出力する。
func (bi *TrenteEtQuaranteInteractor) Hint() string { return bi.cp.HintOutput(bi.Game) }

// ActionLog 棋譜を出力する。
func (bi *TrenteEtQuaranteInteractor) ActionLog() string { return bi.cp.ActionLogOutput(bi.Game) }

// RestoreTrenteEtQuaranteInteractor deserialises JSON into a TrenteEtQuaranteInteractor.
func RestoreTrenteEtQuaranteInteractor(data []byte, cp presenter.TrenteEtQuarantePresenter) (*TrenteEtQuaranteInteractor, error) {
	return restoreAndBuild[domain.TrenteEtQuarante](data, func(g *domain.TrenteEtQuarante) *TrenteEtQuaranteInteractor {
		return &TrenteEtQuaranteInteractor{GameBase: GameBase[interfaces.TrenteEtQuaranteGame]{Game: g}, cp: cp}
	})
}
