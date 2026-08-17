//go:build !js || !wasm || extra3

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// SkatWebInput Skat Web input.
type SkatWebInput struct {
	BaseWebInput
	Accept    *bool          `json:"accept,omitempty"`
	Pickup    *bool          `json:"pickup,omitempty"`
	DiscardA  *int           `json:"discardA,omitempty"`
	DiscardB  *int           `json:"discardB,omitempty"`
	GameType  *int           `json:"gameType,omitempty"`
	TrumpSuit *int           `json:"trumpSuit,omitempty"`
	CardIndex *int           `json:"cardIndex,omitempty"`
	Config    *SkatWebConfig `json:"config,omitempty"`
}

// SkatWebConfig Skat web configuration.
type SkatWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	TargetScore   *int `json:"targetScore,omitempty"`
}

// SkatWebOutputScoreBreakdown はラウンド得点がどう積み上がったか (#5561)。
type SkatWebOutputScoreBreakdown struct {
	Base       int  `json:"base"`
	Matadors   int  `json:"matadors"`
	Multiplier int  `json:"multiplier"`
	Hand       bool `json:"hand"`
	Schneider  bool `json:"schneider"`
	Schwarz    bool `json:"schwarz"`
	Doubled    bool `json:"doubled"`
	Overbid    bool `json:"overbid"`
	// Bid は Overbid のときの最終入札。Value = Bid*2 なので、これが無いと
	// クライアントは式を書けない。
	Bid   int  `json:"bid"`
	Value int  `json:"value"`
	Null  bool `json:"null"`
}

// SkatWebOutputPlayer Skat web output player.
type SkatWebOutputPlayer struct {
	ID              int              `json:"id"`
	IsHuman         bool             `json:"isHuman"`
	CardCount       int              `json:"cardCount"`
	Cards           []*WebOutputCard `json:"cards"`
	Bid             int              `json:"bid"`
	IsDeclarer      bool             `json:"isDeclarer"`
	CardPoints      int              `json:"cardPoints"`
	RoundsWon       int              `json:"roundsWon"`
	RoundsLost      int              `json:"roundsLost"`
	RoundScore      int              `json:"roundScore"`
	CumulativeScore int              `json:"cumulativeScore"`
	TrickCount      int              `json:"trickCount"`
}

// SkatWebOutputHint hint output.
type SkatWebOutputHint struct {
	CardIndex    *int   `json:"cardIndex,omitempty"`
	Bid          *int   `json:"bid,omitempty"`
	GameType     *int   `json:"gameType,omitempty"`
	TrumpSuit    *int   `json:"trumpSuit,omitempty"`
	PickSkat     *bool  `json:"pickSkat,omitempty"`
	DiscardIndex *int   `json:"discardIndex,omitempty"`
	Reason       string `json:"reason"`
}

// SkatWebOutput Skat web response payload.
type SkatWebOutput struct {
	Players             []*SkatWebOutputPlayer `json:"players"`
	Phase               int                    `json:"phase"`
	RoundNumber         int                    `json:"roundNumber"`
	TrickNumber         int                    `json:"trickNumber"`
	CurrentPlayerIdx    int                    `json:"currentPlayerIdx"`
	CurrentTrick        []*WebOutputTrickCard  `json:"currentTrick"`
	ForehandIdx         int                    `json:"forehandIdx"`
	MiddlehandIdx       int                    `json:"middlehandIdx"`
	RearhandIdx         int                    `json:"rearhandIdx"`
	DealerIdx           int                    `json:"dealerIdx"`
	DeclarerIdx         int                    `json:"declarerIdx"`
	CurrentBid          int                    `json:"currentBid"`
	ActiveBidActorIdx   int                    `json:"activeBidActorIdx"`
	GameType            int                    `json:"gameType"`
	TrumpSuit           int                    `json:"trumpSuit"`
	Skat                []*WebOutputCard       `json:"skat,omitempty"`
	OriginalSkat        []*WebOutputCard       `json:"originalSkat,omitempty"`
	PickedSkat          bool                   `json:"pickedSkat"`
	DeclarerCardPoints  int                    `json:"declarerCardPoints"`
	DefendersCardPoints int                    `json:"defendersCardPoints"`
	WinnerSide          int                    `json:"winnerSide"`
	GameValue           int                    `json:"gameValue"`
	// ScoreBreakdown はラウンド得点の内訳 (#5561)。ラウンド前は null。
	ScoreBreakdown *SkatWebOutputScoreBreakdown `json:"scoreBreakdown,omitempty"`
	GameEndFlag    bool                         `json:"gameEndFlag"`
	LeadPlayerIdx  int                          `json:"leadPlayerIdx"`
	Hint           *SkatWebOutputHint           `json:"hint,omitempty"`
	WebOutputBase
	Config SkatWebOutputConfig `json:"config"`
}

