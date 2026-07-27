//go:build !js || !wasm || extra

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// WattenWebInput ヴァッテンWebインプット
type WattenWebInput struct {
	BaseWebInput
	Rank      *int             `json:"rank,omitempty"`
	Suit      *int             `json:"suit,omitempty"`
	CardIndex *int             `json:"cardIndex,omitempty"`
	Hold      *bool            `json:"hold,omitempty"`
	Config    *WattenWebConfig `json:"config,omitempty"`
}

// WattenWebConfig ヴァッテンWeb設定
type WattenWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	TargetScore   *int `json:"targetScore,omitempty"`
	MaxRaises     *int `json:"maxRaises,omitempty"`
}

// WattenWebOutputPlayer ヴァッテンWebアウトプットプレイヤー
type WattenWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
	Team       int              `json:"team"`
	TrickCount int              `json:"trickCount"`
}

// WattenWebOutputHint ヒント出力
type WattenWebOutputHint struct {
	Action    string `json:"action"`
	CardIndex *int   `json:"cardIndex,omitempty"`
	Rank      *int   `json:"rank,omitempty"`
	Suit      *int   `json:"suit,omitempty"`
	Reason    string `json:"reason"`
}

// WattenWebOutput ヴァッテンWebアウトプット
type WattenWebOutput struct {
	Players          []*WattenWebOutputPlayer `json:"players"`
	Phase            int                      `json:"phase"`
	RoundNumber      int                      `json:"roundNumber"`
	TrickNumber      int                      `json:"trickNumber"`
	CurrentPlayerIdx int                      `json:"currentPlayerIdx"`
	DealerIdx        int                      `json:"dealerIdx"`
	LeadPlayerIdx    int                      `json:"leadPlayerIdx"`
	SchlagRank       int                      `json:"schlagRank"`
	CriticalSuit     int                      `json:"criticalSuit"`
	Stake            int                      `json:"stake"`
	PendingStake     int                      `json:"pendingStake"`
	RaiseCount       int                      `json:"raiseCount"`
	RaiserTeam       int                      `json:"raiserTeam"`
	ResponderIdx     int                      `json:"responderIdx"`
	CanRaise         bool                     `json:"canRaise"`
	CurrentTrick     []*WebOutputTrickCard    `json:"currentTrick"`
	TeamScores       [2]int                   `json:"teamScores"`
	TeamTricks       [2]int                   `json:"teamTricks"`
	DealWinnerTeam   int                      `json:"dealWinnerTeam"`
	GameEndFlag      bool                     `json:"gameEndFlag"`
	WinnerTeam       int                      `json:"winnerTeam"`
	Result           int                      `json:"result"`
	Hint             *WattenWebOutputHint     `json:"hint,omitempty"`
	WebOutputBase
	Config WattenWebOutputConfig `json:"config"`
}

// WattenWebOutputConfig ヴァッテン設定アウトプット
type WattenWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	TargetScore   int `json:"targetScore"`
	MaxRaises     int `json:"maxRaises"`
}

// ToConfig builds a WattenConfig from the nested web config, applying bounds checking.
func (c *WattenWebConfig) ToConfig() domain.WattenConfig {
	cfg := domain.DefaultWattenConfig()
	cfg.CpuDifficulty = domain.WattenCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.WattenCpuDifficultyEasy), int(domain.WattenCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.TargetScore, c.TargetScore, 1, 1000)
	webutil.ApplyBoundedInt(&cfg.MaxRaises, c.MaxRaises, 0, 20)
	return cfg
}

// ToConfig builds a WattenConfig from the web input.
func (p WattenWebInput) ToConfig() domain.WattenConfig {
	return configOrDefault(p.Config, (*WattenWebConfig).ToConfig, domain.DefaultWattenConfig())
}

// WattenWebController ヴァッテンWebコントローラークラス
type WattenWebController = GameWebController[usecase.WattenInteractorIF, WattenWebInput, *WattenWebOutput]

// NewWattenWebController and NewWattenWebControllerWithProvider are
// the standard and provider-backed constructors for WattenWebController.
var NewWattenWebController, NewWattenWebControllerWithProvider = webControllerPair[usecase.WattenInteractorIF, WattenWebInput, *WattenWebOutput](
	newWattenDefaultOutput, wattenDispatch,
)

func newWattenDefaultOutput(msg string) *WattenWebOutput {
	return &WattenWebOutput{
		Players:        make([]*WattenWebOutputPlayer, 0),
		CurrentTrick:   make([]*WebOutputTrickCard, 0),
		WinnerTeam:     -1,
		RaiserTeam:     -1,
		ResponderIdx:   -1,
		DealWinnerTeam: -1,
		WebOutputBase:  WebOutputBase{Message: msg},
	}
}

func wattenDispatch(bc *baseController, w http.ResponseWriter, wi usecase.WattenInteractorIF, param WattenWebInput, newDefault func(string) *WattenWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, wi.ResetWithConfig(param.ToConfig()))
	case "d", "declare":
		if !requireParam(bc, w, newDefault, param.Rank == nil || param.Suit == nil, "param error: rank and suit are required.") {
			return true
		}
		bc.writePresenterResponse(w, wi.Declare(*param.Rank, *param.Suit))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, wi.Play(*param.CardIndex))
	case "rz", "raise":
		bc.writePresenterResponse(w, wi.Raise())
	case "resp", "respond":
		if !requireParam(bc, w, newDefault, param.Hold == nil, "param error: hold is required.") {
			return true
		}
		bc.writePresenterResponse(w, wi.Respond(*param.Hold))
	case "nr", "nextround":
		bc.writePresenterResponse(w, wi.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, wi.Hint, wi.ActionLog)
	}
	return true
}
