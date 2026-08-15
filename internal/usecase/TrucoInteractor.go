package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// TrucoInteractorIF トゥルコインタラクターインタフェース
type TrucoInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset マッチ初期化
	Reset() string
	// ResetWithConfig 設定を変更してマッチ初期化
	ResetWithConfig(cfg domain.TrucoConfig) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// Truco Truco を宣言 (または再引き上げ) する
	Truco() string
	// Respond Truco 宣言に応答する (true=受諾 / false=拒否)
	Respond(accept bool) string
	// Next バサ / マノ終了から次へ進む
	Next() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.TrucoConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// TrucoInteractor トゥルコインタラクタークラス
type TrucoInteractor struct {
	GameBase[interfaces.TrucoGame]
	tp presenter.TrucoPresenter
}

// NewTrucoInteractor コンストラクタ
func NewTrucoInteractor(g interfaces.TrucoGame, tp presenter.TrucoPresenter) *TrucoInteractor {
	mustNotNil("TrucoInteractor", map[string]any{"g": g, "tp": tp})
	return &TrucoInteractor{GameBase: GameBase[interfaces.TrucoGame]{Game: g}, tp: tp}
}

// Reset マッチ初期化
func (ti *TrucoInteractor) Reset() string {
	ti.Game.Reset()
	ti.runCpuTurns()
	return ti.tp.Output(ti.Game, nil)
}

// ResetWithConfig 設定を変更してマッチ初期化
func (ti *TrucoInteractor) ResetWithConfig(cfg domain.TrucoConfig) string {
	return resetWithValidatedConfig(ti.Game, ti.tp, cfg, ti.Game.SetConfig, ti.Reset)
}

// Play カードをプレイ
func (ti *TrucoInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(ti.Game, ti.tp); blocked {
		return out
	}
	if err := ti.Game.PlayerPlay(cardIndex); err != nil {
		return ti.tp.Output(ti.Game, err)
	}
	ti.runCpuTurns()
	return ti.tp.Output(ti.Game, nil)
}

// Truco Truco を宣言 (または再引き上げ) する
func (ti *TrucoInteractor) Truco() string {
	if out, blocked := guardGameEnd(ti.Game, ti.tp); blocked {
		return out
	}
	if err := ti.Game.DeclareTruco(); err != nil {
		return ti.tp.Output(ti.Game, err)
	}
	ti.runCpuTurns()
	return ti.tp.Output(ti.Game, nil)
}

// Respond Truco 宣言に応答する
func (ti *TrucoInteractor) Respond(accept bool) string {
	if out, blocked := guardGameEnd(ti.Game, ti.tp); blocked {
		return out
	}
	if err := ti.Game.RespondTruco(accept); err != nil {
		return ti.tp.Output(ti.Game, err)
	}
	ti.runCpuTurns()
	return ti.tp.Output(ti.Game, nil)
}

// Next バサ / マノ終了から次へ進む
func (ti *TrucoInteractor) Next() string {
	ti.Game.Next()
	if out, blocked := guardGameEnd(ti.Game, ti.tp); blocked {
		return out
	}
	ti.runCpuTurns()
	return ti.tp.Output(ti.Game, nil)
}

// GetConfig 現在の設定を取得
func (ti *TrucoInteractor) GetConfig() domain.TrucoConfig {
	return ti.Game.GetConfig()
}

// Hint ヒント取得
func (ti *TrucoInteractor) Hint() string {
	return ti.tp.HintOutput(ti.Game)
}

// ActionLog 棋譜を出力する
func (ti *TrucoInteractor) ActionLog() string {
	return ti.tp.ActionLogOutput(ti.Game)
}

// runCpuTurns マッチが終わるか人間の手番 (プレイ/応答)、またはバサ/マノ終了に
// なるまで CPU のアクションを実行する。Truco は応答 (Respond) フェーズを持つため
// 共通の runCpuTurnsLoop ではなく専用ループを用いる。
func (ti *TrucoInteractor) runCpuTurns() {
	g := ti.Game
	for i := 0; i < MaxCpuIterations; i++ {
		if g.GetGameEndFlag() {
			return
		}
		switch g.GetPhase() {
		case domain.TrucoPhasePlay, domain.TrucoPhaseRespond:
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

// RestoreTrucoInteractor deserialises JSON into a TrucoInteractor.
func RestoreTrucoInteractor(data []byte, tp presenter.TrucoPresenter) (*TrucoInteractor, error) {
	return restoreAndBuild[domain.Truco](data, func(g *domain.Truco) *TrucoInteractor {
		return &TrucoInteractor{GameBase: GameBase[interfaces.TrucoGame]{Game: g}, tp: tp}
	})
}