// SkatWebOutputConfig Skat configuration in the response.
type SkatWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	TargetScore   int `json:"targetScore"`
}

// ToConfig builds a SkatConfig from the nested web config, applying bounds.
func (c *SkatWebConfig) ToConfig() domain.SkatConfig {
	cfg := domain.DefaultSkatConfig()
	cfg.CpuDifficulty = domain.SkatCpuDifficulty(webutil.BoundedIntPtr(
		c.CpuDifficulty,
		int(domain.SkatCpuDifficultyEasy),
		int(domain.SkatCpuDifficultyHard),
		int(cfg.CpuDifficulty),
	))
	webutil.ApplyBoundedInt(&cfg.TargetScore, c.TargetScore, 1, 10000)
	return cfg
}

// ToConfig builds a SkatConfig from the web input.
func (p SkatWebInput) ToConfig() domain.SkatConfig {
	return configOrDefault(p.Config, (*SkatWebConfig).ToConfig, domain.DefaultSkatConfig())
}

// SkatWebController Skat web controller.
type SkatWebController = GameWebController[usecase.SkatInteractorIF, SkatWebInput, *SkatWebOutput]

// NewSkatWebController, NewSkatWebControllerWithProvider standard constructors.
var NewSkatWebController, NewSkatWebControllerWithProvider = webControllerPair[usecase.SkatInteractorIF, SkatWebInput, *SkatWebOutput](
	newSkatDefaultOutput, skatDispatch,
)

func newSkatDefaultOutput(msg string) *SkatWebOutput {
	return &SkatWebOutput{
		Players:           make([]*SkatWebOutputPlayer, 0),
		CurrentTrick:      make([]*WebOutputTrickCard, 0),
		WinnerSide:        domain.SkatWinnerUndecided,
		DeclarerIdx:       -1,
		ActiveBidActorIdx: -1,
		WebOutputBase:     WebOutputBase{Message: msg},
	}
}

func skatDispatch(bc *baseController, w http.ResponseWriter, si usecase.SkatInteractorIF, param SkatWebInput, newDefault func(string) *SkatWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, si.ResetWithConfig(param.ToConfig()))
	case "b", "bid":
		if !requireParam(bc, w, newDefault, param.Accept == nil, "param error: accept is required.") {
			return true
		}
		bc.writePresenterResponse(w, si.Bid(*param.Accept))
	case "ps", "pickskat":
		if !requireParam(bc, w, newDefault, param.Pickup == nil, "param error: pickup is required.") {
			return true
		}
		bc.writePresenterResponse(w, si.PickSkat(*param.Pickup))
	case "d", "discard":
		if !requireParam(bc, w, newDefault, param.DiscardA == nil || param.DiscardB == nil, "param error: discardA and discardB are required.") {
			return true
		}
		bc.writePresenterResponse(w, si.Discard(*param.DiscardA, *param.DiscardB))
	case "g", "game":
		if !requireParam(bc, w, newDefault, param.GameType == nil, "param error: gameType is required.") {
			return true
		}
		trump := 0
		if param.TrumpSuit != nil {
			trump = *param.TrumpSuit
		}
		bc.writePresenterResponse(w, si.DeclareGame(domain.SkatGameType(*param.GameType), trump))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, si.Play(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, si.NextTrick())
	case "nr", "nextround":
		bc.writePresenterResponse(w, si.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, si.Hint, si.ActionLog)
	}
	return true
}
