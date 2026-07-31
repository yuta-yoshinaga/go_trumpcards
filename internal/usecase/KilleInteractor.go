//go:build !js || !wasm || extra3

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// KilleInteractorIF キッレ (Kille) のインタラクターインタフェース
type KilleInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.KilleConfig) string
	// Exchange 隣 (ディーラーなら山) と交換を仕掛ける
	Exchange() string
	// Satisfied 交換しないと宣言する
	Satisfied() string
	// Reenter 買い戻して次のラウンドへ進む
	Reenter() string
	// NextRound 買い戻さずに次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.KilleConfig
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// KilleInteractor キッレ (Kille) のインタラクタークラス
type KilleInteractor struct {
	GameBase[interfaces.KilleGame]
	gp presenter.KillePresenter
}

// NewKilleInteractor コンストラクタ
func NewKilleInteractor(g interfaces.KilleGame, gp presenter.KillePresenter) *KilleInteractor {
	mustNotNil("KilleInteractor", map[string]any{"g": g, "gp": gp})
	return &KilleInteractor{GameBase: GameBase[interfaces.KilleGame]{Game: g}, gp: gp}
}

// Reset ゲーム初期化
func (ki *KilleInteractor) Reset() string {
	ki.Game.Reset()
	ki.runCpuTurns()
	return ki.gp.Output(ki.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ki *KilleInteractor) ResetWithConfig(cfg domain.KilleConfig) string {
	return resetWithValidatedConfig(ki.Game, ki.gp, cfg, ki.Game.SetConfig, ki.Reset)
}

// Exchange 隣 (ディーラーなら山) と交換を仕掛ける
func (ki *KilleInteractor) Exchange() string {
	return ki.act(ki.Game.Exchange)
}

// Satisfied 交換しないと宣言する
func (ki *KilleInteractor) Satisfied() string {
	return ki.act(ki.Game.Satisfied)
}

// act 人間アクションの共通処理 (ガード → 実行 → CPU 進行)
func (ki *KilleInteractor) act(action func(int) error) string {
	if out, blocked := guardNotPlayable(ki.Game, ki.gp); blocked {
		return out
	}
	if err := action(ki.Game.GetCurrentPlayerIdx()); err != nil {
		return ki.gp.Output(ki.Game, err)
	}
	ki.runCpuTurns()
	return ki.gp.Output(ki.Game, nil)
}

// Reenter 買い戻して次のラウンドへ進む
func (ki *KilleInteractor) Reenter() string {
	if out, blocked := guardGameEnd(ki.Game, ki.gp); blocked {
		return out
	}
	if err := ki.Game.Reenter(ki.humanIdx()); err != nil {
		return ki.gp.Output(ki.Game, err)
	}
	return ki.advanceRound()
}

// NextRound 次のラウンドへ進む
func (ki *KilleInteractor) NextRound() string {
	if out, blocked := guardGameEnd(ki.Game, ki.gp); blocked {
		return out
	}
	return ki.advanceRound()
}

// advanceRound は CPU の買い戻しを解決してから次のラウンドを配る。
//
// **買い戻しはラウンドを配る前にしか出来ない。**NextRound が買い戻さなかった
// 脱落者を退場させてしまうので、CPU の判断はその手前で済ませる必要がある。
func (ki *KilleInteractor) advanceRound() string {
	for i := range ki.Game.GetPlayers() {
		p := ki.Game.GetPlayer(i)
		if p == nil || p.GetIsHuman() {
			continue
		}
		if ki.Game.KilleCpuReenterDecide(i) {
			_ = ki.Game.Reenter(i)
		}
	}
	if err := ki.Game.NextRound(); err != nil {
		return ki.gp.Output(ki.Game, err)
	}
	ki.runCpuTurns()
	return ki.gp.Output(ki.Game, nil)
}

// humanIdx は人間の席を返す (見つからなければ 0)。
func (ki *KilleInteractor) humanIdx() int {
	for i := range ki.Game.GetPlayers() {
		if p := ki.Game.GetPlayer(i); p != nil && p.GetIsHuman() {
			return i
		}
	}
	return 0
}

// GetConfig 現在の設定を取得
func (ki *KilleInteractor) GetConfig() domain.KilleConfig {
	return ki.Game.GetConfig()
}

// ActionLog 棋譜を出力する
func (ki *KilleInteractor) ActionLog() string {
	return ki.gp.ActionLogOutput(ki.Game)
}

// killeMaxCpuSteps bounds runCpuTurns so a malformed state can never spin the
// CPU loop forever (defensive — normal play always reaches a human turn, the
// showdown, or game end well within this limit).
const killeMaxCpuSteps = 1000

// runCpuTurns CPUターンを連続実行する
func (ki *KilleInteractor) runCpuTurns() {
	for step := 0; step < killeMaxCpuSteps && !ki.Game.GetGameEndFlag(); step++ {
		if ki.Game.GetPhase() != domain.KillePhaseExchange {
			break
		}
		if ki.Game.IsHumanTurn() {
			break
		}
		ki.Game.CpuPlay()
	}
}

// RestoreKilleInteractor deserialises JSON into a KilleInteractor.
func RestoreKilleInteractor(data []byte, gp presenter.KillePresenter) (*KilleInteractor, error) {
	return restoreAndBuild[domain.Kille](data, func(g *domain.Kille) *KilleInteractor {
		return &KilleInteractor{GameBase: GameBase[interfaces.KilleGame]{Game: g}, gp: gp}
	})
}
