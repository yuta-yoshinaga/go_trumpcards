//go:build !js || !wasm || classic

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// EstimationWebInput エスティメーションWebインプット
type EstimationWebInput struct {
	BaseWebInput
	CardIndex *int                 `json:"cardIndex,omitempty"`
	Suit      *int                 `json:"suit,omitempty"`
	Bid       *int                 `json:"bid,omitempty"`
	Config    *EstimationWebConfig `json:"config,omitempty"`
}

// EstimationWebConfig エスティメーションWeb設定
type EstimationWebConfig struct {
	Rounds *int `json:"rounds,omitempty"`
}

// EstimationWebOutputPlayer エスティメーションWebアウトプットプレイヤー
type EstimationWebOutputPlayer struct {
	ID        int              `json:"id"`
	IsHuman   bool             `json:"isHuman"`
	CardCount int              `json:"cardCount"`
	Cards     []*WebOutputCard `json:"cards"`
	// Bid は宣言したトリック数 (-1: 未宣言)、CallType は宣言の種類。
	Bid        int `json:"bid"`
	CallType   int `json:"callType"`
	TrickCount int `json:"trickCount"`
	RoundScore int `json:"roundScore"`
	TotalScore int `json:"totalScore"`
}

// EstimationWebOutputHint ヒント出力
type EstimationWebOutputHint struct {
	CardIndex *int   `json:"cardIndex,omitempty"`
	Reason    string `json:"reason"`
	Value     int    `json:"value"`
}

// EstimationWebOutput エスティメーションWebアウトプット
type EstimationWebOutput struct {
	Players     []*EstimationWebOutputPlayer `json:"players"`
	Phase       int                          `json:"phase"`
	RoundNumber int                          `json:"roundNumber"`
	TrickNumber int                          `json:"trickNumber"`
	TrumpSuit   int                          `json:"trumpSuit"`
	// RestrictedBid は最後の宣言者が選べない値 (-1: 制限なし)。
	RestrictedBid    int                      `json:"restrictedBid"`
	CurrentPlayerIdx int                      `json:"currentPlayerIdx"`
	BidPlayerIdx     int                      `json:"bidPlayerIdx"`
	LeadPlayerIdx    int                      `json:"leadPlayerIdx"`
	DealerIdx        int                      `json:"dealerIdx"`
	CurrentTrick     []*WebOutputTrickCard    `json:"currentTrick"`
	ValidPlays       []int                    `json:"validPlays"`
	GameEndFlag      bool                     `json:"gameEndFlag"`
	WinnerIdx        int                      `json:"winnerIdx"`
	Hint             *EstimationWebOutputHint `json:"hint,omitempty"`
	WebOutputBase
	Config EstimationWebOutputConfig `json:"config"`
}

// EstimationWebOutputConfig エスティメーション設定アウトプット
type EstimationWebOutputConfig struct {
	Rounds int `json:"rounds"`
}

// ToConfig builds an EstimationConfig from the nested web config, applying bounds checking.
func (c *EstimationWebConfig) ToConfig() domain.EstimationConfig {
	cfg := domain.DefaultEstimationConfig()
	cfg.Rounds = webutil.BoundedIntPtr(c.Rounds,
		domain.EstimationRoundsMin, domain.EstimationRoundsMax, cfg.Rounds)
	return cfg
}

// ToConfig builds an EstimationConfig from the web input.
func (p EstimationWebInput) ToConfig() domain.EstimationConfig {
	return configOrDefault(p.Config, (*EstimationWebConfig).ToConfig, domain.DefaultEstimationConfig())
}

// EstimationWebController エスティメーションWebコントローラークラス
type EstimationWebController = GameWebController[usecase.EstimationInteractorIF, EstimationWebInput, *EstimationWebOutput]

// NewEstimationWebController and NewEstimationWebControllerWithProvider are
// the standard and provider-backed constructors for EstimationWebController.
var NewEstimationWebController, NewEstimationWebControllerWithProvider = webControllerPair[usecase.EstimationInteractorIF, EstimationWebInput, *EstimationWebOutput](
	newEstimationDefaultOutput, estimationDispatch,
)

func newEstimationDefaultOutput(msg string) *EstimationWebOutput {
	return &EstimationWebOutput{
		Players:       make([]*EstimationWebOutputPlayer, 0),
		CurrentTrick:  make([]*WebOutputTrickCard, 0),
		ValidPlays:    make([]int, 0),
		RestrictedBid: -1,
		WinnerIdx:     -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func estimationDispatch(bc *baseController, w http.ResponseWriter, ei usecase.EstimationInteractorIF, param EstimationWebInput, newDefault func(string) *EstimationWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, ei.ResetWithConfig(param.ToConfig()))
	case "t", "trump":
		// **切り札は既定値で埋めない。** 埋めるとプレイヤーが選んでいない
		// スートがそのラウンドの切り札になる。
		if !requireParam(bc, w, newDefault, param.Suit == nil, "param error: suit is required.") {
			return true
		}
		bc.writePresenterResponse(w, ei.SelectTrump(*param.Suit))
	case "b", "bid":
		// **宣言も既定値で埋めない。** 0 は Dash Call という別の宣言なので、
		// 省略を 0 と読むと勝手に Dash を宣言したことになる。
		if !requireParam(bc, w, newDefault, param.Bid == nil, "param error: bid is required.") {
			return true
		}
		bc.writePresenterResponse(w, ei.Bid(*param.Bid))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, ei.Play(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, ei.NextRound())
	case "g", "giveup":
		bc.writePresenterResponse(w, ei.GiveUp())
	default:
		return dispatchHintAndLog(param.Command, bc, w, ei.Hint, ei.ActionLog)
	}
	return true
}
