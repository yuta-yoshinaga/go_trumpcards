package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// PrsiWebInput プルシーWebインプット
type PrsiWebInput struct {
	BaseWebInput
	CardIndex *int           `json:"cardIndex,omitempty"`
	Config    *PrsiWebConfig `json:"config,omitempty"`
}

// PrsiWebConfig プルシーWeb設定
type PrsiWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
}

// PrsiWebOutputPlayer プルシーWebアウトプットプレイヤー
type PrsiWebOutputPlayer struct {
	ID        int              `json:"id"`
	IsHuman   bool             `json:"isHuman"`
	CardCount int              `json:"cardCount"`
	Cards     []*WebOutputCard `json:"cards"`
}

// PrsiWebOutput プルシーWebアウトプット
type PrsiWebOutput struct {
	Players          []*PrsiWebOutputPlayer `json:"players"`
	Phase            int                    `json:"phase"`
	CurrentPlayerIdx int                    `json:"currentPlayerIdx"`
	DiscardTop       *WebOutputCard         `json:"discardTop"`
	DrawPileCount    int                    `json:"drawPileCount"`
	PenaltyDrawCount int                    `json:"penaltyDrawCount"`
	PendingSkips     int                    `json:"pendingSkips"`
	GameEndFlag      bool                   `json:"gameEndFlag"`
	WinnerIdx        int                    `json:"winnerIdx"`
	WebOutputBase
	Config PrsiWebOutputConfig `json:"config"`
}

// PrsiWebOutputConfig プルシー設定アウトプット
type PrsiWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
}

// ToConfig builds a PrsiConfig from the nested web config, applying bounds checking.
func (c *PrsiWebConfig) ToConfig() domain.PrsiConfig {
	cfg := domain.DefaultPrsiConfig()
	cfg.CpuDifficulty = domain.PrsiCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.PrsiCpuDifficultyEasy), int(domain.PrsiCpuDifficultyHard), int(cfg.CpuDifficulty)))
	return cfg
}

// ToConfig builds a PrsiConfig from the web input.
func (p PrsiWebInput) ToConfig() domain.PrsiConfig {
	return configOrDefault(p.Config, (*PrsiWebConfig).ToConfig, domain.DefaultPrsiConfig())
}

// PrsiWebController プルシーWebコントローラークラス
type PrsiWebController = GameWebController[usecase.PrsiInteractorIF, PrsiWebInput, *PrsiWebOutput]

// NewPrsiWebController and NewPrsiWebControllerWithProvider are the standard
// and provider-backed constructors for PrsiWebController.
var NewPrsiWebController, NewPrsiWebControllerWithProvider = webControllerPair[usecase.PrsiInteractorIF, PrsiWebInput, *PrsiWebOutput](
	newPrsiDefaultOutput, prsiDispatch,
)

func newPrsiDefaultOutput(msg string) *PrsiWebOutput {
	return &PrsiWebOutput{
		Players:       make([]*PrsiWebOutputPlayer, 0),
		WinnerIdx:     -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func prsiDispatch(bc *baseController, w http.ResponseWriter, ci usecase.PrsiInteractorIF, param PrsiWebInput, newDefault func(string) *PrsiWebOutput) bool {
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
	default:
		return dispatchLog(param.Command, bc, w, ci.ActionLog)
	}
	return true
}
