package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// CrazyEightsWebInput クレイジーエイトWebインプット
type CrazyEightsWebInput struct {
	BaseWebInput
	CardIndex *int                  `json:"cardIndex,omitempty"`
	Suit      *int                  `json:"suit,omitempty"`
	Config    *CrazyEightsWebConfig `json:"config,omitempty"`
}

// CrazyEightsWebConfig クレイジーエイトWeb設定
type CrazyEightsWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	PointLimit    *int `json:"pointLimit,omitempty"`
}

// CrazyEightsWebOutputPlayer クレイジーエイトWebアウトプットプレイヤー
type CrazyEightsWebOutputPlayer struct {
	ID              int              `json:"id"`
	IsHuman         bool             `json:"isHuman"`
	CardCount       int              `json:"cardCount"`
	Cards           []*WebOutputCard `json:"cards"`
	RoundScore      int              `json:"roundScore"`
	CumulativeScore int              `json:"cumulativeScore"`
}

// CrazyEightsWebOutput クレイジーエイトWebアウトプット
type CrazyEightsWebOutput struct {
	Players          []*CrazyEightsWebOutputPlayer `json:"players"`
	Phase            int                           `json:"phase"`
	RoundNumber      int                           `json:"roundNumber"`
	CurrentPlayerIdx int                           `json:"currentPlayerIdx"`
	DiscardTop       *WebOutputCard                `json:"discardTop"`
	DrawPileCount    int                           `json:"drawPileCount"`
	ChosenSuit       int                           `json:"chosenSuit"`
	GameEndFlag      bool                          `json:"gameEndFlag"`
	WinnerIdx        int                           `json:"winnerIdx"`
	WebOutputBase
	Config CrazyEightsWebOutputConfig `json:"config"`
}

// CrazyEightsWebOutputConfig クレイジーエイト設定アウトプット
type CrazyEightsWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	PointLimit    int `json:"pointLimit"`
}

// ToConfig builds a CrazyEightsConfig from the nested web config, applying bounds checking.
func (c *CrazyEightsWebConfig) ToConfig() domain.CrazyEightsConfig {
	cfg := domain.DefaultCrazyEightsConfig()
	cfg.CpuDifficulty = domain.CrazyEightsCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.CrazyEightsCpuDifficultyEasy), int(domain.CrazyEightsCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.PointLimit, c.PointLimit, 1, 1000)
	return cfg
}

// ToConfig builds a CrazyEightsConfig from the web input.
func (p CrazyEightsWebInput) ToConfig() domain.CrazyEightsConfig {
	return configOrDefault(p.Config, (*CrazyEightsWebConfig).ToConfig, domain.DefaultCrazyEightsConfig())
}

// CrazyEightsWebController クレイジーエイトWebコントローラークラス
type CrazyEightsWebController = GameWebController[usecase.CrazyEightsInteractorIF, CrazyEightsWebInput, *CrazyEightsWebOutput]

// NewCrazyEightsWebController and NewCrazyEightsWebControllerWithProvider are
// the standard and provider-backed constructors for CrazyEightsWebController.
var NewCrazyEightsWebController, NewCrazyEightsWebControllerWithProvider = webControllerPair[usecase.CrazyEightsInteractorIF, CrazyEightsWebInput, *CrazyEightsWebOutput](
	newCrazyEightsDefaultOutput, crazyEightsDispatch,
)

func newCrazyEightsDefaultOutput(msg string) *CrazyEightsWebOutput {
	return &CrazyEightsWebOutput{
		Players:       make([]*CrazyEightsWebOutputPlayer, 0),
		WinnerIdx:     -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func crazyEightsDispatch(bc *baseController, w http.ResponseWriter, ci usecase.CrazyEightsInteractorIF, param CrazyEightsWebInput, newDefault func(string) *CrazyEightsWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, ci.ResetWithConfig(param.ToConfig()))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.Play(*param.CardIndex))
	case "d", "draw":
		bc.writePresenterResponse(w, ci.Draw())
	case "s", "suit":
		if !requireParam(bc, w, newDefault, param.Suit == nil, "param error: suit is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.ChooseSuit(*param.Suit))
	case "nr", "nextround":
		bc.writePresenterResponse(w, ci.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, ci.Hint, ci.ActionLog)
	}
	return true
}
