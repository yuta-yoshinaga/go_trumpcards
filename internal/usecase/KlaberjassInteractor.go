//go:build !js || !wasm || extra3

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// KlaberjassInteractorIF クラバーヤス (Klaberjass) のインタラクターインタフェース
type KlaberjassInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.KlaberjassConfig) string
	// AcceptTrump 表向きカードのスートを切札にする
	AcceptTrump() string
	// CallTrump 好きなスートを切札に指名する
	CallTrump(suit int) string
	// Pass ビッドを見送る
	Pass() string
	// Schmeiss この配りを流すことを提案する
	Schmeiss() string
	// AnswerSchmeiss 投げの提案に答える
	AnswerSchmeiss(accept bool) string
	// PlayCard 手札を1枚出す
	PlayCard(idx int) string
	// NextDeal 次のディールへ進む
	NextDeal() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.KlaberjassConfig
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// KlaberjassInteractor クラバーヤス (Klaberjass) のインタラクタークラス
type KlaberjassInteractor struct {
	GameBase[interfaces.KlaberjassGame]
	gp presenter.KlaberjassPresenter
}

// NewKlaberjassInteractor コンストラクタ
func NewKlaberjassInteractor(g interfaces.KlaberjassGame, gp presenter.KlaberjassPresenter) *KlaberjassInteractor {
	mustNotNil("KlaberjassInteractor", map[string]any{"g": g, "gp": gp})
	return &KlaberjassInteractor{GameBase: GameBase[interfaces.KlaberjassGame]{Game: g}, gp: gp}
}

// Reset ゲーム初期化
func (ki *KlaberjassInteractor) Reset() string {
	ki.Game.Reset()
	ki.runCpuTurns()
	return ki.gp.Output(ki.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ki *KlaberjassInteractor) ResetWithConfig(cfg domain.KlaberjassConfig) string {
	return resetWithValidatedConfig(ki.Game, ki.gp, cfg, ki.Game.SetConfig, ki.Reset)
}

// AcceptTrump 表向きカードのスートを切札にする
func (ki *KlaberjassInteractor) AcceptTrump() string {
	return ki.act(func() error { return ki.Game.AcceptTrump(ki.Game.GetBidPlayerIdx()) })
}

// CallTrump 好きなスートを切札に指名する
func (ki *KlaberjassInteractor) CallTrump(suit int) string {
	return ki.act(func() error { return ki.Game.CallTrump(ki.Game.GetBidPlayerIdx(), suit) })
}

// Pass ビッドを見送る
func (ki *KlaberjassInteractor) Pass() string {
	return ki.act(func() error { return ki.Game.Pass(ki.Game.GetBidPlayerIdx()) })
}

// Schmeiss この配りを流すことを提案する
func (ki *KlaberjassInteractor) Schmeiss() string {
	return ki.act(func() error { return ki.Game.Schmeiss(ki.Game.GetBidPlayerIdx()) })
}

// AnswerSchmeiss 投げの提案に答える
func (ki *KlaberjassInteractor) AnswerSchmeiss(accept bool) string {
	return ki.act(func() error { return ki.Game.AnswerSchmeiss(ki.Game.GetBidPlayerIdx(), accept) })
}

// PlayCard 手札を1枚出す
func (ki *KlaberjassInteractor) PlayCard(idx int) string {
	return ki.act(func() error { return ki.Game.PlayCard(ki.Game.GetCurrentPlayerIdx(), idx) })
}

// act 人間アクションの共通処理 (ガード → 実行 → CPU 進行)
func (ki *KlaberjassInteractor) act(action func() error) string {
	if out, blocked := guardNotPlayable(ki.Game, ki.gp); blocked {
		return out
	}
	if err := action(); err != nil {
		return ki.gp.Output(ki.Game, err)
	}
	ki.runCpuTurns()
	return ki.gp.Output(ki.Game, nil)
}

// NextDeal 次のディールへ進む
func (ki *KlaberjassInteractor) NextDeal() string {
	if out, blocked := guardGameEnd(ki.Game, ki.gp); blocked {
		return out
	}
	if err := ki.Game.NextDeal(); err != nil {
		return ki.gp.Output(ki.Game, err)
	}
	ki.runCpuTurns()
	return ki.gp.Output(ki.Game, nil)
}

// GetConfig 現在の設定を取得
func (ki *KlaberjassInteractor) GetConfig() domain.KlaberjassConfig {
	return ki.Game.GetConfig()
}

// ActionLog 棋譜を出力する
func (ki *KlaberjassInteractor) ActionLog() string {
	return ki.gp.ActionLogOutput(ki.Game)
}

// klaberjassMaxCpuSteps bounds runCpuTurns so a malformed state can never spin
// the CPU loop forever (defensive — normal play always reaches a human turn,
// the settlement, or game end well within this limit).
const klaberjassMaxCpuSteps = 1000

// runCpuTurns CPUターンを連続実行する
func (ki *KlaberjassInteractor) runCpuTurns() {
	for step := 0; step < klaberjassMaxCpuSteps && !ki.Game.GetGameEndFlag(); step++ {
		phase := ki.Game.GetPhase()
		if phase == domain.KlaberjassPhaseHandEnd || phase == domain.KlaberjassPhaseGameEnd {
			break
		}
		if ki.Game.IsHumanTurn() {
			break
		}
		ki.Game.CpuPlay()
	}
}

// RestoreKlaberjassInteractor deserialises JSON into a KlaberjassInteractor.
func RestoreKlaberjassInteractor(data []byte, gp presenter.KlaberjassPresenter) (*KlaberjassInteractor, error) {
	return restoreAndBuild[domain.Klaberjass](data, func(g *domain.Klaberjass) *KlaberjassInteractor {
		return &KlaberjassInteractor{GameBase: GameBase[interfaces.KlaberjassGame]{Game: g}, gp: gp}
	})
}
