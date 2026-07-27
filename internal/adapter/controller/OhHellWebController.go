package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// OhHellWebInput オー・ヘルWebインプット
type OhHellWebInput struct {
	BaseWebInput
	Bid       *int             `json:"bid,omitempty"`
	CardIndex *int             `json:"cardIndex,omitempty"`
	Config    *OhHellWebConfig `json:"config,omitempty"`
}

// OhHellWebConfig オー・ヘルWeb設定
type OhHellWebConfig struct {
	CpuDifficulty  *int `json:"cpuDifficulty,omitempty"`
	MaxHandSize    *int `json:"maxHandSize,omitempty"`
	ScoringVariant *int `json:"scoringVariant,omitempty"`
	RoundDirection *int `json:"roundDirection,omitempty"`
}

// OhHellWebOutputPlayer オー・ヘルWebアウトプットプレイヤー
type OhHellWebOutputPlayer struct {
	ID              int              `json:"id"`
	IsHuman         bool             `json:"isHuman"`
	CardCount       int              `json:"cardCount"`
	Cards           []*WebOutputCard `json:"cards"`
	Bid             int              `json:"bid"`
	RoundScore      int              `json:"roundScore"`
	CumulativeScore int              `json:"cumulativeScore"`
	TrickCount      int              `json:"trickCount"`
}

// OhHellWebOutputHint ヒント出力
type OhHellWebOutputHint struct {
	CardIndex *int   `json:"cardIndex,omitempty"`
	Bid       *int   `json:"bid,omitempty"`
	Reason    string `json:"reason"`
}

// OhHellWebOutput オー・ヘルWebアウトプット
type OhHellWebOutput struct {
	Players          []*OhHellWebOutputPlayer `json:"players"`
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
	Hint             *OhHellWebOutputHint     `json:"hint,omitempty"`
	WebOutputBase
	Config OhHellWebOutputConfig `json:"config"`
}

// OhHellWebOutputConfig オー・ヘル設定アウトプット
type OhHellWebOutputConfig struct {
	CpuDifficulty  int `json:"cpuDifficulty"`
	MaxHandSize    int `json:"maxHandSize"`
	ScoringVariant int `json:"scoringVariant"`
	RoundDirection int `json:"roundDirection"`
}

// ToConfig builds an OhHellConfig from the nested web config, applying bounds checking.
func (c *OhHellWebConfig) ToConfig() domain.OhHellConfig {
	cfg := domain.DefaultOhHellConfig()
	cfg.CpuDifficulty = domain.OhHellCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.OhHellCpuDifficultyEasy), int(domain.OhHellCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.MaxHandSize, c.MaxHandSize, 1, 13)
	cfg.ScoringVariant = domain.OhHellScoringVariant(webutil.BoundedIntPtr(c.ScoringVariant, int(domain.OhHellScoringStandard), int(domain.OhHellScoringPenalty), int(cfg.ScoringVariant)))
	cfg.RoundDirection = domain.OhHellRoundDirection(webutil.BoundedIntPtr(c.RoundDirection, int(domain.OhHellRoundDownOnly), int(domain.OhHellRoundDownAndUp), int(cfg.RoundDirection)))
	return cfg
}

// ToConfig builds an OhHellConfig from the web input.
func (p OhHellWebInput) ToConfig() domain.OhHellConfig {
	return configOrDefault(p.Config, (*OhHellWebConfig).ToConfig, domain.DefaultOhHellConfig())
}

// OhHellWebController オー・ヘルWebコントローラークラス
type OhHellWebController = GameWebController[usecase.OhHellInteractorIF, OhHellWebInput, *OhHellWebOutput]

// NewOhHellWebController and NewOhHellWebControllerWithProvider are
// the standard and provider-backed constructors for OhHellWebController.
var NewOhHellWebController, NewOhHellWebControllerWithProvider = webControllerPair[usecase.OhHellInteractorIF, OhHellWebInput, *OhHellWebOutput](
	newOhHellDefaultOutput, ohHellDispatch,
)

func newOhHellDefaultOutput(msg string) *OhHellWebOutput {
	return &OhHellWebOutput{
		Players:       make([]*OhHellWebOutputPlayer, 0),
		CurrentTrick:  make([]*WebOutputTrickCard, 0),
		WinnerIdx:     -1,
		TrumpSuit:     -1,
		RestrictedBid: -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func ohHellDispatch(bc *baseController, w http.ResponseWriter, oi usecase.OhHellInteractorIF, param OhHellWebInput, newDefault func(string) *OhHellWebOutput) bool {
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
