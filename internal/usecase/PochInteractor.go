//go:build !js || !wasm || extra3

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// PochInteractorIF ポッホインタラクターインタフェース
type PochInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.PochConfig) string
	// Bet pochen で1単位賭ける
	Bet() string
	// Fold pochen で降りる
	Fold() string
	// Play ストップスで手札1枚を出す
	Play(handIdx int) string
	// NextDeal 次のディールへ進む
	NextDeal() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.PochConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// PochInteractor ポッホインタラクタークラス
type PochInteractor struct {
	GameBase[interfaces.PochGame]
	cp presenter.PochPresenter
}

// NewPochInteractor コンストラクタ
func NewPochInteractor(c interfaces.PochGame, cp presenter.PochPresenter) *PochInteractor {
	mustNotNil("PochInteractor", map[string]any{"c": c, "cp": cp})
	return &PochInteractor{GameBase: GameBase[interfaces.PochGame]{Game: c}, cp: cp}
}

// pochHumanIdx 人間プレイヤーの座席。
const pochHumanIdx = 0

// pochCpuTurnCap は CPU ループの安全弁。
const pochCpuTurnCap = 2000

// Reset ゲーム初期化
func (pi *PochInteractor) Reset() string {
	pi.Game.Reset()
	pi.runCpuTurns()
	return pi.cp.Output(pi.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (pi *PochInteractor) ResetWithConfig(cfg domain.PochConfig) string {
	return resetWithValidatedConfig(pi.Game, pi.cp, cfg, pi.Game.SetConfig, pi.Reset)
}

// Bet pochen で1単位賭ける
func (pi *PochInteractor) Bet() string {
	return pi.apply(func() error { return pi.Game.Bet(pochHumanIdx) })
}

// Fold pochen で降りる
func (pi *PochInteractor) Fold() string {
	return pi.apply(func() error { return pi.Game.Fold(pochHumanIdx) })
}

// Play ストップスで手札1枚を出す
func (pi *PochInteractor) Play(handIdx int) string {
	return pi.apply(func() error { return pi.Game.Play(pochHumanIdx, handIdx) })
}

// apply は「決着していたら何もしない → 実行 → 失敗なら伝える → CPU を回す」を
// 共通化する。
func (pi *PochInteractor) apply(fn func() error) string {
	if pi.Game.GetGameEndFlag() {
		return pi.cp.Output(pi.Game, nil)
	}
	if err := fn(); err != nil {
		return pi.cp.Output(pi.Game, err)
	}
	pi.runCpuTurns()
	return pi.cp.Output(pi.Game, nil)
}

// NextDeal 次のディールへ進む
func (pi *PochInteractor) NextDeal() string {
	if err := pi.Game.NextDeal(); err != nil {
		return pi.cp.Output(pi.Game, err)
	}
	pi.runCpuTurns()
	return pi.cp.Output(pi.Game, nil)
}

// GetConfig 現在の設定を取得
func (pi *PochInteractor) GetConfig() domain.PochConfig { return pi.Game.GetConfig() }

// Hint ヒント取得
func (pi *PochInteractor) Hint() string { return pi.cp.HintOutput(pi.Game) }

// ActionLog 棋譜を出力する
func (pi *PochInteractor) ActionLog() string { return pi.cp.ActionLogOutput(pi.Game) }

// runCpuTurns 人間の手番に戻るか終局するまで CPU を進める。
//
// ディール終了では止める。勝手に配り直すと、9 プールの精算を読む間もなく画面が
// 変わってしまう。
//
// **人間が降りていても回し続ける。**pochen で降りた人にも手番は回ってこないが、
// 残りの CPU 同士で決着させないとストップスに進めない。
func (pi *PochInteractor) runCpuTurns() {
	for range pochCpuTurnCap {
		if pi.Game.GetGameEndFlag() {
			return
		}
		phase := pi.Game.GetPhase()
		if phase != domain.PochPhasePochen && phase != domain.PochPhaseStops {
			return
		}
		idx := pi.Game.GetCurrentPlayerIdx()
		if idx < 0 {
			return
		}
		if idx == pochHumanIdx && !pi.humanIsOut(phase) {
			return
		}
		action := pi.Game.PochCpuDecide(idx)
		var err error
		switch action.Type {
		case "bet":
			err = pi.Game.Bet(idx)
		case "fold":
			err = pi.Game.Fold(idx)
		default:
			if action.HandIdx < 0 {
				return
			}
			err = pi.Game.Play(idx, action.HandIdx)
		}
		if err != nil {
			return
		}
	}
}

// humanIsOut は人間の手番を飛ばしてよいかを返す。pochen で降りていれば飛ばす。
func (pi *PochInteractor) humanIsOut(phase domain.PochPhase) bool {
	if phase != domain.PochPhasePochen {
		return false
	}
	p := pi.Game.GetPlayer(pochHumanIdx)
	return p != nil && p.IsFolded()
}

// RestorePochInteractor deserialises JSON into a PochInteractor.
func RestorePochInteractor(data []byte, cp presenter.PochPresenter) (*PochInteractor, error) {
	return restoreAndBuild[domain.Poch](data, func(g *domain.Poch) *PochInteractor {
		return &PochInteractor{GameBase: GameBase[interfaces.PochGame]{Game: g}, cp: cp}
	})
}
