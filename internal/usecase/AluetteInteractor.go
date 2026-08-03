//go:build !js || !wasm || extra2

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// AluetteInteractorIF アリュエットのインタラクターインタフェース
//
// **Discard は無い。**タロー系から写すとスカルト用の Discard が付いてくるが、
// アリュエットには余剰札を伏せる工程が無い。
type AluetteInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.AluetteConfig) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound メーヌをスコアリングして次のメーヌへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.AluetteConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// AluetteInteractor アリュエットのインタラクタークラス
type AluetteInteractor struct {
	GameBase[interfaces.AluetteGame]
	tp presenter.AluettePresenter
}

// NewAluetteInteractor コンストラクタ
func NewAluetteInteractor(g interfaces.AluetteGame, tp presenter.AluettePresenter) *AluetteInteractor {
	mustNotNil("AluetteInteractor", map[string]any{"g": g, "tp": tp})
	return &AluetteInteractor{GameBase: GameBase[interfaces.AluetteGame]{Game: g}, tp: tp}
}

// aluetteTrickPhases アリュエットのトリックフェーズ定数
func aluetteTrickPhases() trickPhases[domain.AluettePhase] {
	return trickPhases[domain.AluettePhase]{
		play:     domain.AluettePhasePlay,
		trickEnd: domain.AluettePhaseTrickEnd,
		roundEnd: domain.AluettePhaseRoundEnd,
		gameEnd:  domain.AluettePhaseGameEnd,
	}
}

// Reset ゲーム初期化
func (ci *AluetteInteractor) Reset() string {
	ci.Game.Reset()
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ci *AluetteInteractor) ResetWithConfig(cfg domain.AluetteConfig) string {
	return resetWithValidatedConfig(ci.Game, ci.tp, cfg, ci.Game.SetConfig, ci.Reset)
}

// Play カードをプレイ
func (ci *AluetteInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(ci.Game, ci.tp); blocked {
		return out
	}
	if err := ci.Game.PlayerPlay(cardIndex); err != nil {
		return ci.tp.Output(ci.Game, err)
	}
	// 人間が最後のカードを出してトリックが揃った場合、即座に解決する。
	if ci.Game.GetPhase() == domain.AluettePhaseTrickEnd {
		ci.Game.ResolveTrick()
	}
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// NextTrick 次のトリックへ進む
func (ci *AluetteInteractor) NextTrick() string {
	ci.Game.NextTrick()
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// NextRound メーヌをスコアリングして次のメーヌへ進む
func (ci *AluetteInteractor) NextRound() string {
	ci.Game.ScoreRound()
	if out, blocked := guardGameEnd(ci.Game, ci.tp); blocked {
		return out
	}
	ci.Game.NextRound()
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// GetConfig 現在の設定を取得
func (ci *AluetteInteractor) GetConfig() domain.AluetteConfig {
	return ci.Game.GetConfig()
}

// Hint ヒント取得
func (ci *AluetteInteractor) Hint() string {
	return ci.tp.HintOutput(ci.Game)
}

// ActionLog 棋譜を出力する
func (ci *AluetteInteractor) ActionLog() string {
	return ci.tp.ActionLogOutput(ci.Game)
}

// advance CPU のプレイを人間の手番かトリック/メーヌ終了まで自動実行する。
//
// **スカルトのループは要らない。**Minchiate / Tarocchini はディーラーが捨てる
// までプレイフェーズに入らないので二段構えだったが、アリュエットは配ったら
// すぐトリックが始まる。
func (ci *AluetteInteractor) advance() {
	runCpuTurnsLoop(ci.Game, aluetteTrickPhases())
}

// RestoreAluetteInteractor deserialises JSON into an AluetteInteractor.
func RestoreAluetteInteractor(data []byte, tp presenter.AluettePresenter) (*AluetteInteractor, error) {
	return restoreAndBuild[domain.Aluette](data, func(g *domain.Aluette) *AluetteInteractor {
		return &AluetteInteractor{GameBase: GameBase[interfaces.AluetteGame]{Game: g}, tp: tp}
	})
}
