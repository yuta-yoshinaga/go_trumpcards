package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// PageOneWebInput ページワンWebインプット
type PageOneWebInput struct {
	BaseWebInput
	CardIndex *int              `json:"cardIndex,omitempty"`
	Config    *PageOneWebConfig `json:"config,omitempty"`
}

// PageOneWebConfig ページワンWeb設定
type PageOneWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	PointLimit    *int `json:"pointLimit,omitempty"`
}

// PageOneWebOutputPlayer ページワンWebアウトプットプレイヤー
type PageOneWebOutputPlayer struct {
	ID              int              `json:"id"`
	IsHuman         bool             `json:"isHuman"`
	CardCount       int              `json:"cardCount"`
	Cards           []*WebOutputCard `json:"cards"`
	RoundScore      int              `json:"roundScore"`
	CumulativeScore int              `json:"cumulativeScore"`
	HasDeclared     bool             `json:"hasDeclared"`
}

// PageOneWebOutput ページワンWebアウトプット
type PageOneWebOutput struct {
	Players          []*PageOneWebOutputPlayer `json:"players"`
	Phase            int                       `json:"phase"`
	RoundNumber      int                       `json:"roundNumber"`
	CurrentPlayerIdx int                       `json:"currentPlayerIdx"`
	DiscardTop       *WebOutputCard            `json:"discardTop"`
	DrawPileCount    int                       `json:"drawPileCount"`
	GameEndFlag      bool                      `json:"gameEndFlag"`
	WinnerIdx        int                       `json:"winnerIdx"`
	WebOutputBase
	Config PageOneWebOutputConfig `json:"config"`
}

// PageOneWebOutputConfig ページワン設定アウトプット
type PageOneWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	PointLimit    int `json:"pointLimit"`
}

// ToConfig builds a PageOneConfig from the nested web config, applying bounds checking.
func (c *PageOneWebConfig) ToConfig() domain.PageOneConfig {
	cfg := domain.DefaultPageOneConfig()
	cfg.CpuDifficulty = domain.PageOneCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.PageOneCpuDifficultyEasy), int(domain.PageOneCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.PointLimit, c.PointLimit, 1, 1000)
	return cfg
}

// ToConfig builds a PageOneConfig from the web input.
func (p PageOneWebInput) ToConfig() domain.PageOneConfig {
	return configOrDefault(p.Config, (*PageOneWebConfig).ToConfig, domain.DefaultPageOneConfig())
}

// PageOneWebController ページワンWebコントローラークラス
type PageOneWebController = GameWebController[usecase.PageOneInteractorIF, PageOneWebInput, *PageOneWebOutput]

// NewPageOneWebController and NewPageOneWebControllerWithProvider are the
// standard and provider-backed constructors for PageOneWebController.
var NewPageOneWebController, NewPageOneWebControllerWithProvider = webControllerPair[usecase.PageOneInteractorIF, PageOneWebInput, *PageOneWebOutput](
	newPageOneDefaultOutput, pageOneDispatch,
)

func newPageOneDefaultOutput(msg string) *PageOneWebOutput {
	return &PageOneWebOutput{
		Players:       make([]*PageOneWebOutputPlayer, 0),
		WinnerIdx:     -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func pageOneDispatch(bc *baseController, w http.ResponseWriter, ci usecase.PageOneInteractorIF, param PageOneWebInput, newDefault func(string) *PageOneWebOutput) bool {
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
	case "dc", "declare":
		bc.writePresenterResponse(w, ci.Declare())
	case "sk", "skip":
		bc.writePresenterResponse(w, ci.SkipDeclare())
	case "nr", "nextround":
		bc.writePresenterResponse(w, ci.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, ci.Hint, ci.ActionLog)
	}
	return true
}
