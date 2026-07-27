//go:build !js || !wasm || extra

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// RookWebInput ルーク(Rook) Webインプット
type RookWebInput struct {
	BaseWebInput
	Bid            *int           `json:"bid,omitempty"`
	DiscardIndices []int          `json:"discardIndices,omitempty"`
	TrumpColor     *int           `json:"trumpColor,omitempty"`
	CardIndex      *int           `json:"cardIndex,omitempty"`
	Config         *RookWebConfig `json:"config,omitempty"`
}

// RookWebConfig ルーク(Rook) Web設定
type RookWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	TargetScore   *int `json:"targetScore,omitempty"`
}

// RookWebOutputPlayer ルーク(Rook) Webアウトプットプレイヤー
type RookWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
	Team       int              `json:"team"`
	TrickCount int              `json:"trickCount"`
	Points     int              `json:"points"`
	Bid        int              `json:"bid"`
	Passed     bool             `json:"passed"`
	IsDeclarer bool             `json:"isDeclarer"`
}

// RookWebOutputHint ヒント出力
type RookWebOutputHint struct {
	Bid            *int   `json:"bid,omitempty"`
	Pass           *bool  `json:"pass,omitempty"`
	DiscardIndices []int  `json:"discardIndices,omitempty"`
	TrumpColor     *int   `json:"trumpColor,omitempty"`
	CardIndex      *int   `json:"cardIndex,omitempty"`
	Reason         string `json:"reason"`
}

// RookWebOutput ルーク(Rook) Webアウトプット
type RookWebOutput struct {
	Players          []*RookWebOutputPlayer `json:"players"`
	Phase            int                    `json:"phase"`
	RoundNumber      int                    `json:"roundNumber"`
	TrickNumber      int                    `json:"trickNumber"`
	CurrentPlayerIdx int                    `json:"currentPlayerIdx"`
	BidPlayerIdx     int                    `json:"bidPlayerIdx"`
	DealerIdx        int                    `json:"dealerIdx"`
	LeadPlayerIdx    int                    `json:"leadPlayerIdx"`
	TrumpColor       int                    `json:"trumpColor"`
	ContractBid      int                    `json:"contractBid"`
	DeclarerIdx      int                    `json:"declarerIdx"`
	HighestBid       int                    `json:"highestBid"`
	HighestBidder    int                    `json:"highestBidder"`
	NestCount        int                    `json:"nestCount"`
	Nest             []*WebOutputCard       `json:"nest"`
	CurrentTrick     []*WebOutputTrickCard  `json:"currentTrick"`
	TeamScores       [2]int                 `json:"teamScores"`
	TeamPoints       [2]int                 `json:"teamPoints"`
	GameEndFlag      bool                   `json:"gameEndFlag"`
	WinnerTeam       int                    `json:"winnerTeam"`
	Hint             *RookWebOutputHint     `json:"hint,omitempty"`
	WebOutputBase
	Config RookWebOutputConfig `json:"config"`
}

// RookWebOutputConfig ルーク(Rook) 設定アウトプット
type RookWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	TargetScore   int `json:"targetScore"`
}

// ToConfig builds a RookConfig from the nested web config, applying bounds checking.
func (c *RookWebConfig) ToConfig() domain.RookConfig {
	cfg := domain.DefaultRookConfig()
	cfg.CpuDifficulty = domain.RookCpuDifficulty(webutil.BoundedIntPtr(
		c.CpuDifficulty,
		int(domain.RookCpuDifficultyEasy),
		int(domain.RookCpuDifficultyHard),
		int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.TargetScore, c.TargetScore, 1, 100000)
	return cfg
}

// ToConfig builds a RookConfig from the web input.
func (p RookWebInput) ToConfig() domain.RookConfig {
	return configOrDefault(p.Config, (*RookWebConfig).ToConfig, domain.DefaultRookConfig())
}

// RookWebController ルーク(Rook) Webコントローラークラス
type RookWebController = GameWebController[usecase.RookInteractorIF, RookWebInput, *RookWebOutput]

// NewRookWebController and NewRookWebControllerWithProvider are the standard
// and provider-backed constructors for RookWebController.
var NewRookWebController, NewRookWebControllerWithProvider = webControllerPair[usecase.RookInteractorIF, RookWebInput, *RookWebOutput](
	newRookDefaultOutput, rookDispatch,
)

func newRookDefaultOutput(msg string) *RookWebOutput {
	return &RookWebOutput{
		Players:       make([]*RookWebOutputPlayer, 0),
		CurrentTrick:  make([]*WebOutputTrickCard, 0),
		Nest:          make([]*WebOutputCard, 0),
		WinnerTeam:    -1,
		DeclarerIdx:   -1,
		HighestBidder: -1,
		TrumpColor:    -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func rookDispatch(bc *baseController, w http.ResponseWriter, fi usecase.RookInteractorIF, param RookWebInput, newDefault func(string) *RookWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, fi.ResetWithConfig(param.ToConfig()))
	case "b", "bid":
		if !requireParam(bc, w, newDefault, param.Bid == nil, "param error: bid is required.") {
			return true
		}
		bc.writePresenterResponse(w, fi.Bid(*param.Bid))
	case "pa", "pass":
		bc.writePresenterResponse(w, fi.Pass())
	case "e", "exchange":
		if !requireParam(bc, w, newDefault, len(param.DiscardIndices) == 0, "param error: discardIndices is required.") {
			return true
		}
		if !requireParam(bc, w, newDefault, param.TrumpColor == nil, "param error: trumpColor is required.") {
			return true
		}
		bc.writePresenterResponse(w, fi.ExchangeNest(param.DiscardIndices, *param.TrumpColor))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, fi.Play(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, fi.NextTrick())
	case "nr", "nextround":
		bc.writePresenterResponse(w, fi.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, fi.Hint, fi.ActionLog)
	}
	return true
}
