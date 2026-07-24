//go:build !js || !wasm || solo

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BidWhistWebInput Bid Whist Webインプット
type BidWhistWebInput struct {
	BaseWebInput
	BidTricks      *int               `json:"bidTricks,omitempty"`
	BidDirection   *int               `json:"bidDirection,omitempty"`
	TrumpSuit      *int               `json:"trumpSuit,omitempty"`
	DiscardIndices []int              `json:"discardIndices,omitempty"`
	CardIndex      *int               `json:"cardIndex,omitempty"`
	Config         *BidWhistWebConfig `json:"config,omitempty"`
}

// BidWhistWebConfig Bid Whist Web設定
type BidWhistWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	TargetScore   *int `json:"targetScore,omitempty"`
}

// BidWhistWebOutputBid ビッド情報
type BidWhistWebOutputBid struct {
	Tricks    int `json:"tricks"`
	Direction int `json:"direction"`
}

// BidWhistWebOutputPlayer Bid Whist Webアウトプットプレイヤー
type BidWhistWebOutputPlayer struct {
	ID         int                   `json:"id"`
	IsHuman    bool                  `json:"isHuman"`
	CardCount  int                   `json:"cardCount"`
	Cards      []*WebOutputCard      `json:"cards"`
	Team       int                   `json:"team"`
	TrickCount int                   `json:"trickCount"`
	Bid        *BidWhistWebOutputBid `json:"bid,omitempty"`
	Passed     bool                  `json:"passed"`
	IsDeclarer bool                  `json:"isDeclarer"`
}

// BidWhistWebOutputTrickCard トリック中の1枚
type BidWhistWebOutputTrickCard struct {
	PlayerIdx int            `json:"playerIdx"`
	Card      *WebOutputCard `json:"card"`
}

// BidWhistWebOutputHint ヒント出力
type BidWhistWebOutputHint struct {
	BidTricks      *int   `json:"bidTricks,omitempty"`
	BidDirection   *int   `json:"bidDirection,omitempty"`
	Pass           *bool  `json:"pass,omitempty"`
	TrumpSuit      *int   `json:"trumpSuit,omitempty"`
	DiscardIndices []int  `json:"discardIndices,omitempty"`
	CardIndex      *int   `json:"cardIndex,omitempty"`
	Reason         string `json:"reason"`
}

// BidWhistWebOutput Bid Whist Webアウトプット
type BidWhistWebOutput struct {
	Players           []*BidWhistWebOutputPlayer    `json:"players"`
	Phase             int                           `json:"phase"`
	RoundNumber       int                           `json:"roundNumber"`
	TrickNumber       int                           `json:"trickNumber"`
	CurrentPlayerIdx  int                           `json:"currentPlayerIdx"`
	BidPlayerIdx      int                           `json:"bidPlayerIdx"`
	DealerIdx         int                           `json:"dealerIdx"`
	LeadPlayerIdx     int                           `json:"leadPlayerIdx"`
	TrumpSuit         int                           `json:"trumpSuit"`
	ContractTricks    int                           `json:"contractTricks"`
	ContractDirection int                           `json:"contractDirection"`
	DeclarerIdx       int                           `json:"declarerIdx"`
	HighestBid        *BidWhistWebOutputBid         `json:"highestBid,omitempty"`
	HighestBidder     int                           `json:"highestBidder"`
	KittyCount        int                           `json:"kittyCount"`
	KittyIndices      []int                         `json:"kittyIndices"`
	CurrentTrick      []*BidWhistWebOutputTrickCard `json:"currentTrick"`
	TeamScores        [2]int                        `json:"teamScores"`
	GameEndFlag       bool                          `json:"gameEndFlag"`
	WinnerTeam        int                           `json:"winnerTeam"`
	Hint              *BidWhistWebOutputHint        `json:"hint,omitempty"`
	WebOutputBase
	Config BidWhistWebOutputConfig `json:"config"`
}

// BidWhistWebOutputConfig Bid Whist 設定アウトプット
type BidWhistWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	TargetScore   int `json:"targetScore"`
}

// ToConfig builds a BidWhistConfig from the nested web config, applying bounds checking.
func (c *BidWhistWebConfig) ToConfig() domain.BidWhistConfig {
	cfg := domain.DefaultBidWhistConfig()
	cfg.CpuDifficulty = domain.BidWhistCpuDifficulty(webutil.BoundedIntPtr(
		c.CpuDifficulty,
		int(domain.BidWhistCpuDifficultyEasy),
		int(domain.BidWhistCpuDifficultyHard),
		int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.TargetScore, c.TargetScore, 1, 100000)
	return cfg
}

// ToConfig builds a BidWhistConfig from the web input.
func (p BidWhistWebInput) ToConfig() domain.BidWhistConfig {
	return configOrDefault(p.Config, (*BidWhistWebConfig).ToConfig, domain.DefaultBidWhistConfig())
}

// BidWhistWebController Bid Whist Webコントローラークラス
type BidWhistWebController = GameWebController[usecase.BidWhistInteractorIF, BidWhistWebInput, *BidWhistWebOutput]

// NewBidWhistWebController and NewBidWhistWebControllerWithProvider are the
// standard and provider-backed constructors for BidWhistWebController.
var NewBidWhistWebController, NewBidWhistWebControllerWithProvider = webControllerPair[usecase.BidWhistInteractorIF, BidWhistWebInput, *BidWhistWebOutput](
	newBidWhistDefaultOutput, bidWhistDispatch,
)

func newBidWhistDefaultOutput(msg string) *BidWhistWebOutput {
	return &BidWhistWebOutput{
		Players:       make([]*BidWhistWebOutputPlayer, 0),
		CurrentTrick:  make([]*BidWhistWebOutputTrickCard, 0),
		KittyIndices:  make([]int, 0),
		WinnerTeam:    -1,
		DeclarerIdx:   -1,
		HighestBidder: -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func bidWhistDispatch(bc *baseController, w http.ResponseWriter, bi usecase.BidWhistInteractorIF, param BidWhistWebInput, newDefault func(string) *BidWhistWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, bi.ResetWithConfig(param.ToConfig()))
	case "b", "bid":
		if !requireParam(bc, w, newDefault, param.BidTricks == nil || param.BidDirection == nil,
			"param error: bidTricks and bidDirection are required.") {
			return true
		}
		bc.writePresenterResponse(w, bi.Bid(*param.BidTricks, *param.BidDirection))
	case "pa", "pass":
		bc.writePresenterResponse(w, bi.Pass())
	case "t", "trump":
		if !requireParam(bc, w, newDefault, param.TrumpSuit == nil, "param error: trumpSuit is required.") {
			return true
		}
		bc.writePresenterResponse(w, bi.DeclareTrump(*param.TrumpSuit))
	case "e", "exchange":
		if !requireParam(bc, w, newDefault, len(param.DiscardIndices) != domain.BidWhistKittySize,
			"param error: exactly 6 discardIndices are required.") {
			return true
		}
		bc.writePresenterResponse(w, bi.ExchangeKitty(param.DiscardIndices))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, bi.Play(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, bi.NextTrick())
	case "nr", "nextround":
		bc.writePresenterResponse(w, bi.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, bi.Hint, bi.ActionLog)
	}
	return true
}
