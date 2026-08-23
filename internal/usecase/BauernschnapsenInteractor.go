//go:build !js || !wasm || extra

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// BauernschnapsenInteractorIF バウエルンシュナプセンインタラクターインタフェース
type BauernschnapsenInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.BauernschnapsenConfig) string
	// DeclareContract 人間の契約を宣言する
	DeclareContract(c domain.BauernschnapsenContract, trumpSuit int) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// DeclareMarriage マリアージュを宣言する
	DeclareMarriage(cardIndex int) string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound ラウンドをスコアリングして次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.BauernschnapsenConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// BauernschnapsenInteractor バウエルンシュナプセンインタラクタークラス
type BauernschnapsenInteractor struct {
	GameBase[interfaces.BauernschnapsenGame]
	gp presenter.BauernschnapsenPresenter
}

// NewBauernschnapsenInteractor コンストラクタ
func NewBauernschnapsenInteractor(g interfaces.BauernschnapsenGame, gp presenter.BauernschnapsenPresenter) *BauernschnapsenInteractor {
	mustNotNil("BauernschnapsenInteractor", map[string]any{"g": g, "gp": gp})
	return &BauernschnapsenInteractor{GameBase: GameBase[interfaces.BauernschnapsenGame]{Game: g}, gp: gp}
}

// Reset ゲーム初期化
func (gi *BauernschnapsenInteractor) Reset() string {
	gi.Game.Reset()
	gi.runContractTurns()
	gi.runCpuTurns()
	return gi.gp.Output(gi.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (gi *BauernschnapsenInteractor) ResetWithConfig(cfg domain.BauernschnapsenConfig) string {
	return resetWithValidatedConfig(gi.Game, gi.gp, cfg, gi.Game.SetConfig, gi.Reset)
}

// DeclareContract は人間 (席 0) の契約を宣言し、残りの CPU 席を進める。
//
// **これが無いと盤面は最初の手番で固まる。** Reset 直後は契約フェーズで、
// 契約が決まるまで PlayerPlay は「フェーズが違う」で弾かれる。
func (gi *BauernschnapsenInteractor) DeclareContract(c domain.BauernschnapsenContract, trumpSuit int) string {
	if err := gi.Game.DeclareContract(0, c, trumpSuit); err != nil {
		return gi.gp.Output(gi.Game, err)
	}
	gi.runContractTurns()
	gi.runCpuTurns()
	return gi.gp.Output(gi.Game, nil)
}

// runContractTurns は人間の手番に当たるまで CPU の宣言を進める。
//
// 上限は席数。CpuDeclareContract が手番を進めなかった場合に無限ループへ
// 落ちないようにするための保険で、正常時は席数未満で抜ける。
func (gi *BauernschnapsenInteractor) runContractTurns() {
	for i := 0; i < domain.BauernschnapsenPlayerCnt; i++ {
		if gi.Game.GetPhase() != domain.BauernschnapsenPhaseContract || gi.Game.IsHumanContractTurn() {
			return
		}
		before := gi.Game.GetCurrentPlayerIdx()
		gi.Game.CpuDeclareContract()
		if gi.Game.GetCurrentPlayerIdx() == before &&
			gi.Game.GetPhase() == domain.BauernschnapsenPhaseContract {
			return
		}
	}
}

// Play カードをプレイ
func (gi *BauernschnapsenInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(gi.Game, gi.gp); blocked {
		return out
	}
	err := gi.Game.PlayerPlay(cardIndex)
	if err != nil {
		return gi.gp.Output(gi.Game, err)
	}
	gi.runCpuTurns()
	return gi.gp.Output(gi.Game, nil)
}

// DeclareMarriage マリアージュを宣言する
func (gi *BauernschnapsenInteractor) DeclareMarriage(cardIndex int) string {
	if out, blocked := guardNotPlayable(gi.Game, gi.gp); blocked {
		return out
	}
	err := gi.Game.PlayerDeclareMarriage(cardIndex)
	if err != nil {
		return gi.gp.Output(gi.Game, err)
	}
	gi.runCpuTurns()
	return gi.gp.Output(gi.Game, nil)
}

// NextTrick トリックを解決して次のトリックへ進む
func (gi *BauernschnapsenInteractor) NextTrick() string {
	gi.Game.ResolveTrick()
	if out, blocked := guardGameEnd(gi.Game, gi.gp); blocked {
		return out
	}
	gi.Game.NextTrick()
	gi.runCpuTurns()
	return gi.gp.Output(gi.Game, nil)
}

// NextRound ラウンドをスコアリングして次のラウンドへ進む
func (gi *BauernschnapsenInteractor) NextRound() string {
	gi.Game.ScoreRound()
	if out, blocked := guardGameEnd(gi.Game, gi.gp); blocked {
		return out
	}
	gi.Game.NextRound()
	gi.runContractTurns()
	gi.runCpuTurns()
	return gi.gp.Output(gi.Game, nil)
}

// GetConfig 現在の設定を取得
func (gi *BauernschnapsenInteractor) GetConfig() domain.BauernschnapsenConfig {
	return gi.Game.GetConfig()
}

// Hint ヒント取得
func (gi *BauernschnapsenInteractor) Hint() string {
	return gi.gp.HintOutput(gi.Game)
}

// ActionLog 棋譜を出力する
func (gi *BauernschnapsenInteractor) ActionLog() string {
	return gi.gp.ActionLogOutput(gi.Game)
}

// runCpuTurns プレイフェーズでCPUターンを自動実行する
func (gi *BauernschnapsenInteractor) runCpuTurns() {
	runCpuTurnsLoop(gi.Game, trickPhases[domain.BauernschnapsenPhase]{
		play:     domain.BauernschnapsenPhasePlay,
		trickEnd: domain.BauernschnapsenPhaseTrickEnd,
		roundEnd: domain.BauernschnapsenPhaseRoundEnd,
		gameEnd:  domain.BauernschnapsenPhaseGameEnd,
	})
}

// RestoreBauernschnapsenInteractor deserialises JSON into a BauernschnapsenInteractor.
func RestoreBauernschnapsenInteractor(data []byte, gp presenter.BauernschnapsenPresenter) (*BauernschnapsenInteractor, error) {
	g, err := restoreGame[domain.Bauernschnapsen](data)
	if err != nil {
		return nil, err
	}
	return &BauernschnapsenInteractor{GameBase: GameBase[interfaces.BauernschnapsenGame]{Game: g}, gp: gp}, nil
}
