package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// AllFoursInteractorIF All Fours インタラクターインタフェース
type AllFoursInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.AllFoursConfig) string
	// Beg 非親が stand / beg を選ぶ (beg=true で beg)
	Beg(beg bool) string
	// RespondBeg 親が beg に応答する (run=true で run the cards)
	RespondBeg(run bool) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextTrick 次のトリックへ進む (現在のトリックを解決してから進める)
	NextTrick() string
	// NextRound ラウンドをスコアリングして次のディールへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.AllFoursConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// AllFoursInteractor All Fours インタラクタークラス
type AllFoursInteractor struct {
	GameBase[interfaces.AllFoursGame]
	pp presenter.AllFoursPresenter
}

// NewAllFoursInteractor コンストラクタ
func NewAllFoursInteractor(g interfaces.AllFoursGame, pp presenter.AllFoursPresenter) *AllFoursInteractor {
	mustNotNil("AllFoursInteractor", map[string]any{"g": g, "pp": pp})
	return &AllFoursInteractor{GameBase: GameBase[interfaces.AllFoursGame]{Game: g}, pp: pp}
}

// allFoursTrickPhases AllFours のトリックフェーズ定数
func allFoursTrickPhases() trickPhases[domain.AllFoursPhase] {
	return trickPhases[domain.AllFoursPhase]{
		play:     domain.AllFoursPhasePlay,
		trickEnd: domain.AllFoursPhaseTrickEnd,
		roundEnd: domain.AllFoursPhaseRoundEnd,
		gameEnd:  domain.AllFoursPhaseGameEnd,
	}
}

// Reset ゲーム初期化
func (ai *AllFoursInteractor) Reset() string {
	ai.Game.Reset()
	ai.runCpuTurns()
	return ai.pp.Output(ai.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ai *AllFoursInteractor) ResetWithConfig(cfg domain.AllFoursConfig) string {
	return resetWithValidatedConfig(ai.Game, ai.pp, cfg, ai.Game.SetConfig, ai.Reset)
}

// Beg 非親が stand / beg を選ぶ
func (ai *AllFoursInteractor) Beg(beg bool) string {
	if out, blocked := guardGameEnd(ai.Game, ai.pp); blocked {
		return out
	}
	if err := ai.Game.PlayerBeg(beg); err != nil {
		return ai.pp.Output(ai.Game, err)
	}
	ai.runCpuTurns()
	return ai.pp.Output(ai.Game, nil)
}

// RespondBeg 親が beg に応答する
func (ai *AllFoursInteractor) RespondBeg(run bool) string {
	if out, blocked := guardGameEnd(ai.Game, ai.pp); blocked {
		return out
	}
	if err := ai.Game.PlayerRespondBeg(run); err != nil {
		return ai.pp.Output(ai.Game, err)
	}
	ai.runCpuTurns()
	return ai.pp.Output(ai.Game, nil)
}

// Play カードをプレイ
func (ai *AllFoursInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(ai.Game, ai.pp); blocked {
		return out
	}
	if err := ai.Game.PlayerPlay(cardIndex); err != nil {
		return ai.pp.Output(ai.Game, err)
	}
	ai.runCpuTurns()
	return ai.pp.Output(ai.Game, nil)
}

// NextTrick 現在のトリックを解決してから次のトリックへ進める。
func (ai *AllFoursInteractor) NextTrick() string {
	if ai.Game.GetPhase() == domain.AllFoursPhaseTrickEnd {
		ai.Game.ResolveTrick()
	}
	if ai.Game.GetPhase() == domain.AllFoursPhaseRoundEnd {
		// 最終トリック: ラウンド終了へ。スコアリングは NextRound で行う。
		return ai.pp.Output(ai.Game, nil)
	}
	if ai.Game.GetPhase() == domain.AllFoursPhaseTrickEnd {
		ai.Game.NextTrick()
	}
	ai.runCpuTurns()
	return ai.pp.Output(ai.Game, nil)
}

// NextRound ラウンドをスコアリングして次のディールへ進む
func (ai *AllFoursInteractor) NextRound() string {
	ai.Game.ScoreRound()
	if out, blocked := guardGameEnd(ai.Game, ai.pp); blocked {
		return out
	}
	ai.Game.NextRound()
	ai.runCpuTurns()
	return ai.pp.Output(ai.Game, nil)
}

// GetConfig 現在の設定を取得
func (ai *AllFoursInteractor) GetConfig() domain.AllFoursConfig {
	return ai.Game.GetConfig()
}

// Hint ヒント取得
func (ai *AllFoursInteractor) Hint() string {
	return ai.pp.HintOutput(ai.Game)
}

// ActionLog 棋譜を出力する
func (ai *AllFoursInteractor) ActionLog() string {
	return ai.pp.ActionLogOutput(ai.Game)
}

// runCpuTurns CPUの手番 (beg応答 / プレイ) を人間の手番もしくは
// トリック/ラウンド終了になるまで自動実行する。
//
// AllFours では非親 (人間) が最初にリードし、各トリックを CPU が完了させる場合は
// runCpuTurnsLoop が解決する。CPU がトリックに勝って次をリードする場合も同ループが
// 連続して処理する。人間がトリックを完了した場合 (phase=trickEnd) はループが停止し、
// プレイヤーが NextTrick を押すまで待機する。
func (ai *AllFoursInteractor) runCpuTurns() {
	runCpuTurnsLoop(ai.Game, allFoursTrickPhases())
}

// RestoreAllFoursInteractor deserialises JSON into an AllFoursInteractor.
func RestoreAllFoursInteractor(data []byte, pp presenter.AllFoursPresenter) (*AllFoursInteractor, error) {
	return restoreAndBuild[domain.AllFours](data, func(g *domain.AllFours) *AllFoursInteractor {
		return &AllFoursInteractor{GameBase: GameBase[interfaces.AllFoursGame]{Game: g}, pp: pp}
	})
}
