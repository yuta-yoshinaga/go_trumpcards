//go:build !js || !wasm || extra4

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// PutWebInput プットWebインプット
type PutWebInput struct {
	BaseWebInput
	CardIndex *int          `json:"cardIndex,omitempty"`
	Config    *PutWebConfig `json:"config,omitempty"`
}

// PutWebConfig プットWeb設定
type PutWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	MatchTarget   *int `json:"matchTarget,omitempty"`
}

// PutWebOutputPlayer プットWebアウトプットプレイヤー
type PutWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
	TrickCount int              `json:"trickCount"`
}

// PutWebOutputHint ヒント出力
type PutWebOutputHint struct {
	Action    string `json:"action"`
	CardIndex *int   `json:"cardIndex,omitempty"`
	Reason    string `json:"reason"`
}

// PutWebOutput プットWebアウトプット
type PutWebOutput struct {
	Players          []*PutWebOutputPlayer `json:"players"`
	Phase            int                   `json:"phase"`
	HandNumber       int                   `json:"handNumber"`
	TrickNumber      int                   `json:"trickNumber"`
	CurrentPlayerIdx int                   `json:"currentPlayerIdx"`
	ResponderIdx     int                   `json:"responderIdx"`
	CurrentTrick     []*WebOutputTrickCard `json:"currentTrick"`
	TrickResults     []int                 `json:"trickResults"`
	LeadPlayerIdx    int                   `json:"leadPlayerIdx"`
	ManoIdx          int                   `json:"manoIdx"`
	DealerIdx        int                   `json:"dealerIdx"`
	HandStake        int                   `json:"handStake"`
	AcceptedLevel    int                   `json:"acceptedLevel"`
	PendingLevel     int                   `json:"pendingLevel"`
	PutCallerIdx     int                   `json:"putCallerIdx"`
	CanDeclarePut    bool                  `json:"canDeclarePut"`
	MatchTarget      int                   `json:"matchTarget"`
	MatchPoints      []int                 `json:"matchPoints"`
	HandWinnerIdx    int                   `json:"handWinnerIdx"`
	GameEndFlag      bool                  `json:"gameEndFlag"`
	WinnerIdx        int                   `json:"winnerIdx"`
	Hint             *PutWebOutputHint     `json:"hint,omitempty"`
	WebOutputBase
	Config PutWebOutputConfig `json:"config"`
}

// PutWebOutputConfig プット設定アウトプット
type PutWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	MatchTarget   int `json:"matchTarget"`
}

// ToConfig builds a PutConfig from the nested web config, applying bounds checking.
func (c *PutWebConfig) ToConfig() domain.PutConfig {
	cfg := domain.DefaultPutConfig()
	cfg.CpuDifficulty = domain.PutCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty,
		int(domain.PutCpuDifficultyNormal), int(domain.PutCpuDifficultyNormal),
		int(cfg.CpuDifficulty)))
	cfg.MatchTarget = webutil.BoundedIntPtr(c.MatchTarget,
		domain.PutMinMatchTarget, domain.PutMaxMatchTarget, cfg.MatchTarget)
	return cfg
}

// ToConfig builds a PutConfig from the web input.
func (p PutWebInput) ToConfig() domain.PutConfig {
	return configOrDefault(p.Config, (*PutWebConfig).ToConfig, domain.DefaultPutConfig())
}

// PutWebController プットWebコントローラークラス
type PutWebController = GameWebController[usecase.PutInteractorIF, PutWebInput, *PutWebOutput]

// NewPutWebController and NewPutWebControllerWithProvider are the standard
// and provider-backed constructors for PutWebController.
var NewPutWebController, NewPutWebControllerWithProvider = webControllerPair[usecase.PutInteractorIF, PutWebInput, *PutWebOutput](
	newPutDefaultOutput, putDispatch,
)

func newPutDefaultOutput(msg string) *PutWebOutput {
	return &PutWebOutput{
		Players:       make([]*PutWebOutputPlayer, 0),
		CurrentTrick:  make([]*WebOutputTrickCard, 0),
		MatchPoints:   make([]int, 0),
		WinnerIdx:     -1,
		HandWinnerIdx: -1,
		ResponderIdx:  -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func putDispatch(bc *baseController, w http.ResponseWriter, ti usecase.PutInteractorIF, param PutWebInput, newDefault func(string) *PutWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, ti.ResetWithConfig(param.ToConfig()))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, ti.Play(*param.CardIndex))
	case "t", "put":
		bc.writePresenterResponse(w, ti.Put())
	case "a", "accept":
		bc.writePresenterResponse(w, ti.Respond(true))
	case "d", "decline":
		bc.writePresenterResponse(w, ti.Respond(false))
	case "n", "next":
		bc.writePresenterResponse(w, ti.Next())
	default:
		return dispatchHintAndLog(param.Command, bc, w, ti.Hint, ti.ActionLog)
	}
	return true
}
