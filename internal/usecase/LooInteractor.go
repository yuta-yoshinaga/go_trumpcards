//go:build !js || !wasm || extra3

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// LooInteractorIF はルー (Loo) のインタラクターインタフェース。
type LooInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化 (新規ゲーム)
	Reset() string
	// ResetWithConfig 設定を変更してゲームを初期化
	ResetWithConfig(cfg domain.LooConfig) string
	// Decide 参加 (play=true) / 降り (play=false) を決める
	Decide(play bool) string
	// Play 手札を出す
	Play(cardIndex int) string
	// NextRound ディールを精算して次のディールへ進む
	NextRound() string
	// GetConfig 現在の設定を返す
	GetConfig() domain.LooConfig
	// Hint ヒントを出力する
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// LooInteractor はルーインタラクター。
type LooInteractor struct {
	GameBase[interfaces.LooGame]
	cp presenter.LooPresenter
}

// NewLooInteractor コンストラクタ。
func NewLooInteractor(lg interfaces.LooGame, cp presenter.LooPresenter) *LooInteractor {
	mustNotNil("LooInteractor", map[string]any{"lg": lg, "cp": cp})
	return &LooInteractor{GameBase: GameBase[interfaces.LooGame]{Game: lg}, cp: cp}
}

// Reset ゲーム初期化 (新規ゲーム)。
func (li *LooInteractor) Reset() string {
	li.Game.Reset()
	li.advance()
	return li.cp.Output(li.Game, nil)
}

// ResetWithConfig 設定を変更してゲームを初期化。
func (li *LooInteractor) ResetWithConfig(cfg domain.LooConfig) string {
	return resetWithValidatedConfig(li.Game, li.cp, cfg, li.Game.SetConfig, li.Reset)
}

// Decide 参加 (play/pass) を決める。
func (li *LooInteractor) Decide(play bool) string {
	if out, blocked := guardGameEnd(li.Game, li.cp); blocked {
		return out
	}
	if err := li.Game.PlayerDecide(play); err != nil {
		return li.cp.Output(li.Game, err)
	}
	li.advance()
	return li.cp.Output(li.Game, nil)
}

// Play 手札を出す。
func (li *LooInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(li.Game, li.cp); blocked {
		return out
	}
	if err := li.Game.PlayerPlay(cardIndex); err != nil {
		return li.cp.Output(li.Game, err)
	}
	// 人間が最後のカードを出してトリック完了した場合、即座に解決する。トリックが最終
	// トリックでなければ次トリックへ進める (ResolveTrick は TrickEnd のまま残すため、
	// CPU 経路と同様に NextTrick を呼ばないと play に戻れず停止してしまう)。
	if li.Game.GetPhase() == domain.LooPhaseTrickEnd {
		li.Game.ResolveTrick()
		if li.Game.GetPhase() == domain.LooPhaseTrickEnd {
			li.Game.NextTrick()
		}
	}
	li.advance()
	return li.cp.Output(li.Game, nil)
}

// NextRound ディールを精算して次のディールへ進む。
func (li *LooInteractor) NextRound() string {
	li.Game.ScoreRound()
	if out, blocked := guardGameEnd(li.Game, li.cp); blocked {
		return out
	}
	li.Game.NextRound()
	li.advance()
	return li.cp.Output(li.Game, nil)
}

// GetConfig 現在の設定を返す。
func (li *LooInteractor) GetConfig() domain.LooConfig { return li.Game.GetConfig() }

// Hint ヒントを出力する。
func (li *LooInteractor) Hint() string { return li.cp.HintOutput(li.Game) }

// ActionLog 棋譜を出力する。
func (li *LooInteractor) ActionLog() string { return li.cp.ActionLogOutput(li.Game) }

// looMaxCpuIterations は advance の防御的な反復上限。
const looMaxCpuIterations = 1000

// advance はゲーム終了・人間の意思決定待ち・トリック終了/ディール終了のいずれかに
// 到達するまで CPU ステップを回す。decide・play の両フェーズを CpuPlay 経由で進め、
// トリックが揃ったら解決して次トリックへ移る。トリック終了・ディール終了では人間の
// 操作 (NextRound) を待つため自動進行しない。
func (li *LooInteractor) advance() {
	for i := 0; i < looMaxCpuIterations; i++ {
		if li.Game.GetGameEndFlag() {
			return
		}
		phase := li.Game.GetPhase()
		switch phase {
		case domain.LooPhaseDecide, domain.LooPhasePlay:
			if li.Game.IsHumanTurn() {
				return
			}
			li.Game.CpuPlay()
			if li.Game.GetPhase() == domain.LooPhaseTrickEnd {
				li.Game.ResolveTrick()
				if li.Game.GetPhase() == domain.LooPhaseRoundEnd {
					return
				}
				li.Game.NextTrick()
			}
		default:
			// TrickEnd / RoundEnd では人間の操作を待つ。
			return
		}
	}
}

// RestoreLooInteractor deserialises JSON into a LooInteractor.
func RestoreLooInteractor(data []byte, cp presenter.LooPresenter) (*LooInteractor, error) {
	return restoreAndBuild[domain.Loo](data, func(g *domain.Loo) *LooInteractor {
		return &LooInteractor{GameBase: GameBase[interfaces.LooGame]{Game: g}, cp: cp}
	})
}
