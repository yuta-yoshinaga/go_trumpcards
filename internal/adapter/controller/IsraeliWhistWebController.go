//go:build !js || !wasm || extra2

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// IsraeliWhistWebInput イスラエリホイストWebインプット
type IsraeliWhistWebInput struct {
	BaseWebInput
	CardIndex *int                   `json:"cardIndex,omitempty"`
	Suit      *int                   `json:"suit,omitempty"`
	Bid       *int                   `json:"bid,omitempty"`
	Config    *IsraeliWhistWebConfig `json:"config,omitempty"`
}

// IsraeliWhistWebConfig イスラエリホイストWeb設定
type IsraeliWhistWebConfig struct {
	Rounds *int `json:"rounds,omitempty"`
}

// IsraeliWhistWebOutputPlayer イスラエリホイストWebアウトプットプレイヤー
type IsraeliWhistWebOutputPlayer struct {
	ID        int              `json:"id"`
	IsHuman   bool             `json:"isHuman"`
	CardCount int              `json:"cardCount"`
	Cards     []*WebOutputCard `json:"cards"`
	// AuctionBid / AuctionSuit は 1 段階目、Bid は 2 段階目の宣言。
	AuctionBid  int  `json:"auctionBid"`
	AuctionSuit int  `json:"auctionSuit"`
	Passed      bool `json:"passed"`
	Bid         int  `json:"bid"`
	TrickCount  int  `json:"trickCount"`
	RoundScore  int  `json:"roundScore"`
	TotalScore  int  `json:"totalScore"`
}

// IsraeliWhistWebOutputHint ヒント出力
type IsraeliWhistWebOutputHint struct {
	CardIndex *int   `json:"cardIndex,omitempty"`
	Reason    string `json:"reason"`
	Value     int    `json:"value"`
	Suit      int    `json:"suit"`
}

// IsraeliWhistWebOutput イスラエリホイストWebアウトプット
type IsraeliWhistWebOutput struct {
	Players     []*IsraeliWhistWebOutputPlayer `json:"players"`
	Phase       int                            `json:"phase"`
	RoundNumber int                            `json:"roundNumber"`
	TrickNumber int                            `json:"trickNumber"`
	TrumpSuit   int                            `json:"trumpSuit"`
	// DeclarerIdx / HighBid / HighSuit は 1 段階目の結果。
	DeclarerIdx int `json:"declarerIdx"`
	HighBid     int `json:"highBid"`
	HighSuit    int `json:"highSuit"`
	// MinimumBid は人間が宣言できる下限（落札者のノルマ）。
	MinimumBid int `json:"minimumBid"`
	// RestrictedBid は最後の宣言者が選べない値 (-1: 制限なし)。
	RestrictedBid    int                        `json:"restrictedBid"`
	CurrentPlayerIdx int                        `json:"currentPlayerIdx"`
	AuctionPlayerIdx int                        `json:"auctionPlayerIdx"`
	BidPlayerIdx     int                        `json:"bidPlayerIdx"`
	LeadPlayerIdx    int                        `json:"leadPlayerIdx"`
	DealerIdx        int                        `json:"dealerIdx"`
	CurrentTrick     []*WebOutputTrickCard      `json:"currentTrick"`
	ValidPlays       []int                      `json:"validPlays"`
	GameEndFlag      bool                       `json:"gameEndFlag"`
	WinnerIdx        int                        `json:"winnerIdx"`
	Hint             *IsraeliWhistWebOutputHint `json:"hint,omitempty"`
	WebOutputBase
	Config IsraeliWhistWebOutputConfig `json:"config"`
}

// IsraeliWhistWebOutputConfig イスラエリホイスト設定アウトプット
type IsraeliWhistWebOutputConfig struct {
	Rounds int `json:"rounds"`
}

// ToConfig builds an IsraeliWhistConfig from the nested web config, applying bounds checking.
func (c *IsraeliWhistWebConfig) ToConfig() domain.IsraeliWhistConfig {
	cfg := domain.DefaultIsraeliWhistConfig()
	cfg.Rounds = webutil.BoundedIntPtr(c.Rounds,
		domain.IsraeliWhistRoundsMin, domain.IsraeliWhistRoundsMax, cfg.Rounds)
	return cfg
}

// ToConfig builds an IsraeliWhistConfig from the web input.
func (p IsraeliWhistWebInput) ToConfig() domain.IsraeliWhistConfig {
	return configOrDefault(p.Config, (*IsraeliWhistWebConfig).ToConfig, domain.DefaultIsraeliWhistConfig())
}

// IsraeliWhistWebController イスラエリホイストWebコントローラークラス
type IsraeliWhistWebController = GameWebController[usecase.IsraeliWhistInteractorIF, IsraeliWhistWebInput, *IsraeliWhistWebOutput]

// NewIsraeliWhistWebController and NewIsraeliWhistWebControllerWithProvider are
// the standard and provider-backed constructors for IsraeliWhistWebController.
var NewIsraeliWhistWebController, NewIsraeliWhistWebControllerWithProvider = webControllerPair[usecase.IsraeliWhistInteractorIF, IsraeliWhistWebInput, *IsraeliWhistWebOutput](
	newIsraeliWhistDefaultOutput, israeliWhistDispatch,
)

func newIsraeliWhistDefaultOutput(msg string) *IsraeliWhistWebOutput {
	return &IsraeliWhistWebOutput{
		Players:       make([]*IsraeliWhistWebOutputPlayer, 0),
		CurrentTrick:  make([]*WebOutputTrickCard, 0),
		ValidPlays:    make([]int, 0),
		DeclarerIdx:   -1,
		RestrictedBid: -1,
		WinnerIdx:     -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func israeliWhistDispatch(bc *baseController, w http.ResponseWriter, wi usecase.IsraeliWhistInteractorIF, param IsraeliWhistWebInput, newDefault func(string) *IsraeliWhistWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, wi.ResetWithConfig(param.ToConfig()))
	case "a", "auction":
		// **入札は数とスートの両方が要る。** どちらかを既定値で埋めると、
		// 出していない入札で競り落としたことになる。
		if !requireParam(bc, w, newDefault, param.Bid == nil || param.Suit == nil,
			"param error: bid and suit are required.") {
			return true
		}
		bc.writePresenterResponse(w, wi.AuctionBid(*param.Bid, *param.Suit))
	case "pass":
		bc.writePresenterResponse(w, wi.AuctionPass())
	case "b", "bid":
		// **0 は「取らない」という宣言。** 省略を 0 と読まない。
		if !requireParam(bc, w, newDefault, param.Bid == nil, "param error: bid is required.") {
			return true
		}
		bc.writePresenterResponse(w, wi.Bid(*param.Bid))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, wi.Play(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, wi.NextRound())
	case "g", "giveup":
		bc.writePresenterResponse(w, wi.GiveUp())
	default:
		return dispatchHintAndLog(param.Command, bc, w, wi.Hint, wi.ActionLog)
	}
	return true
}
