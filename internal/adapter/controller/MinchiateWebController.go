//go:build !js || !wasm || solo

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// MinchiateWebInput ミンキアーテのWebインプット
type MinchiateWebInput struct {
	BaseWebInput
	// CardIndex プレイするカードのインデックス
	CardIndex *int `json:"cardIndex,omitempty"`
	// CardIndices スカルトで捨てるカードのインデックス (MinchiateSurplus 枚)
	CardIndices []int `json:"cardIndices,omitempty"`
	// Config ゲーム設定
	Config *MinchiateWebConfig `json:"config,omitempty"`
}

// MinchiateWebConfig ミンキアーテのWeb設定
type MinchiateWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	TargetRounds  *int `json:"targetRounds,omitempty"`
}

// MinchiateWebOutputPlayer ミンキアーテのWebアウトプットプレイヤー
type MinchiateWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
	TrickCount int              `json:"trickCount"`
	Team       int              `json:"team"`
	IsDealer   bool             `json:"isDealer"`
}

// MinchiateWebOutput ミンキアーテのWebアウトプット
type MinchiateWebOutput struct {
	Players          []*MinchiateWebOutputPlayer    `json:"players"`
	Phase            int                            `json:"phase"`
	RoundNumber      int                            `json:"roundNumber"`
	TrickNumber      int                            `json:"trickNumber"`
	CurrentPlayerIdx int                            `json:"currentPlayerIdx"`
	LeadPlayerIdx    int                            `json:"leadPlayerIdx"`
	DealerIdx        int                            `json:"dealerIdx"`
	ScartoCount      int                            `json:"scartoCount"`
	CurrentTrick     []*WebOutputTrickCard          `json:"currentTrick"`
	TeamScores       [2]int                         `json:"teamScores"`
	RoundTricks      [domain.MinchiatePlayerCnt]int `json:"roundTricks"`
	LastTrickWinner  int                            `json:"lastTrickWinner"`
	PlayableIndices  []int                          `json:"playableIndices"`
	GameEndFlag      bool                           `json:"gameEndFlag"`
	WinnerTeam       int                            `json:"winnerTeam"`
	IsHumanTurn      bool                           `json:"isHumanTurn"`
	IsHumanScarto    bool                           `json:"isHumanScarto"`
	Hint             *WebOutputCardHint             `json:"hint,omitempty"`
	WebOutputBase
	Config MinchiateWebOutputConfig `json:"config"`
}

// MinchiateWebOutputConfig ミンキアーテの設定アウトプット
type MinchiateWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	TargetRounds  int `json:"targetRounds"`
}

// ToConfig builds a MinchiateConfig from the nested web config, applying bounds checking.
func (c *MinchiateWebConfig) ToConfig() domain.MinchiateConfig {
	cfg := domain.DefaultMinchiateConfig()
	cfg.CpuDifficulty = domain.MinchiateCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty,
		int(domain.MinchiateCpuDifficultyEasy), int(domain.MinchiateCpuDifficultyHard), int(cfg.CpuDifficulty)))
	// **下限は MinchiatePlayerCnt。**1 を許すと境界検査は通るが Validate が落とすので、
	// リセットが黙って無視される。倍数条件は Validate 側が見る。
	webutil.ApplyBoundedInt(&cfg.TargetRounds, c.TargetRounds, domain.MinchiatePlayerCnt, 1000)
	return cfg
}

// ToConfig builds a MinchiateConfig from the web input.
func (p MinchiateWebInput) ToConfig() domain.MinchiateConfig {
	return configOrDefault(p.Config, (*MinchiateWebConfig).ToConfig, domain.DefaultMinchiateConfig())
}

// MinchiateWebController ミンキアーテのWebコントローラークラス
type MinchiateWebController = GameWebController[usecase.MinchiateInteractorIF, MinchiateWebInput, *MinchiateWebOutput]

// NewMinchiateWebController and NewMinchiateWebControllerWithProvider are the standard
// and provider-backed constructors for MinchiateWebController.
var NewMinchiateWebController, NewMinchiateWebControllerWithProvider = webControllerPair[usecase.MinchiateInteractorIF, MinchiateWebInput, *MinchiateWebOutput](
	newMinchiateDefaultOutput, minchiateDispatch,
)

func newMinchiateDefaultOutput(msg string) *MinchiateWebOutput {
	return &MinchiateWebOutput{
		Players:         make([]*MinchiateWebOutputPlayer, 0),
		CurrentTrick:    make([]*WebOutputTrickCard, 0),
		PlayableIndices: make([]int, 0),
		LastTrickWinner: -1,
		WinnerTeam:      -1,
		WebOutputBase:   WebOutputBase{Message: msg},
	}
}

func minchiateDispatch(bc *baseController, w http.ResponseWriter, di usecase.MinchiateInteractorIF, param MinchiateWebInput, newDefault func(string) *MinchiateWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, di.ResetWithConfig(param.ToConfig()))
	case "s", "scarto", "d", "discard":
		if !requireParam(bc, w, newDefault, param.CardIndices == nil, "param error: cardIndices is required.") {
			return true
		}
		bc.writePresenterResponse(w, di.Discard(param.CardIndices))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, di.Play(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, di.NextTrick())
	case "nr", "nextround":
		bc.writePresenterResponse(w, di.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, di.Hint, di.ActionLog)
	}
	return true
}
