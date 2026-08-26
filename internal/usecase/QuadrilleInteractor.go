//go:build !js || !wasm || extra4

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// QuadrilleInteractorIF カドリール (Quadrille) のインタラクターインタフェース
type QuadrilleInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.QuadrilleConfig) string
	// Bid ビッドする (pass/entrar/solo + 切り札スート)
	Bid(bid domain.QuadrilleBid, trumpSuit int) string
	// CallKing 落札者が味方を呼ぶ王を指名する
	CallKing(suit int) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound ディールをスコアリングして次のディールへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.QuadrilleConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// QuadrilleInteractor カドリールのインタラクタークラス
type QuadrilleInteractor struct {
	GameBase[interfaces.QuadrilleGame]
	tp presenter.QuadrillePresenter
}

// NewQuadrilleInteractor コンストラクタ
func NewQuadrilleInteractor(g interfaces.QuadrilleGame, tp presenter.QuadrillePresenter) *QuadrilleInteractor {
	mustNotNil("QuadrilleInteractor", map[string]any{"g": g, "tp": tp})
	return &QuadrilleInteractor{GameBase: GameBase[interfaces.QuadrilleGame]{Game: g}, tp: tp}
}

// quadrilleTrickPhases Quadrille のトリックフェーズ定数
func quadrilleTrickPhases() trickPhases[domain.QuadrillePhase] {
	return trickPhases[domain.QuadrillePhase]{
		play:     domain.QuadrillePhasePlay,
		trickEnd: domain.QuadrillePhaseTrickEnd,
		roundEnd: domain.QuadrillePhaseRoundEnd,
		gameEnd:  domain.QuadrillePhaseGameEnd,
	}
}

// Reset ゲーム初期化
func (ci *QuadrilleInteractor) Reset() string {
	ci.Game.Reset()
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ci *QuadrilleInteractor) ResetWithConfig(cfg domain.QuadrilleConfig) string {
	return resetWithValidatedConfig(ci.Game, ci.tp, cfg, ci.Game.SetConfig, ci.Reset)
}

// Bid ビッドする
func (ci *QuadrilleInteractor) Bid(bid domain.QuadrilleBid, trumpSuit int) string {
	if out, blocked := guardGameEnd(ci.Game, ci.tp); blocked {
		return out
	}
	if err := ci.Game.PlayerBid(bid, trumpSuit); err != nil {
		return ci.tp.Output(ci.Game, err)
	}
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// CallKing は人間の落札者が味方を呼ぶ王を指名する。
//
// **これが無いと盤面は落札の直後で固まる。** 王呼びフェーズを抜けるまで
// PlayerPlay は「フェーズが違う」で弾かれ続ける。
func (ci *QuadrilleInteractor) CallKing(suit int) string {
	if out, blocked := guardGameEnd(ci.Game, ci.tp); blocked {
		return out
	}
	if err := ci.Game.DeclareKing(ci.Game.GetQuadrilleIdx(), suit); err != nil {
		return ci.tp.Output(ci.Game, err)
	}
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// Play カードをプレイ
func (ci *QuadrilleInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(ci.Game, ci.tp); blocked {
		return out
	}
	if err := ci.Game.PlayerPlay(cardIndex); err != nil {
		return ci.tp.Output(ci.Game, err)
	}
	// 人間が最後のカードを出してトリック完了した場合、即座に解決する。
	if ci.Game.GetPhase() == domain.QuadrillePhaseTrickEnd {
		ci.Game.ResolveTrick()
	}
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// NextTrick 次のトリックへ進む
func (ci *QuadrilleInteractor) NextTrick() string {
	ci.Game.NextTrick()
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// NextRound ディールをスコアリングして次のディールへ進む
func (ci *QuadrilleInteractor) NextRound() string {
	ci.Game.ScoreRound()
	if out, blocked := guardGameEnd(ci.Game, ci.tp); blocked {
		return out
	}
	ci.Game.NextRound()
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// GetConfig 現在の設定を取得
func (ci *QuadrilleInteractor) GetConfig() domain.QuadrilleConfig {
	return ci.Game.GetConfig()
}

// Hint ヒント取得
func (ci *QuadrilleInteractor) Hint() string {
	return ci.tp.HintOutput(ci.Game)
}

// ActionLog 棋譜を出力する
func (ci *QuadrilleInteractor) ActionLog() string {
	return ci.tp.ActionLogOutput(ci.Game)
}

// advance CPU のビッド・プレイを人間の手番もしくはトリック/ラウンド終了になるまで自動実行する。
func (ci *QuadrilleInteractor) advance() {
	runCpuBidsLoop[domain.QuadrillePhase](ci.Game, domain.QuadrillePhaseBid)
	ci.runKingCall()
	runCpuTurnsLoop(ci.Game, quadrilleTrickPhases())
}

// runKingCall は CPU が落札していれば王呼びを進める。
//
// 人間が落札していれば CallKing を待って止まる。落札者が王 4 枚を全部
// 持っていた場合 (Roi seul) は startKingCall がそのままプレイへ進めるので、
// ここには来ない。
func (ci *QuadrilleInteractor) runKingCall() {
	if ci.Game.GetPhase() != domain.QuadrillePhaseKingCall || ci.Game.IsHumanKingCallTurn() {
		return
	}
	ci.Game.CpuDeclareKing()
}

// RestoreQuadrilleInteractor deserialises JSON into an QuadrilleInteractor.
func RestoreQuadrilleInteractor(data []byte, tp presenter.QuadrillePresenter) (*QuadrilleInteractor, error) {
	return restoreAndBuild[domain.Quadrille](data, func(g *domain.Quadrille) *QuadrilleInteractor {
		return &QuadrilleInteractor{GameBase: GameBase[interfaces.QuadrilleGame]{Game: g}, tp: tp}
	})
}
