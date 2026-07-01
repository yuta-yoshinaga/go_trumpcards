//go:build !js || !wasm || extra

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// CinchInteractorIF はチンチ (Cinch) のインタラクターインタフェース。
type CinchInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化 (新規ゲーム)
	Reset() string
	// ResetWithConfig 設定を変更してゲームを初期化
	ResetWithConfig(cfg domain.CinchConfig) string
	// Bid ビッドする (0=pass, 1..CinchMaxBid)
	Bid(bid int) string
	// NameTrump ビッド勝者が切り札スートを宣言する
	NameTrump(suit int) string
	// Play 手札を出す
	Play(cardIndex int) string
	// NextRound ディールをスコアリングして次のディールへ進む
	NextRound() string
	// GetConfig 現在の設定を返す
	GetConfig() domain.CinchConfig
	// Hint ヒントを出力する
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// CinchInteractor はチンチインタラクター。
type CinchInteractor struct {
	GameBase[interfaces.CinchGame]
	cp presenter.CinchPresenter
}

// NewCinchInteractor コンストラクタ。
func NewCinchInteractor(cg interfaces.CinchGame, cp presenter.CinchPresenter) *CinchInteractor {
	mustNotNil("CinchInteractor", map[string]any{"cg": cg, "cp": cp})
	return &CinchInteractor{GameBase: GameBase[interfaces.CinchGame]{Game: cg}, cp: cp}
}

// Reset ゲーム初期化 (新規ゲーム)。
func (ci *CinchInteractor) Reset() string {
	ci.Game.Reset()
	ci.advance()
	return ci.cp.Output(ci.Game, nil)
}

// ResetWithConfig 設定を変更してゲームを初期化。
func (ci *CinchInteractor) ResetWithConfig(cfg domain.CinchConfig) string {
	return resetWithValidatedConfig(ci.Game, ci.cp, cfg, ci.Game.SetConfig, ci.Reset)
}

// Bid ビッドする。
func (ci *CinchInteractor) Bid(bid int) string {
	if out, blocked := guardGameEnd(ci.Game, ci.cp); blocked {
		return out
	}
	if err := ci.Game.PlayerBid(bid); err != nil {
		return ci.cp.Output(ci.Game, err)
	}
	ci.advance()
	return ci.cp.Output(ci.Game, nil)
}

// NameTrump ビッド勝者が切り札スートを宣言する。
func (ci *CinchInteractor) NameTrump(suit int) string {
	if out, blocked := guardGameEnd(ci.Game, ci.cp); blocked {
		return out
	}
	if err := ci.Game.NameTrump(suit); err != nil {
		return ci.cp.Output(ci.Game, err)
	}
	ci.advance()
	return ci.cp.Output(ci.Game, nil)
}

// Play 手札を出す。
func (ci *CinchInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(ci.Game, ci.cp); blocked {
		return out
	}
	if err := ci.Game.PlayerPlay(cardIndex); err != nil {
		return ci.cp.Output(ci.Game, err)
	}
	// 人間が最後のカードを出してトリック完了した場合、即座に解決する。
	if ci.Game.GetPhase() == domain.CinchPhaseTrickEnd {
		ci.Game.ResolveTrick()
	}
	ci.advance()
	return ci.cp.Output(ci.Game, nil)
}

// NextRound ディールをスコアリングして次のディールへ進む。
func (ci *CinchInteractor) NextRound() string {
	ci.Game.ScoreRound()
	if out, blocked := guardGameEnd(ci.Game, ci.cp); blocked {
		return out
	}
	ci.Game.NextRound()
	ci.advance()
	return ci.cp.Output(ci.Game, nil)
}

// GetConfig 現在の設定を返す。
func (ci *CinchInteractor) GetConfig() domain.CinchConfig { return ci.Game.GetConfig() }

// Hint ヒントを出力する。
func (ci *CinchInteractor) Hint() string { return ci.cp.HintOutput(ci.Game) }

// ActionLog 棋譜を出力する。
func (ci *CinchInteractor) ActionLog() string { return ci.cp.ActionLogOutput(ci.Game) }

// cinchMaxCpuIterations は advance の防御的な反復上限。
const cinchMaxCpuIterations = 1000

// advance はゲーム終了・人間の意思決定待ち・トリック終了/ラウンド終了のいずれかに
// 到達するまで CPU ステップを回す。ビッド・切り札宣言・プレイの全フェーズを
// CpuPlay 経由で進め、トリックが揃ったら解決して次トリックへ移る。トリック終了・
// ラウンド終了・ゲーム終了では人間の操作 (NextRound) を待つため自動進行しない。
func (ci *CinchInteractor) advance() {
	for i := 0; i < cinchMaxCpuIterations; i++ {
		if ci.Game.GetGameEndFlag() {
			return
		}
		phase := ci.Game.GetPhase()
		switch phase {
		case domain.CinchPhaseBid, domain.CinchPhaseNameTrump, domain.CinchPhasePlay:
			if ci.Game.IsHumanTurn() {
				return
			}
			ci.Game.CpuPlay()
			if ci.Game.GetPhase() == domain.CinchPhaseTrickEnd {
				ci.Game.ResolveTrick()
				if ci.Game.GetPhase() == domain.CinchPhaseRoundEnd {
					return
				}
				ci.Game.NextTrick()
			}
		default:
			// TrickEnd / RoundEnd / GameEnd では人間の操作を待つ。
			return
		}
	}
}

// RestoreCinchInteractor deserialises JSON into a CinchInteractor.
func RestoreCinchInteractor(data []byte, cp presenter.CinchPresenter) (*CinchInteractor, error) {
	return restoreAndBuild[domain.Cinch](data, func(g *domain.Cinch) *CinchInteractor {
		return &CinchInteractor{GameBase: GameBase[interfaces.CinchGame]{Game: g}, cp: cp}
	})
}
