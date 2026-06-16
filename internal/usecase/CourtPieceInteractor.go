//go:build !js || !wasm || casino

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// CourtPieceInteractorIF Court Piece インタラクターインタフェース
type CourtPieceInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.CourtPieceConfig) string
	// DeclareTrump トランプスートを宣言
	DeclareTrump(suit int) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound ラウンドをスコアリングして次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.CourtPieceConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// CourtPieceInteractor Court Piece インタラクタークラス
type CourtPieceInteractor struct {
	GameBase[interfaces.CourtPieceGame]
	tp presenter.CourtPiecePresenter
}

// NewCourtPieceInteractor コンストラクタ
func NewCourtPieceInteractor(t interfaces.CourtPieceGame, tp presenter.CourtPiecePresenter) *CourtPieceInteractor {
	mustNotNil("CourtPieceInteractor", map[string]any{"t": t, "tp": tp})
	return &CourtPieceInteractor{GameBase: GameBase[interfaces.CourtPieceGame]{Game: t}, tp: tp}
}

// Reset ゲーム初期化
func (ti *CourtPieceInteractor) Reset() string {
	ti.Game.Reset()
	ti.runCpuTrumpDeclaration()
	if ti.Game.GetPhase() == domain.CourtPiecePhasePlay {
		ti.runCpuTurns()
	}
	return ti.tp.Output(ti.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ti *CourtPieceInteractor) ResetWithConfig(cfg domain.CourtPieceConfig) string {
	return resetWithValidatedConfig(ti.Game, ti.tp, cfg, ti.Game.SetConfig, ti.Reset)
}

// DeclareTrump トランプスートを宣言
func (ti *CourtPieceInteractor) DeclareTrump(suit int) string {
	if out, blocked := guardGameEnd(ti.Game, ti.tp); blocked {
		return out
	}
	if err := ti.Game.PlayerDeclareTrump(suit); err != nil {
		return ti.tp.Output(ti.Game, err)
	}
	if ti.Game.GetPhase() == domain.CourtPiecePhasePlay {
		ti.runCpuTurns()
	}
	return ti.tp.Output(ti.Game, nil)
}

// Play カードをプレイ
func (ti *CourtPieceInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(ti.Game, ti.tp); blocked {
		return out
	}
	if err := ti.Game.PlayerPlay(cardIndex); err != nil {
		return ti.tp.Output(ti.Game, err)
	}
	// 人間が最後 (4枚目) のカードを出してトリックが完了した場合、即座に解決する。
	// PlayerPlay → playCard が phase を TrickEnd にした直後に runCpuTurns へ抜けると、
	// runCpuTurnsLoop はトリック解決を行わずに break してしまい、leadPlayerIdx と
	// トリック数の更新が漏れる (Tarneeb/Whist/OhHell/Bridge と同じ理由で同じ処置を入れる)。
	if ti.Game.GetPhase() == domain.CourtPiecePhaseTrickEnd {
		ti.Game.ResolveTrick()
	}
	ti.runCpuTurns()
	return ti.tp.Output(ti.Game, nil)
}

// NextTrick 次のトリックへ進む
func (ti *CourtPieceInteractor) NextTrick() string {
	if out, blocked := guardGameEnd(ti.Game, ti.tp); blocked {
		return out
	}
	ti.Game.NextTrick()
	ti.runCpuTurns()
	return ti.tp.Output(ti.Game, nil)
}

// NextRound ラウンドをスコアリングして次のラウンドへ進む
func (ti *CourtPieceInteractor) NextRound() string {
	ti.Game.ScoreRound()
	if out, blocked := guardGameEnd(ti.Game, ti.tp); blocked {
		return out
	}
	ti.Game.NextRound()
	ti.runCpuTrumpDeclaration()
	if ti.Game.GetPhase() == domain.CourtPiecePhasePlay {
		ti.runCpuTurns()
	}
	return ti.tp.Output(ti.Game, nil)
}

// GetConfig 現在の設定を取得
func (ti *CourtPieceInteractor) GetConfig() domain.CourtPieceConfig {
	return ti.Game.GetConfig()
}

// Hint ヒント取得
func (ti *CourtPieceInteractor) Hint() string {
	return ti.tp.HintOutput(ti.Game)
}

// ActionLog 棋譜を出力する
func (ti *CourtPieceInteractor) ActionLog() string {
	return ti.tp.ActionLogOutput(ti.Game)
}

// runCpuTrumpDeclaration トランプ宣言フェーズが続く限り CPU をループ実行する。
// 呼び手 (Hakim) が CPU の間、宣言を自動処理する。
func (ti *CourtPieceInteractor) runCpuTrumpDeclaration() {
	const maxIterations = domain.CourtPiecePlayerCnt
	for i := 0; i < maxIterations; i++ {
		if ti.Game.GetPhase() != domain.CourtPiecePhaseTrumpDeclaration {
			return
		}
		if ti.Game.IsHumanTrumpTurn() {
			return
		}
		ti.Game.CpuDeclareTrump()
	}
}

// runCpuTurns ヒューマンの手番またはトリック / ラウンド終了まで CPU ターンを実行
func (ti *CourtPieceInteractor) runCpuTurns() {
	runCpuTurnsLoop(ti.Game, trickPhases[domain.CourtPiecePhase]{
		play:     domain.CourtPiecePhasePlay,
		trickEnd: domain.CourtPiecePhaseTrickEnd,
		roundEnd: domain.CourtPiecePhaseRoundEnd,
		gameEnd:  domain.CourtPiecePhaseGameEnd,
	})
}

// RestoreCourtPieceInteractor deserialises JSON into a CourtPieceInteractor.
func RestoreCourtPieceInteractor(data []byte, tp presenter.CourtPiecePresenter) (*CourtPieceInteractor, error) {
	return restoreAndBuild[domain.CourtPiece](data, func(g *domain.CourtPiece) *CourtPieceInteractor {
		return &CourtPieceInteractor{GameBase: GameBase[interfaces.CourtPieceGame]{Game: g}, tp: tp}
	})
}
