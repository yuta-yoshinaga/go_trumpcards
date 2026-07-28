//go:build !js || !wasm || casino

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// EcarteWebInput エカルテWebインプット
type EcarteWebInput struct {
	BaseWebInput
	Accept         *bool            `json:"accept,omitempty"`
	CardIndex      *int             `json:"cardIndex,omitempty"`
	DiscardIndices []int            `json:"discardIndices,omitempty"`
	Config         *EcarteWebConfig `json:"config,omitempty"`
}

// EcarteWebConfig エカルテWeb設定
type EcarteWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	TargetScore   *int `json:"targetScore,omitempty"`
}

// EcarteWebOutputPlayer エカルテWebアウトプットプレイヤー
type EcarteWebOutputPlayer struct {
	ID              int              `json:"id"`
	IsHuman         bool             `json:"isHuman"`
	CardCount       int              `json:"cardCount"`
	Cards           []*WebOutputCard `json:"cards"`
	RoundScore      int              `json:"roundScore"`
	CumulativeScore int              `json:"cumulativeScore"`
	TrickCount      int              `json:"trickCount"`
}

// EcarteWebOutputHint ヒント出力
type EcarteWebOutputHint struct {
	CardIndex *int   `json:"cardIndex,omitempty"`
	Action    string `json:"action,omitempty"`
	Reason    string `json:"reason"`
}

// EcarteWebOutput エカルテWebアウトプット
type EcarteWebOutput struct {
	Players          []*EcarteWebOutputPlayer `json:"players"`
	DealPoints       []int                    `json:"dealPoints"`
	MatchScore       []int                    `json:"matchScore"`
	Phase            int                      `json:"phase"`
	NegStep          int                      `json:"negStep"`
	RoundNumber      int                      `json:"roundNumber"`
	TrickNumber      int                      `json:"trickNumber"`
	CurrentPlayerIdx int                      `json:"currentPlayerIdx"`
	DealerIdx        int                      `json:"dealerIdx"`
	ElderIdx         int                      `json:"elderIdx"`
	LeadPlayerIdx    int                      `json:"leadPlayerIdx"`
	TrumpSuit        int                      `json:"trumpSuit"`
	TrumpCard        *WebOutputCard           `json:"trumpCard,omitempty"`
	CurrentTrick     []*WebOutputTrickCard    `json:"currentTrick"`
	StockRemaining   int                      `json:"stockRemaining"`
	RefusalByDealer  bool                     `json:"refusalByDealer"`
	ValidPlays       []int                    `json:"validPlays"`
	GameEndFlag      bool                     `json:"gameEndFlag"`
	WinnerIdx        int                      `json:"winnerIdx"`
	Hint             *EcarteWebOutputHint     `json:"hint,omitempty"`
	WebOutputBase
	Config EcarteWebOutputConfig `json:"config"`
}

// EcarteWebOutputConfig エカルテ設定アウトプット
type EcarteWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	TargetScore   int `json:"targetScore"`
}

// ToConfig builds an EcarteConfig from the nested web config, applying bounds checking.
func (c *EcarteWebConfig) ToConfig() domain.EcarteConfig {
	cfg := domain.DefaultEcarteConfig()
	cfg.CpuDifficulty = domain.EcarteCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty,
		int(domain.EcarteCpuDifficultyEasy), int(domain.EcarteCpuDifficultyHard),
		int(cfg.CpuDifficulty)))
	cfg.TargetScore = webutil.BoundedIntPtr(c.TargetScore, 1, domain.EcarteMaxTargetScore, cfg.TargetScore)
	return cfg
}

// ToConfig builds an EcarteConfig from the web input.
func (p EcarteWebInput) ToConfig() domain.EcarteConfig {
	return configOrDefault(p.Config, (*EcarteWebConfig).ToConfig, domain.DefaultEcarteConfig())
}

// EcarteWebController エカルテWebコントローラークラス
type EcarteWebController = GameWebController[usecase.EcarteInteractorIF, EcarteWebInput, *EcarteWebOutput]

// NewEcarteWebController and NewEcarteWebControllerWithProvider are
// the standard and provider-backed constructors for EcarteWebController.
var NewEcarteWebController, NewEcarteWebControllerWithProvider = webControllerPair[usecase.EcarteInteractorIF, EcarteWebInput, *EcarteWebOutput](
	newEcarteDefaultOutput, ecarteDispatch,
)

func newEcarteDefaultOutput(msg string) *EcarteWebOutput {
	return &EcarteWebOutput{
		Players:       make([]*EcarteWebOutputPlayer, 0),
		DealPoints:    make([]int, 0),
		MatchScore:    make([]int, 0),
		CurrentTrick:  make([]*WebOutputTrickCard, 0),
		ValidPlays:    make([]int, 0),
		WinnerIdx:     -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func ecarteDispatch(bc *baseController, w http.ResponseWriter, ei usecase.EcarteInteractorIF, param EcarteWebInput, newDefault func(string) *EcarteWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, ei.ResetWithConfig(param.ToConfig()))
	case "propose":
		bc.writePresenterResponse(w, ei.Propose())
	case "stand":
		bc.writePresenterResponse(w, ei.Stand())
	case "respond":
		if !requireParam(bc, w, newDefault, param.Accept == nil, "param error: accept is required.") {
			return true
		}
		bc.writePresenterResponse(w, ei.Respond(*param.Accept))
	case "discard":
		indices := param.DiscardIndices
		if indices == nil {
			indices = []int{}
		}
		bc.writePresenterResponse(w, ei.Discard(indices))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, ei.Play(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, ei.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, ei.Hint, ei.ActionLog)
	}
	return true
}
