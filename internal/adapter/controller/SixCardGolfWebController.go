package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// SixCardGolfWebInput SixCardGolf Webインプット
type SixCardGolfWebInput struct {
	BaseWebInput
	Position *int                  `json:"position,omitempty"`
	Config   *SixCardGolfWebConfig `json:"config,omitempty"`
}

// SixCardGolfWebConfig SixCardGolf Web設定
type SixCardGolfWebConfig struct {
	PlayerCount   *int `json:"playerCount,omitempty"`
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	Rounds        *int `json:"rounds,omitempty"`
}

// SixCardGolfWebOutputSlot グリッドスロットアウトプット
type SixCardGolfWebOutputSlot struct {
	Card   *WebOutputCard `json:"card"`
	FaceUp bool           `json:"faceUp"`
}

// SixCardGolfWebOutputPlayer プレイヤーアウトプット
type SixCardGolfWebOutputPlayer struct {
	ID              int                         `json:"id"`
	IsHuman         bool                        `json:"isHuman"`
	Grid            []*SixCardGolfWebOutputSlot `json:"grid"`
	RoundScore      int                         `json:"roundScore"`
	CumulativeScore int                         `json:"cumulativeScore"`
	AllFaceUp       bool                        `json:"allFaceUp"`
}

// SixCardGolfWebOutput SixCardGolf Webアウトプット
type SixCardGolfWebOutput struct {
	Players          []*SixCardGolfWebOutputPlayer `json:"players"`
	Phase            int                           `json:"phase"`
	RoundNumber      int                           `json:"roundNumber"`
	TotalRounds      int                           `json:"totalRounds"`
	CurrentPlayerIdx int                           `json:"currentPlayerIdx"`
	DiscardTop       *WebOutputCard                `json:"discardTop"`
	DrawPileCount    int                           `json:"drawPileCount"`
	DrawnCard        *WebOutputCard                `json:"drawnCard"`
	DrawnFromDiscard bool                          `json:"drawnFromDiscard"`
	CanFlip          bool                          `json:"canFlip"`
	FinalTurnTrigger int                           `json:"finalTurnTrigger"`
	GameEndFlag      bool                          `json:"gameEndFlag"`
	WinnerIdx        int                           `json:"winnerIdx"`
	WebOutputBase
	Config SixCardGolfWebOutputConfig `json:"config"`
}

// SixCardGolfWebOutputConfig 設定アウトプット
type SixCardGolfWebOutputConfig struct {
	PlayerCount   int `json:"playerCount"`
	CpuDifficulty int `json:"cpuDifficulty"`
	Rounds        int `json:"rounds"`
}

// ToConfig builds a SixCardGolfConfig from the nested web config.
func (c *SixCardGolfWebConfig) ToConfig() domain.SixCardGolfConfig {
	cfg := domain.DefaultSixCardGolfConfig()
	webutil.ApplyBoundedInt(&cfg.PlayerCount, c.PlayerCount, domain.SixCardGolfPlayerMin, domain.SixCardGolfPlayerMax)
	cfg.CpuDifficulty = domain.SixCardGolfCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.SixCardGolfCpuEasy), int(domain.SixCardGolfCpuHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.Rounds, c.Rounds, 1, 18)
	return cfg
}

// ToConfig builds a SixCardGolfConfig from the web input.
func (p SixCardGolfWebInput) ToConfig() domain.SixCardGolfConfig {
	return configOrDefault(p.Config, (*SixCardGolfWebConfig).ToConfig, domain.DefaultSixCardGolfConfig())
}

// SixCardGolfWebController SixCardGolf Webコントローラー
type SixCardGolfWebController = GameWebController[usecase.SixCardGolfInteractorIF, SixCardGolfWebInput, *SixCardGolfWebOutput]

// NewSixCardGolfWebController and NewSixCardGolfWebControllerWithProvider are
// the standard and provider-backed constructors for SixCardGolfWebController.
var NewSixCardGolfWebController, NewSixCardGolfWebControllerWithProvider = webControllerPair[usecase.SixCardGolfInteractorIF, SixCardGolfWebInput, *SixCardGolfWebOutput](
	newSixCardGolfDefaultOutput, sixCardGolfDispatch,
)

func newSixCardGolfDefaultOutput(msg string) *SixCardGolfWebOutput {
	return &SixCardGolfWebOutput{
		Players:          make([]*SixCardGolfWebOutputPlayer, 0),
		WinnerIdx:        -1,
		FinalTurnTrigger: -1,
		WebOutputBase:    WebOutputBase{Message: msg},
	}
}

func sixCardGolfDispatch(bc *baseController, w http.ResponseWriter, ci usecase.SixCardGolfInteractorIF, param SixCardGolfWebInput, newDefault func(string) *SixCardGolfWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, ci.ResetWithConfig(param.ToConfig()))
	case "fi", "flipinitial":
		if !requireParam(bc, w, newDefault, param.Position == nil, "param error: position is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.FlipInitial(*param.Position))
	case "ds", "drawstock":
		bc.writePresenterResponse(w, ci.DrawStock())
	case "dd", "drawdiscard":
		bc.writePresenterResponse(w, ci.DrawDiscard())
	case "sw", "swap":
		if !requireParam(bc, w, newDefault, param.Position == nil, "param error: position is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.SwapCard(*param.Position))
	case "di", "discard":
		bc.writePresenterResponse(w, ci.DiscardDrawn())
	case "fl", "flip":
		if !requireParam(bc, w, newDefault, param.Position == nil, "param error: position is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.FlipCard(*param.Position))
	case "sf", "skipflip":
		bc.writePresenterResponse(w, ci.SkipFlip())
	case "nr", "nextround":
		bc.writePresenterResponse(w, ci.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, ci.Hint, ci.ActionLog)
	}
	return true
}
