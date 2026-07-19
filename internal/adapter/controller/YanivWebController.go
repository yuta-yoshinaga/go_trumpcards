//go:build !js || !wasm || solo

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// YanivWebInput Yaniv Webインプット
type YanivWebInput struct {
	BaseWebInput
	CardIndices []int           `json:"cardIndices,omitempty"`
	End         *int            `json:"end,omitempty"`
	Config      *YanivWebConfig `json:"config,omitempty"`
}

// YanivWebConfig Yaniv Web設定
type YanivWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	ScoreLimit    *int `json:"scoreLimit,omitempty"`
}

// YanivWebOutputPlayer Yaniv Webアウトプットプレイヤー
type YanivWebOutputPlayer struct {
	ID           int              `json:"id"`
	IsHuman      bool             `json:"isHuman"`
	CardCount    int              `json:"cardCount"`
	Cards        []*WebOutputCard `json:"cards"`
	Score        int              `json:"score"`
	HandTotal    int              `json:"handTotal"`
	IsEliminated bool             `json:"isEliminated"`
}

// YanivWebOutput Yaniv Webアウトプット
type YanivWebOutput struct {
	Players          []*YanivWebOutputPlayer `json:"players"`
	Phase            int                     `json:"phase"`
	RoundNumber      int                     `json:"roundNumber"`
	CurrentPlayerIdx int                     `json:"currentPlayerIdx"`
	PickupCards      []*WebOutputCard        `json:"pickupCards"`
	DrawPileCount    int                     `json:"drawPileCount"`
	GameEndFlag      bool                    `json:"gameEndFlag"`
	WinnerIdx        int                     `json:"winnerIdx"`
	CallerIdx        int                     `json:"callerIdx"`
	AsafWinnerIdx    int                     `json:"asafWinnerIdx"`
	IsAsaf           bool                    `json:"isAsaf"`
	RoundScores      []int                   `json:"roundScores"`
	WebOutputBase
	Config YanivWebOutputConfig `json:"config"`
}

// YanivWebOutputConfig Yaniv設定アウトプット
type YanivWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	ScoreLimit    int `json:"scoreLimit"`
}

// ToConfig builds a YanivConfig from the nested web config, applying bounds checking.
func (c *YanivWebConfig) ToConfig() domain.YanivConfig {
	cfg := domain.DefaultYanivConfig()
	cfg.CpuDifficulty = domain.YanivCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.YanivCpuDifficultyEasy), int(domain.YanivCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.ScoreLimit, c.ScoreLimit, domain.YanivMinScoreLimit, domain.YanivMaxScoreLimit)
	return cfg
}

// ToConfig builds a YanivConfig from the web input.
func (p YanivWebInput) ToConfig() domain.YanivConfig {
	return configOrDefault(p.Config, (*YanivWebConfig).ToConfig, domain.DefaultYanivConfig())
}

// YanivWebController Yaniv Webコントローラークラス
type YanivWebController = GameWebController[usecase.YanivInteractorIF, YanivWebInput, *YanivWebOutput]

// NewYanivWebController and NewYanivWebControllerWithProvider are the standard
// and provider-backed constructors for YanivWebController.
var NewYanivWebController, NewYanivWebControllerWithProvider = webControllerPair[usecase.YanivInteractorIF, YanivWebInput, *YanivWebOutput](
	newYanivDefaultOutput, yanivDispatch,
)

func newYanivDefaultOutput(msg string) *YanivWebOutput {
	return &YanivWebOutput{
		Players:       make([]*YanivWebOutputPlayer, 0),
		PickupCards:   make([]*WebOutputCard, 0),
		WinnerIdx:     -1,
		CallerIdx:     -1,
		AsafWinnerIdx: -1,
		RoundScores:   make([]int, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func yanivDispatch(bc *baseController, w http.ResponseWriter, ci usecase.YanivInteractorIF, param YanivWebInput, newDefault func(string) *YanivWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, ci.ResetWithConfig(param.ToConfig()))
	case "d", "discard":
		indices := param.CardIndices
		if indices == nil {
			indices = []int{}
		}
		bc.writePresenterResponse(w, ci.Discard(indices))
	case "y", "yaniv":
		bc.writePresenterResponse(w, ci.DeclareYaniv())
	case "ds", "drawstock":
		bc.writePresenterResponse(w, ci.DrawFromStock())
	case "dp", "drawpickup":
		if !requireParam(bc, w, newDefault, param.End == nil, "param error: end is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.DrawFromPickup(*param.End))
	case "nr", "nextround":
		bc.writePresenterResponse(w, ci.NextRound())
	default:
		return dispatchLog(param.Command, bc, w, ci.ActionLog)
	}
	return true
}
