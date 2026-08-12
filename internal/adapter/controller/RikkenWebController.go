//go:build !js || !wasm || extra2

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// RikkenWebInput リッケンWebインプット
type RikkenWebInput struct {
	BaseWebInput
	CardIndex int `json:"cardIndex,omitempty"`
	// Contract は宣言する契約。**パス (0) を送れる必要がある**のでポインタです。
	Contract *int `json:"contract,omitempty"`
	// Suit は切り札。0 は無効値なので、こちらもポインタで受けます。
	Suit   *int `json:"suit,omitempty"`
	Rounds *int `json:"rounds,omitempty"`
}

// RikkenWebOutputPlayer はリッケンWebアウトプットの席情報
type RikkenWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
	TrickCount int              `json:"trickCount"`
	Score      int              `json:"score"`
	// IsDeclarerSide は宣言側かどうか。**組は席では決まりません。**
	IsDeclarerSide bool `json:"isDeclarerSide"`
	HasPassed      bool `json:"hasPassed"`
}

// RikkenWebOutput リッケンWebアウトプット
type RikkenWebOutput struct {
	Players         []*RikkenWebOutputPlayer `json:"players"`
	Phase           int                      `json:"phase"`
	ValidPlays      []int                    `json:"validPlays"`
	DealerIdx       int                      `json:"dealerIdx"`
	Contract        int                      `json:"contract"`
	DeclarerIdx     int                      `json:"declarerIdx"`
	PartnerIdx      int                      `json:"partnerIdx"`
	CalledCard      *WebOutputCard           `json:"calledCard,omitempty"`
	TrumpSuit       int                      `json:"trumpSuit"`
	CurrentTurn     int                      `json:"currentTurn"`
	IsHumanTurn     bool                     `json:"isHumanTurn"`
	CurrentTrick    []*WebOutputTrickCard    `json:"currentTrick"`
	LastTrick       []*WebOutputTrickCard    `json:"lastTrick"`
	LastTrickWinner int                      `json:"lastTrickWinner"`
	TrickCount      int                      `json:"trickCount"`
	DeclarerTricks  int                      `json:"declarerTricks"`
	RoundNumber     int                      `json:"roundNumber"`
	GameEndFlag     bool                     `json:"gameEndFlag"`
	WinnerIdx       int                      `json:"winnerIdx"`
	Config          *RikkenWebOutConfig      `json:"config,omitempty"`
	WebOutputBase
}

// RikkenWebOutConfig はリッケンの設定
type RikkenWebOutConfig struct {
	Rounds int `json:"rounds"`
}

// RikkenWebController リッケンWebコントローラークラス
type RikkenWebController = GameWebController[usecase.RikkenInteractorIF, RikkenWebInput, *RikkenWebOutput]

// NewRikkenWebController and NewRikkenWebControllerWithProvider are
// the standard and provider-backed constructors for RikkenWebController.
var NewRikkenWebController, NewRikkenWebControllerWithProvider = webControllerPair[usecase.RikkenInteractorIF, RikkenWebInput, *RikkenWebOutput](
	newRikkenDefaultOutput, rikkenDispatch,
)

func newRikkenDefaultOutput(msg string) *RikkenWebOutput {
	return &RikkenWebOutput{
		Players:         make([]*RikkenWebOutputPlayer, 0),
		ValidPlays:      make([]int, 0),
		CurrentTrick:    make([]*WebOutputTrickCard, 0),
		LastTrick:       make([]*WebOutputTrickCard, 0),
		DeclarerIdx:     -1,
		PartnerIdx:      -1,
		LastTrickWinner: -1,
		WinnerIdx:       -1,
		WebOutputBase:   WebOutputBase{Message: msg},
	}
}

func rikkenDispatch(bc *baseController, w http.ResponseWriter, ri usecase.RikkenInteractorIF, param RikkenWebInput, newOut func(string) *RikkenWebOutput) bool {
	switch param.Command {
	case "p", "play":
		bc.writePresenterResponse(w, ri.PlayCard(param.CardIndex))
	case "b", "bid":
		// **パスは 0 という有効な値。** 省略と区別が要ります。
		if !requireParam(bc, w, newOut, param.Contract == nil, "param error: contract is required.") {
			return true
		}
		bc.writePresenterResponse(w, ri.Bid(*param.Contract))
	case "c", "call":
		if !requireParam(bc, w, newOut, param.Suit == nil, "param error: suit is required.") {
			return true
		}
		bc.writePresenterResponse(w, ri.Call(*param.Suit))
	case "next":
		bc.writePresenterResponse(w, ri.NextRound())
	case "giveup", "g":
		bc.writePresenterResponse(w, ri.GiveUp())
	case "h", "hint":
		bc.writePresenterResponse(w, ri.Hint())
	default:
		return dispatchResetAndLog(param.Command, bc, w, ri.Reset, ri.ActionLog)
	}
	return true
}
