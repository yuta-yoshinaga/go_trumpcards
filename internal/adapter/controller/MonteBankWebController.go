//go:build !js || !wasm || casino

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// MonteBankWebInput モンテバンクWebインプット
type MonteBankWebInput struct {
	BaseWebInput
	// Idx は賭ける場札の位置。**0 が正当な値**なので、省略と区別する必要がある。
	Idx *int `json:"idx,omitempty"`
	// Bet は賭け金。
	Bet *int `json:"bet,omitempty"`
}

// MonteBankWebOutCfg はモンテバンクの設定
type MonteBankWebOutCfg struct {
	InitialChips int `json:"initialChips"`
	DefaultBet   int `json:"defaultBet"`
}

// MonteBankWebOutputCard は場札 1 枚とその賭けやすさ
type MonteBankWebOutputCard struct {
	Card *WebOutputCard `json:"card"`
	// SuitCount はこのスートが場札に何枚出ているか。
	//
	// **これが賭けの良し悪しを決める唯一の数字。** 1 なら互角、2 以上なら
	// 賭けるだけ損。ページに数え直させない。
	SuitCount int `json:"suitCount"`
	// RemainingOfSuit は山に残っている同スートの枚数。
	RemainingOfSuit int `json:"remainingOfSuit"`
	// IsEven は 3:1 でちょうど互角の賭けか (SuitCount == 1)。
	IsEven bool `json:"isEven"`
	// IsPicked は賭けた札か。
	IsPicked bool `json:"isPicked"`
}

// MonteBankWebOutput モンテバンクWebアウトプット
type MonteBankWebOutput struct {
	Phase  int                       `json:"phase"`
	Layout []*MonteBankWebOutputCard `json:"layout"`
	// Gate はめくった 1 枚。賭ける前は null。
	Gate *WebOutputCard `json:"gate,omitempty"`
	// Pick は賭けた場札の位置。賭ける前は -1。
	Pick int `json:"pick"`
	Bet  int `json:"bet"`
	// Result は 0=none, 1=win, 2=lose。
	Result         int  `json:"result"`
	Payout         int  `json:"payout"`
	Chips          int  `json:"chips"`
	RoundNumber    int  `json:"roundNumber"`
	RemainingCards int  `json:"remainingCards"`
	GameEndFlag    bool `json:"gameEndFlag"`
	// PayoutMultiplier は的中したときの倍率 (3)。
	PayoutMultiplier int `json:"payoutMultiplier"`

	Config *MonteBankWebOutCfg `json:"config,omitempty"`
	WebOutputBase
}

// MonteBankWebController モンテバンクWebコントローラークラス
type MonteBankWebController = GameWebController[usecase.MonteBankInteractorIF, MonteBankWebInput, *MonteBankWebOutput]

// NewMonteBankWebController and NewMonteBankWebControllerWithProvider
// are the standard and provider-backed constructors.
var NewMonteBankWebController, NewMonteBankWebControllerWithProvider = webControllerPair[usecase.MonteBankInteractorIF, MonteBankWebInput, *MonteBankWebOutput](
	newMonteBankDefaultOutput, monteBankDispatch,
)

func newMonteBankDefaultOutput(msg string) *MonteBankWebOutput {
	return &MonteBankWebOutput{
		Layout:        make([]*MonteBankWebOutputCard, 0),
		Pick:          -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func monteBankDispatch(bc *baseController, w http.ResponseWriter, ci usecase.MonteBankInteractorIF, param MonteBankWebInput, newOut func(string) *MonteBankWebOutput) bool {
	switch param.Command {
	case "b", "bet":
		// **場札 0 に賭けるのは普通の操作。** 省略と同一視すると、いちばん左の
		// 札を選んだリクエストが全部 400 になる。
		if !requireParam(bc, w, newOut, param.Idx == nil, "param error: idx is required.") {
			return true
		}
		if !requireParam(bc, w, newOut, param.Bet == nil, "param error: bet is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.PlaceBet(*param.Idx, *param.Bet))
	case "next":
		bc.writePresenterResponse(w, ci.NextRound())
	case "hint":
		bc.writePresenterResponse(w, ci.Hint())
	default:
		return dispatchResetAndLog(param.Command, bc, w, ci.Reset, ci.ActionLog)
	}
	return true
}
