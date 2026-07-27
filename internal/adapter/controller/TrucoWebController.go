package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// TrucoWebInput トゥルコWebインプット
type TrucoWebInput struct {
	BaseWebInput
	CardIndex *int            `json:"cardIndex,omitempty"`
	Config    *TrucoWebConfig `json:"config,omitempty"`
}

// TrucoWebConfig トゥルコWeb設定
type TrucoWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	MatchTarget   *int `json:"matchTarget,omitempty"`
}

// TrucoWebOutputPlayer トゥルコWebアウトプットプレイヤー
type TrucoWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
	TrickCount int              `json:"trickCount"`
}

// TrucoWebOutputHint ヒント出力
type TrucoWebOutputHint struct {
	Action    string `json:"action"`
	CardIndex *int   `json:"cardIndex,omitempty"`
	Reason    string `json:"reason"`
}

// TrucoWebOutput トゥルコWebアウトプット
type TrucoWebOutput struct {
	Players          []*TrucoWebOutputPlayer `json:"players"`
	Phase            int                     `json:"phase"`
	HandNumber       int                     `json:"handNumber"`
	TrickNumber      int                     `json:"trickNumber"`
	CurrentPlayerIdx int                     `json:"currentPlayerIdx"`
	ResponderIdx     int                     `json:"responderIdx"`
	CurrentTrick     []*WebOutputTrickCard   `json:"currentTrick"`
	TrickResults     []int                   `json:"trickResults"`
	LeadPlayerIdx    int                     `json:"leadPlayerIdx"`
	ManoIdx          int                     `json:"manoIdx"`
	DealerIdx        int                     `json:"dealerIdx"`
	HandStake        int                     `json:"handStake"`
	AcceptedLevel    int                     `json:"acceptedLevel"`
	PendingLevel     int                     `json:"pendingLevel"`
	TrucoCallerIdx   int                     `json:"trucoCallerIdx"`
	CanDeclareTruco  bool                    `json:"canDeclareTruco"`
	MatchTarget      int                     `json:"matchTarget"`
	MatchPoints      []int                   `json:"matchPoints"`
	HandWinnerIdx    int                     `json:"handWinnerIdx"`
	GameEndFlag      bool                    `json:"gameEndFlag"`
	WinnerIdx        int                     `json:"winnerIdx"`
	Hint             *TrucoWebOutputHint     `json:"hint,omitempty"`
	WebOutputBase
	Config TrucoWebOutputConfig `json:"config"`
}

// TrucoWebOutputConfig トゥルコ設定アウトプット
type TrucoWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	MatchTarget   int `json:"matchTarget"`
}

// ToConfig builds a TrucoConfig from the nested web config, applying bounds checking.
func (c *TrucoWebConfig) ToConfig() domain.TrucoConfig {
	cfg := domain.DefaultTrucoConfig()
	cfg.CpuDifficulty = domain.TrucoCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty,
		int(domain.TrucoCpuDifficultyNormal), int(domain.TrucoCpuDifficultyNormal),
		int(cfg.CpuDifficulty)))
	cfg.MatchTarget = webutil.BoundedIntPtr(c.MatchTarget,
		domain.TrucoMinMatchTarget, domain.TrucoMaxMatchTarget, cfg.MatchTarget)
	return cfg
}

// ToConfig builds a TrucoConfig from the web input.
func (p TrucoWebInput) ToConfig() domain.TrucoConfig {
	return configOrDefault(p.Config, (*TrucoWebConfig).ToConfig, domain.DefaultTrucoConfig())
}

// TrucoWebController トゥルコWebコントローラークラス
type TrucoWebController = GameWebController[usecase.TrucoInteractorIF, TrucoWebInput, *TrucoWebOutput]

// NewTrucoWebController and NewTrucoWebControllerWithProvider are the standard
// and provider-backed constructors for TrucoWebController.
var NewTrucoWebController, NewTrucoWebControllerWithProvider = webControllerPair[usecase.TrucoInteractorIF, TrucoWebInput, *TrucoWebOutput](
	newTrucoDefaultOutput, trucoDispatch,
)

func newTrucoDefaultOutput(msg string) *TrucoWebOutput {
	return &TrucoWebOutput{
		Players:       make([]*TrucoWebOutputPlayer, 0),
		CurrentTrick:  make([]*WebOutputTrickCard, 0),
		MatchPoints:   make([]int, 0),
		WinnerIdx:     -1,
		HandWinnerIdx: -1,
		ResponderIdx:  -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func trucoDispatch(bc *baseController, w http.ResponseWriter, ti usecase.TrucoInteractorIF, param TrucoWebInput, newDefault func(string) *TrucoWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, ti.ResetWithConfig(param.ToConfig()))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, ti.Play(*param.CardIndex))
	case "t", "truco":
		bc.writePresenterResponse(w, ti.Truco())
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
