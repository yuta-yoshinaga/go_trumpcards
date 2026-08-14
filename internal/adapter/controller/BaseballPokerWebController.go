//go:build !js || !wasm || casino

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BaseballPokerWebInput ベースボールポーカーWebインプット
type BaseballPokerWebInput struct {
	BaseWebInput
	// Amount はベット / レイズの額。
	Amount *int `json:"amount,omitempty"`
}

// BaseballPokerWebOutCfg はベースボールポーカーの設定
type BaseballPokerWebOutCfg struct {
	Seats        int `json:"seats"`
	InitialChips int `json:"initialChips"`
	Ante         int `json:"ante"`
}

// BaseballPokerWebOutputSeat は 1 席の状態
type BaseballPokerWebOutputSeat struct {
	Name    string `json:"name"`
	IsHuman bool   `json:"isHuman"`
	Chips   int    `json:"chips"`
	Bet     int    `json:"bet"`
	// Cards は手札。**見えない伏せ札は null で埋める** ── 詰めると表札の
	// 並び順が崩れ、どの札が公開されたのか画面が復元できない。
	Cards []*WebOutputCard `json:"cards"`
	// FaceUp は Cards と同じ長さで、その位置が表向きかを持つ。
	FaceUp []bool `json:"faceUp"`
	// BonusCards は表の 4 で受け取った追加札の枚数。
	BonusCards int  `json:"bonusCards"`
	Folded     bool `json:"folded"`
	AllIn      bool `json:"allIn"`
	IsTurn     bool `json:"isTurn"`
	// IsBuying はこの席が買い増しを迫られているか。
	IsBuying bool `json:"isBuying"`
	// HandRank と BestHand と UsedWild はショーダウン後のみ。
	HandRank  int              `json:"handRank"`
	UsedWild  bool             `json:"usedWild"`
	BestHand  []*WebOutputCard `json:"bestHand"`
	WonAmount int              `json:"wonAmount"`
}

// BaseballPokerWebOutput ベースボールポーカーWebアウトプット
type BaseballPokerWebOutput struct {
	Phase int                           `json:"phase"`
	Seats []*BaseballPokerWebOutputSeat `json:"seats"`
	// Street は配り終えた表札の数 (1..4)。
	Street int `json:"street"`
	// StreetTotal は表札の総数 (4)。残りのベットラウンド数が読める。
	StreetTotal int `json:"streetTotal"`
	// WildValues はワイルドの値 (3 と 9)。**画面に書き写させない。**
	WildValues []int `json:"wildValues"`
	// BonusValue は追加札をくれる表札の値 (4)。
	BonusValue int `json:"bonusValue"`
	// BuyInValue は買い増しを迫る表札の値 (3)。
	BuyInValue  int  `json:"buyInValue"`
	Pot         int  `json:"pot"`
	CurrentBet  int  `json:"currentBet"`
	ToCall      int  `json:"toCall"`
	RaiseCount  int  `json:"raiseCount"`
	CanRaise    bool `json:"canRaise"`
	TurnSeat    int  `json:"turnSeat"`
	HumanSeat   int  `json:"humanSeat"`
	IsHumanTurn bool `json:"isHumanTurn"`
	// BuyerSeat は買い増しを迫られている席 (-1 なら誰もいない)。
	BuyerSeat int `json:"buyerSeat"`
	// BuyCost は買い増しの額。
	BuyCost int `json:"buyCost"`
	// IsBuying は人間が買い増しを迫られているか。
	IsBuying       bool `json:"isBuying"`
	HandNumber     int  `json:"handNumber"`
	RemainingCards int  `json:"remainingCards"`
	WinnerSeat     int  `json:"winnerSeat"`
	GameEndFlag    bool `json:"gameEndFlag"`

	Config *BaseballPokerWebOutCfg `json:"config,omitempty"`
	WebOutputBase
}

// BaseballPokerWebController ベースボールポーカーWebコントローラークラス
type BaseballPokerWebController = GameWebController[usecase.BaseballPokerInteractorIF, BaseballPokerWebInput, *BaseballPokerWebOutput]

// NewBaseballPokerWebController and NewBaseballPokerWebControllerWithProvider
// are the standard and provider-backed constructors.
var NewBaseballPokerWebController, NewBaseballPokerWebControllerWithProvider = webControllerPair[usecase.BaseballPokerInteractorIF, BaseballPokerWebInput, *BaseballPokerWebOutput](
	newBaseballPokerDefaultOutput, baseballPokerDispatch,
)

func newBaseballPokerDefaultOutput(msg string) *BaseballPokerWebOutput {
	return &BaseballPokerWebOutput{
		Seats:         make([]*BaseballPokerWebOutputSeat, 0),
		WildValues:    []int{domain.BaseballWildThree, domain.BaseballWildNine},
		BuyerSeat:     -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

// baseballPokerAction はコマンドとドメインのアクション定数の対応。
var baseballPokerAction = map[string]int{
	"fold": domain.BaseballActionFold, "f": domain.BaseballActionFold,
	"check": domain.BaseballActionCheck, "k": domain.BaseballActionCheck,
	"call": domain.BaseballActionCall, "c": domain.BaseballActionCall,
	"bet": domain.BaseballActionBet, "b": domain.BaseballActionBet,
	"raise": domain.BaseballActionRaise, "r2": domain.BaseballActionRaise,
}

// baseballPokerBuyAnswer は買い増しの返事コマンドの対応。
//
// **名前で受けて数値では受けない。** 数値の本文にすると、送り忘れが
// 「0 番の返事」= 支払いに化けて、降りたつもりの席がポットを払う。
var baseballPokerBuyAnswer = map[string]int{
	"pay": domain.BaseballBuyPay, "p": domain.BaseballBuyPay,
	"buyfold": domain.BaseballBuyFold,
}

func baseballPokerDispatch(bc *baseController, w http.ResponseWriter, ci usecase.BaseballPokerInteractorIF, param BaseballPokerWebInput, newOut func(string) *BaseballPokerWebOutput) bool {
	switch param.Command {
	case "fold", "f", "check", "k", "call", "c":
		// **額を取らない手。** 送られていても無視する。
		bc.writePresenterResponse(w, ci.Action(baseballPokerAction[param.Command], 0))
	case "bet", "b", "raise", "r2":
		// **額は必須。** 0 を「省略」と同一視しないよう、未送信だけを弾く。
		if !requireParam(bc, w, newOut, param.Amount == nil, "param error: amount is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.Action(baseballPokerAction[param.Command], *param.Amount))
	case "pay", "p", "buyfold":
		bc.writePresenterResponse(w, ci.AnswerBuyIn(baseballPokerBuyAnswer[param.Command]))
	case "next":
		bc.writePresenterResponse(w, ci.NextHand())
	case "hint":
		bc.writePresenterResponse(w, ci.Hint())
	default:
		return dispatchResetAndLog(param.Command, bc, w, ci.Reset, ci.ActionLog)
	}
	return true
}
