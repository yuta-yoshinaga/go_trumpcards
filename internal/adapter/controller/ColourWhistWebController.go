//go:build !js || !wasm || extra4

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// ColourWhistWebInput カラーホイストWebインプット
type ColourWhistWebInput struct {
	BaseWebInput
	CardIndex int `json:"cardIndex,omitempty"`
	// Contract は宣言する契約。**パス (0) を送れる必要がある**のでポインタです。
	Contract *int `json:"contract,omitempty"`
	// Suit は切り札。0 は無効値なのでこちらもポインタで受けます。
	Suit   *int `json:"suit,omitempty"`
	Rounds *int `json:"rounds,omitempty"`
}

// ColourWhistWebOutputPlayer はカラーホイストWebアウトプットの席情報
type ColourWhistWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
	TrickCount int              `json:"trickCount"`
	Score      int              `json:"score"`
	// IsDeclarerSide は契約側かどうか。**組は席では決まりません。**
	IsDeclarerSide bool `json:"isDeclarerSide"`
	HasPassed      bool `json:"hasPassed"`
}

// ColourWhistWebOutput カラーホイストWebアウトプット
type ColourWhistWebOutput struct {
	Players     []*ColourWhistWebOutputPlayer `json:"players"`
	Phase       int                           `json:"phase"`
	ValidPlays  []int                         `json:"validPlays"`
	DealerIdx   int                           `json:"dealerIdx"`
	Contract    int                           `json:"contract"`
	DeclarerIdx int                           `json:"declarerIdx"`
	PartnerIdx  int                           `json:"partnerIdx"`
	CalledCard  *WebOutputCard                `json:"calledCard,omitempty"`
	TrumpSuit   int                           `json:"trumpSuit"`
	// TroelForced は配りで Troel が強制成立したか。
	TroelForced     bool                  `json:"troelForced"`
	CurrentTurn     int                   `json:"currentTurn"`
	IsHumanTurn     bool                  `json:"isHumanTurn"`
	CurrentTrick    []*WebOutputTrickCard `json:"currentTrick"`
	LastTrick       []*WebOutputTrickCard `json:"lastTrick"`
	LastTrickWinner int                   `json:"lastTrickWinner"`
	TrickCount      int                   `json:"trickCount"`
	DeclarerTricks  int                   `json:"declarerTricks"`
	RoundNumber     int                   `json:"roundNumber"`
	GameEndFlag     bool                  `json:"gameEndFlag"`
	WinnerIdx       int                   `json:"winnerIdx"`
	Config          *ColourWhistWebOutCfg `json:"config,omitempty"`
	WebOutputBase
}

// ColourWhistWebOutCfg はカラーホイストの設定
type ColourWhistWebOutCfg struct {
	Rounds int `json:"rounds"`
}

// ColourWhistWebController カラーホイストWebコントローラークラス
type ColourWhistWebController = GameWebController[usecase.ColourWhistInteractorIF, ColourWhistWebInput, *ColourWhistWebOutput]

// NewColourWhistWebController and NewColourWhistWebControllerWithProvider are
// the standard and provider-backed constructors for ColourWhistWebController.
var NewColourWhistWebController, NewColourWhistWebControllerWithProvider = webControllerPair[usecase.ColourWhistInteractorIF, ColourWhistWebInput, *ColourWhistWebOutput](
	newColourWhistDefaultOutput, colourWhistDispatch,
)

func newColourWhistDefaultOutput(msg string) *ColourWhistWebOutput {
	return &ColourWhistWebOutput{
		Players:         make([]*ColourWhistWebOutputPlayer, 0),
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

func colourWhistDispatch(bc *baseController, w http.ResponseWriter, ci usecase.ColourWhistInteractorIF, param ColourWhistWebInput, newOut func(string) *ColourWhistWebOutput) bool {
	switch param.Command {
	case "p", "play":
		bc.writePresenterResponse(w, ci.PlayCard(param.CardIndex))
	case "b", "bid":
		// **パスは 0 という有効な値。** 省略と区別が要ります。
		if !requireParam(bc, w, newOut, param.Contract == nil, "param error: contract is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.Bid(*param.Contract))
	case "c", "call":
		if !requireParam(bc, w, newOut, param.Suit == nil, "param error: suit is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.Call(*param.Suit))
	case "next":
		bc.writePresenterResponse(w, ci.NextRound())
	case "giveup", "g":
		bc.writePresenterResponse(w, ci.GiveUp())
	case "h", "hint":
		bc.writePresenterResponse(w, ci.Hint())
	default:
		return dispatchResetAndLog(param.Command, bc, w, ci.Reset, ci.ActionLog)
	}
	return true
}
