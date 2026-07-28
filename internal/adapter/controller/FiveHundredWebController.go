//go:build !js || !wasm || solo

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// FiveHundredWebInput 500 Webインプット
type FiveHundredWebInput struct {
	BaseWebInput
	BidKind        *int                  `json:"bidKind,omitempty"`
	BidTricks      *int                  `json:"bidTricks,omitempty"`
	BidSuit        *int                  `json:"bidSuit,omitempty"`
	DiscardIndices []int                 `json:"discardIndices,omitempty"`
	CardIndex      *int                  `json:"cardIndex,omitempty"`
	JokerSuit      *int                  `json:"jokerSuit,omitempty"`
	Config         *FiveHundredWebConfig `json:"config,omitempty"`
}

// FiveHundredWebConfig 500 Web設定
type FiveHundredWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	TargetScore   *int `json:"targetScore,omitempty"`
}

// FiveHundredWebOutputBid ビッド情報
type FiveHundredWebOutputBid struct {
	Kind   int `json:"kind"`
	Tricks int `json:"tricks"`
	Suit   int `json:"suit"`
	Value  int `json:"value"`
}

// FiveHundredWebOutputPlayer 500 Webアウトプットプレイヤー
type FiveHundredWebOutputPlayer struct {
	ID         int                      `json:"id"`
	IsHuman    bool                     `json:"isHuman"`
	CardCount  int                      `json:"cardCount"`
	Cards      []*WebOutputCard         `json:"cards"`
	Team       int                      `json:"team"`
	TrickCount int                      `json:"trickCount"`
	Bid        *FiveHundredWebOutputBid `json:"bid,omitempty"`
	Passed     bool                     `json:"passed"`
	IsDeclarer bool                     `json:"isDeclarer"`
}

// FiveHundredWebOutputHint ヒント出力
type FiveHundredWebOutputHint struct {
	BidKind        *int   `json:"bidKind,omitempty"`
	BidTricks      *int   `json:"bidTricks,omitempty"`
	BidSuit        *int   `json:"bidSuit,omitempty"`
	Pass           *bool  `json:"pass,omitempty"`
	DiscardIndices []int  `json:"discardIndices,omitempty"`
	CardIndex      *int   `json:"cardIndex,omitempty"`
	JokerSuit      *int   `json:"jokerSuit,omitempty"`
	Reason         string `json:"reason"`
}

// FiveHundredWebOutput 500 Webアウトプット
type FiveHundredWebOutput struct {
	Players          []*FiveHundredWebOutputPlayer `json:"players"`
	Phase            int                           `json:"phase"`
	RoundNumber      int                           `json:"roundNumber"`
	TrickNumber      int                           `json:"trickNumber"`
	CurrentPlayerIdx int                           `json:"currentPlayerIdx"`
	BidPlayerIdx     int                           `json:"bidPlayerIdx"`
	DealerIdx        int                           `json:"dealerIdx"`
	LeadPlayerIdx    int                           `json:"leadPlayerIdx"`
	TrumpSuit        int                           `json:"trumpSuit"`
	ContractKind     int                           `json:"contractKind"`
	ContractTricks   int                           `json:"contractTricks"`
	ContractValue    int                           `json:"contractValue"`
	DeclarerIdx      int                           `json:"declarerIdx"`
	HighestBid       *FiveHundredWebOutputBid      `json:"highestBid,omitempty"`
	HighestBidder    int                           `json:"highestBidder"`
	JokerLeadSuit    int                           `json:"jokerLeadSuit"`
	KittyCount       int                           `json:"kittyCount"`
	CurrentTrick     []*WebOutputTrickCard         `json:"currentTrick"`
	TeamScores       [2]int                        `json:"teamScores"`
	GameEndFlag      bool                          `json:"gameEndFlag"`
	WinnerTeam       int                           `json:"winnerTeam"`
	Hint             *FiveHundredWebOutputHint     `json:"hint,omitempty"`
	WebOutputBase
	Config FiveHundredWebOutputConfig `json:"config"`
}

// FiveHundredWebOutputConfig 500 設定アウトプット
type FiveHundredWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	TargetScore   int `json:"targetScore"`
}

// ToConfig builds a FiveHundredConfig from the nested web config, applying bounds checking.
func (c *FiveHundredWebConfig) ToConfig() domain.FiveHundredConfig {
	cfg := domain.DefaultFiveHundredConfig()
	cfg.CpuDifficulty = domain.FiveHundredCpuDifficulty(webutil.BoundedIntPtr(
		c.CpuDifficulty,
		int(domain.FiveHundredCpuDifficultyEasy),
		int(domain.FiveHundredCpuDifficultyHard),
		int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.TargetScore, c.TargetScore, 1, 100000)
	return cfg
}

// ToConfig builds a FiveHundredConfig from the web input.
func (p FiveHundredWebInput) ToConfig() domain.FiveHundredConfig {
	return configOrDefault(p.Config, (*FiveHundredWebConfig).ToConfig, domain.DefaultFiveHundredConfig())
}

// FiveHundredWebController 500 Webコントローラークラス
type FiveHundredWebController = GameWebController[usecase.FiveHundredInteractorIF, FiveHundredWebInput, *FiveHundredWebOutput]

// NewFiveHundredWebController and NewFiveHundredWebControllerWithProvider are
// the standard and provider-backed constructors for FiveHundredWebController.
var NewFiveHundredWebController, NewFiveHundredWebControllerWithProvider = webControllerPair[usecase.FiveHundredInteractorIF, FiveHundredWebInput, *FiveHundredWebOutput](
	newFiveHundredDefaultOutput, fiveHundredDispatch,
)

func newFiveHundredDefaultOutput(msg string) *FiveHundredWebOutput {
	return &FiveHundredWebOutput{
		Players:       make([]*FiveHundredWebOutputPlayer, 0),
		CurrentTrick:  make([]*WebOutputTrickCard, 0),
		WinnerTeam:    -1,
		DeclarerIdx:   -1,
		HighestBidder: -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func fiveHundredDispatch(bc *baseController, w http.ResponseWriter, fi usecase.FiveHundredInteractorIF, param FiveHundredWebInput, newDefault func(string) *FiveHundredWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, fi.ResetWithConfig(param.ToConfig()))
	case "b", "bid":
		if !requireParam(bc, w, newDefault, param.BidKind == nil, "param error: bidKind is required.") {
			return true
		}
		tricks := 0
		if param.BidTricks != nil {
			tricks = *param.BidTricks
		}
		suit := -1
		if param.BidSuit != nil {
			suit = *param.BidSuit
		}
		bc.writePresenterResponse(w, fi.Bid(domain.FiveHundredContractKind(*param.BidKind), tricks, suit))
	case "pa", "pass":
		bc.writePresenterResponse(w, fi.Pass())
	case "e", "exchange":
		if !requireParam(bc, w, newDefault, len(param.DiscardIndices) == 0, "param error: discardIndices is required.") {
			return true
		}
		bc.writePresenterResponse(w, fi.ExchangeKitty(param.DiscardIndices))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		jokerSuit := -1
		if param.JokerSuit != nil {
			jokerSuit = *param.JokerSuit
		}
		bc.writePresenterResponse(w, fi.Play(*param.CardIndex, jokerSuit))
	case "n", "next":
		bc.writePresenterResponse(w, fi.NextTrick())
	case "nr", "nextround":
		bc.writePresenterResponse(w, fi.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, fi.Hint, fi.ActionLog)
	}
	return true
}
