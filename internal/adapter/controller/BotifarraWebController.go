//go:build !js || !wasm || classic

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BotifarraWebInput ボティファラWebインプット
type BotifarraWebInput struct {
	BaseWebInput
	CardIndex int `json:"cardIndex,omitempty"`
	// Suit は宣言する切り札。**切り札なし (-1) を送れる必要がある**ので
	// `omitempty` は付けません (0 = スペードも有効な値です)。
	Suit        *int `json:"suit,omitempty"`
	TargetScore *int `json:"targetScore,omitempty"`
}

// BotifarraWebOutputPlayer はボティファラWebアウトプットの席情報
type BotifarraWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	Team       int              `json:"team"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
	TrickCount int              `json:"trickCount"`
}

// BotifarraWebOutput ボティファラWebアウトプット
type BotifarraWebOutput struct {
	Players []*BotifarraWebOutputPlayer `json:"players"`
	Phase   int                         `json:"phase"`
	// ValidPlays は出せる手札の位置。**常に配列**で返します。
	ValidPlays      []int                  `json:"validPlays"`
	DealerIdx       int                    `json:"dealerIdx"`
	DeclarerIdx     int                    `json:"declarerIdx"`
	TrumpSuit       int                    `json:"trumpSuit"`
	Multiplier      int                    `json:"multiplier"`
	CurrentTurn     int                    `json:"currentTurn"`
	IsHumanTurn     bool                   `json:"isHumanTurn"`
	CurrentTrick    []*WebOutputTrickCard  `json:"currentTrick"`
	LastTrick       []*WebOutputTrickCard  `json:"lastTrick"`
	LastTrickWinner int                    `json:"lastTrickWinner"`
	TrickCount      int                    `json:"trickCount"`
	RoundPoints     []int                  `json:"roundPoints"`
	Scores          []int                  `json:"scores"`
	GameEndFlag     bool                   `json:"gameEndFlag"`
	WinnerTeam      int                    `json:"winnerTeam"`
	Config          *BotifarraWebOutConfig `json:"config,omitempty"`
	WebOutputBase
}

// BotifarraWebOutConfig はボティファラの設定
type BotifarraWebOutConfig struct {
	TargetScore   int  `json:"targetScore"`
	AllowDoubling bool `json:"allowDoubling"`
}

// BotifarraWebController ボティファラWebコントローラークラス
type BotifarraWebController = GameWebController[usecase.BotifarraInteractorIF, BotifarraWebInput, *BotifarraWebOutput]

// NewBotifarraWebController and NewBotifarraWebControllerWithProvider are
// the standard and provider-backed constructors for BotifarraWebController.
var NewBotifarraWebController, NewBotifarraWebControllerWithProvider = webControllerPair[usecase.BotifarraInteractorIF, BotifarraWebInput, *BotifarraWebOutput](
	newBotifarraDefaultOutput, botifarraDispatch,
)

func newBotifarraDefaultOutput(msg string) *BotifarraWebOutput {
	return &BotifarraWebOutput{
		Players:         make([]*BotifarraWebOutputPlayer, 0),
		ValidPlays:      make([]int, 0),
		CurrentTrick:    make([]*WebOutputTrickCard, 0),
		LastTrick:       make([]*WebOutputTrickCard, 0),
		RoundPoints:     []int{0, 0},
		Scores:          []int{0, 0},
		DeclarerIdx:     -1,
		LastTrickWinner: -1,
		WinnerTeam:      -1,
		WebOutputBase:   WebOutputBase{Message: msg},
	}
}

func botifarraDispatch(bc *baseController, w http.ResponseWriter, bi usecase.BotifarraInteractorIF, param BotifarraWebInput, _ func(string) *BotifarraWebOutput) bool {
	switch param.Command {
	case "p", "play":
		bc.writePresenterResponse(w, bi.PlayCard(param.CardIndex))
	case "declare":
		// **切り札なしは -1。** ポインタで受けるので「送らなかった」と区別できます。
		if param.Suit == nil {
			bc.writePresenterResponse(w, bi.Declare(0))
			return true
		}
		bc.writePresenterResponse(w, bi.Declare(*param.Suit))
	case "delegate":
		bc.writePresenterResponse(w, bi.Delegate())
	case "double":
		bc.writePresenterResponse(w, bi.Double())
	case "passdouble":
		bc.writePresenterResponse(w, bi.PassDouble())
	case "next":
		bc.writePresenterResponse(w, bi.NextRound())
	case "giveup", "g":
		bc.writePresenterResponse(w, bi.GiveUp())
	case "h", "hint":
		bc.writePresenterResponse(w, bi.Hint())
	default:
		return dispatchResetAndLog(param.Command, bc, w, bi.Reset, bi.ActionLog)
	}
	return true
}
