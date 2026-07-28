//go:build !js || !wasm || extra3

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// KoiKoiInteractorIF はこいこい (Koi-Koi) のインタラクターインタフェース。
type KoiKoiInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化 (新規ゲーム)
	Reset() string
	// ResetWithConfig 設定を変更してゲームを初期化
	ResetWithConfig(cfg domain.KoiKoiConfig) string
	// Play 手札を出す (fieldIdx で 2 枚一致時の捕獲対象を指定; 不要なら -1)
	Play(handIdx, fieldIdx int) string
	// Decide こいこい決断 (true=こいこい, false=勝負)
	Decide(koikoi bool) string
	// NextRound 次のラウンドを開始する
	NextRound() string
	// GetConfig 現在の設定を返す
	GetConfig() domain.KoiKoiConfig
	// Hint ヒントを出力する
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// KoiKoiInteractor はこいこいインタラクター。
type KoiKoiInteractor struct {
	GameBase[interfaces.KoiKoiGame]
	cp presenter.KoiKoiPresenter
}

// NewKoiKoiInteractor コンストラクタ。
func NewKoiKoiInteractor(kg interfaces.KoiKoiGame, cp presenter.KoiKoiPresenter) *KoiKoiInteractor {
	mustNotNil("KoiKoiInteractor", map[string]any{"kg": kg, "cp": cp})
	return &KoiKoiInteractor{GameBase: GameBase[interfaces.KoiKoiGame]{Game: kg}, cp: cp}
}

// Reset ゲーム初期化 (新規ゲーム)。
func (ki *KoiKoiInteractor) Reset() string {
	ki.Game.Reset()
	ki.advance()
	return ki.cp.Output(ki.Game, nil)
}

// ResetWithConfig 設定を変更してゲームを初期化。
func (ki *KoiKoiInteractor) ResetWithConfig(cfg domain.KoiKoiConfig) string {
	return resetWithValidatedConfig(ki.Game, ki.cp, cfg, ki.Game.SetConfig, ki.Reset)
}

// Play 手札を出す。
func (ki *KoiKoiInteractor) Play(handIdx, fieldIdx int) string {
	if out, blocked := guardNotPlayable(ki.Game, ki.cp); blocked {
		return out
	}
	if err := ki.Game.PlayerPlay(handIdx, fieldIdx); err != nil {
		return ki.cp.Output(ki.Game, err)
	}
	ki.advance()
	return ki.cp.Output(ki.Game, nil)
}

// Decide こいこい決断。
func (ki *KoiKoiInteractor) Decide(koikoi bool) string {
	if out, blocked := guardNotPlayable(ki.Game, ki.cp); blocked {
		return out
	}
	if err := ki.Game.PlayerDecide(koikoi); err != nil {
		return ki.cp.Output(ki.Game, err)
	}
	ki.advance()
	return ki.cp.Output(ki.Game, nil)
}

// NextRound 次のラウンドを開始する。
func (ki *KoiKoiInteractor) NextRound() string {
	ki.Game.NextRound()
	ki.advance()
	return ki.cp.Output(ki.Game, nil)
}

// GetConfig 現在の設定を返す。
func (ki *KoiKoiInteractor) GetConfig() domain.KoiKoiConfig { return ki.Game.GetConfig() }

// Hint ヒントを出力する。
func (ki *KoiKoiInteractor) Hint() string { return ki.cp.HintOutput(ki.Game) }

// ActionLog 棋譜を出力する。
func (ki *KoiKoiInteractor) ActionLog() string { return ki.cp.ActionLogOutput(ki.Game) }

// koikoiMaxCpuIterations は advance の防御的な反復上限。
const koikoiMaxCpuIterations = 1000

// advance はゲーム終了・ラウンド終了・人間の手番のいずれかに到達するまで CPU の
// プレイ/決断を回す。
func (ki *KoiKoiInteractor) advance() {
	for i := 0; i < koikoiMaxCpuIterations; i++ {
		if ki.Game.GetGameEndFlag() {
			return
		}
		switch ki.Game.GetPhase() {
		case domain.KoiKoiPhasePlay:
			if ki.Game.IsHumanTurn() {
				return
			}
			ki.Game.CpuPlay()
		case domain.KoiKoiPhaseKoiKoiDecision:
			if ki.Game.IsHumanTurn() {
				return
			}
			ki.Game.CpuDecide()
		default:
			// RoundEnd / GameEnd は人間の操作待ち。
			return
		}
	}
}

// RestoreKoiKoiInteractor deserialises JSON into a KoiKoiInteractor.
func RestoreKoiKoiInteractor(data []byte, cp presenter.KoiKoiPresenter) (*KoiKoiInteractor, error) {
	return restoreAndBuild[domain.KoiKoi](data, func(g *domain.KoiKoi) *KoiKoiInteractor {
		return &KoiKoiInteractor{GameBase: GameBase[interfaces.KoiKoiGame]{Game: g}, cp: cp}
	})
}
