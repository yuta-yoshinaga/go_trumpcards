package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// WizardWebInput ウィザードWebインプット
type WizardWebInput struct {
	BaseWebInput
	Bid       *int             `json:"bid,omitempty"`
	CardIndex *int             `json:"cardIndex,omitempty"`
	Config    *WizardWebConfig `json:"config,omitempty"`
}

// WizardWebConfig ウィザードWeb設定
type WizardWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
}

// WizardWebOutputPlayer ウィザードWebアウトプットプレイヤー
type WizardWebOutputPlayer struct {
	ID              int              `json:"id"`
	IsHuman         bool             `json:"isHuman"`
	CardCount       int              `json:"cardCount"`
	Cards           []*WebOutputCard `json:"cards"`
	Bid             int              `json:"bid"`
	RoundScore      int              `json:"roundScore"`
	CumulativeScore int              `json:"cumulativeScore"`
	TrickCount      int              `json:"trickCount"`
}

// WizardWebOutputHint ヒント出力
type WizardWebOutputHint struct {
	CardIndex *int   `json:"cardIndex,omitempty"`
	Bid       *int   `json:"bid,omitempty"`
	Reason    string `json:"reason"`
}

// WizardWebOutput ウィザードWebアウトプット
type WizardWebOutput struct {
	Players          []*WizardWebOutputPlayer `json:"players"`
	Phase            int                      `json:"phase"`
	RoundNumber      int                      `json:"roundNumber"`
	TotalRounds      int                      `json:"totalRounds"`
	HandSize         int                      `json:"handSize"`
	TrickNumber      int                      `json:"trickNumber"`
	CurrentPlayerIdx int                      `json:"currentPlayerIdx"`
	BidPlayerIdx     int                      `json:"bidPlayerIdx"`
	DealerIdx        int                      `json:"dealerIdx"`
	TrumpCard        *WebOutputCard           `json:"trumpCard"`
	TrumpSuit        int                      `json:"trumpSuit"`
	RestrictedBid    int                      `json:"restrictedBid"`
	CurrentTrick     []*WebOutputTrickCard    `json:"currentTrick"`
	GameEndFlag      bool                     `json:"gameEndFlag"`
	WinnerIdx        int                      `json:"winnerIdx"`
	LeadPlayerIdx    int                      `json:"leadPlayerIdx"`
	Hint             *WizardWebOutputHint     `json:"hint,omitempty"`
	WebOutputBase
	Config WizardWebOutputConfig `json:"config"`
}

// WizardWebOutputConfig ウィザード設定アウトプット
type WizardWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
}

// ToConfig builds a WizardConfig from the nested web config, applying bounds checking.
func (c *WizardWebConfig) ToConfig() domain.WizardConfig {
	cfg := domain.DefaultWizardConfig()
	cfg.CpuDifficulty = domain.WizardCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.WizardCpuDifficultyEasy), int(domain.WizardCpuDifficultyHard), int(cfg.CpuDifficulty)))
	return cfg
}

// ToConfig builds a WizardConfig from the web input.
func (p WizardWebInput) ToConfig() domain.WizardConfig {
	return configOrDefault(p.Config, (*WizardWebConfig).ToConfig, domain.DefaultWizardConfig())
}

// WizardWebController ウィザードWebコントローラークラス
type WizardWebController = GameWebController[usecase.WizardInteractorIF, WizardWebInput, *WizardWebOutput]

// NewWizardWebController and NewWizardWebControllerWithProvider are
// the standard and provider-backed constructors for WizardWebController.
var NewWizardWebController, NewWizardWebControllerWithProvider = webControllerPair[usecase.WizardInteractorIF, WizardWebInput, *WizardWebOutput](
	newWizardDefaultOutput, wizardDispatch,
)

func newWizardDefaultOutput(msg string) *WizardWebOutput {
	return &WizardWebOutput{
		Players:       make([]*WizardWebOutputPlayer, 0),
		CurrentTrick:  make([]*WebOutputTrickCard, 0),
		WinnerIdx:     -1,
		TrumpSuit:     -1,
		RestrictedBid: -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func wizardDispatch(bc *baseController, w http.ResponseWriter, oi usecase.WizardInteractorIF, param WizardWebInput, newDefault func(string) *WizardWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, oi.ResetWithConfig(param.ToConfig()))
	case "b", "bid":
		if !requireParam(bc, w, newDefault, param.Bid == nil, "param error: bid is required.") {
			return true
		}
		bc.writePresenterResponse(w, oi.Bid(*param.Bid))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, oi.Play(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, oi.NextTrick())
	case "nr", "nextround":
		bc.writePresenterResponse(w, oi.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, oi.Hint, oi.ActionLog)
	}
	return true
}
