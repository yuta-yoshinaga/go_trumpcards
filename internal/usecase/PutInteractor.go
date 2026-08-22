//go:build !js || !wasm || extra4

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// PutInteractorIF プットインタラクターインタフェース
type PutInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset マッチ初期化
	Reset() string
	// ResetWithConfig 設定を変更してマッチ初期化
	ResetWithConfig(cfg domain.PutConfig) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// Put Put を宣言 (または再引き上げ) する
	Put() string
	// Respond Put 宣言に応答する (true=受諾 / false=拒否)
	Respond(accept bool) string
	// Next バサ / マノ終了から次へ進む
	Next() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.PutConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// PutInteractor プットインタラクタークラス
type PutInteractor struct {
	GameBase[interfaces.PutGame]
	tp presenter.PutPresenter
}

// NewPutInteractor コンストラクタ
func NewPutInteractor(g interfaces.PutGame, tp presenter.PutPresenter) *PutInteractor {
	mustNotNil("PutInteractor", map[string]any{"g": g, "tp": tp})
	return &PutInteractor{GameBase: GameBase[interfaces.PutGame]{Game: g}, tp: tp}
}

// Reset マッチ初期化
func (ti *PutInteractor) Reset() string {
	ti.Game.Reset()
	ti.runCpuTurns()
	return ti.tp.Output(ti.Game, nil)
}

// ResetWithConfig 設定を変更してマッチ初期化
func (ti *PutInteractor) ResetWithConfig(cfg domain.PutConfig) string {
	return resetWithValidatedConfig(ti.Game, ti.tp, cfg, ti.Game.SetConfig, ti.Reset)
}

// Play カードをプレイ
func (ti *PutInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(ti.Game, ti.tp); blocked {
		return out
	}
	if err := ti.Game.PlayerPlay(cardIndex); err != nil {
		return ti.tp.Output(ti.Game, err)
	}
	ti.runCpuTurns()
	return ti.tp.Output(ti.Game, nil)
}

// Put Put を宣言 (または再引き上げ) する
func (ti *PutInteractor) Put() string {
	if out, blocked := guardGameEnd(ti.Game, ti.tp); blocked {
		return out
	}
	if err := ti.Game.DeclarePut(); err != nil {
		return ti.tp.Output(ti.Game, err)
	}
	ti.runCpuTurns()
	return ti.tp.Output(ti.Game, nil)
}

// Respond Put 宣言に応答する
func (ti *PutInteractor) Respond(accept bool) string {
	if out, blocked := guardGameEnd(ti.Game, ti.tp); blocked {
		return out
	}
	if err := ti.Game.RespondPut(accept); err != nil {
		return ti.tp.Output(ti.Game, err)
	}
	ti.runCpuTurns()
	return ti.tp.Output(ti.Game, nil)
}

// Next バサ / マノ終了から次へ進む
func (ti *PutInteractor) Next() string {
	ti.Game.Next()
	if out, blocked := guardGameEnd(ti.Game, ti.tp); blocked {
		return out
	}
	ti.runCpuTurns()
	return ti.tp.Output(ti.Game, nil)
}

// GetConfig 現在の設定を取得
func (ti *PutInteractor) GetConfig() domain.PutConfig {
	return ti.Game.GetConfig()
}

// Hint ヒント取得
func (ti *PutInteractor) Hint() string {
	return ti.tp.HintOutput(ti.Game)
}

// ActionLog 棋譜を出力する
func (ti *PutInteractor) ActionLog() string {
	return ti.tp.ActionLogOutput(ti.Game)
}

// runCpuTurns マッチが終わるか人間の手番 (プレイ/応答)、またはバサ/マノ終了に
// なるまで CPU のアクションを実行する。Put は応答 (Respond) フェーズを持つため
// 共通の runCpuTurnsLoop ではなく専用ループを用いる。
func (ti *PutInteractor) runCpuTurns() {
	g := ti.Game
	for i := 0; i < MaxCpuIterations; i++ {
		if g.GetGameEndFlag() {
			return
		}
		switch g.GetPhase() {
		case domain.PutPhasePlay, domain.PutPhaseRespond:
			if g.IsHumanTurn() {
				return
			}
			g.CpuStep()
		default:
			// TrickEnd / HandEnd / GameEnd は人間の操作待ちで一時停止
			return
		}
	}
}

// RestorePutInteractor deserialises JSON into a PutInteractor.
func RestorePutInteractor(data []byte, tp presenter.PutPresenter) (*PutInteractor, error) {
	return restoreAndBuild[domain.Put](data, func(g *domain.Put) *PutInteractor {
		return &PutInteractor{GameBase: GameBase[interfaces.PutGame]{Game: g}, tp: tp}
	})
}
