//go:build !js || !wasm || extra

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// KingoWebInput キンゴWebインプット
type KingoWebInput struct {
	BaseWebInput
	// Amount は張り額。
	Amount *int `json:"amount,omitempty"`
}

// KingoWebOutCfg はキンゴの設定
type KingoWebOutCfg struct {
	Seats        int `json:"seats"`
	InitialChips int `json:"initialChips"`
	MinBet       int `json:"minBet"`
	Rounds       int `json:"rounds"`
}

// KingoWebOutputSeat は 1 席の状態
type KingoWebOutputSeat struct {
	Name    string `json:"name"`
	IsHuman bool   `json:"isHuman"`
	Chips   int    `json:"chips"`
	Bet     int    `json:"bet"`
	// Cards は手札 3 枚。**配る前と、決着前の他人の手は空。**
	Cards []*WebOutputCard `json:"cards"`
	// Rank は役 (0=なし, 1=2枚そろい, 2=嵐)。決着後のみ。
	Rank int `json:"rank"`
	// MatchedValue はそろえた数字。決着後のみ。
	MatchedValue int `json:"matchedValue"`
	// IsBanker はこのラウンドの親か。
	IsBanker  bool `json:"isBanker"`
	WonAmount int  `json:"wonAmount"`
}

// KingoWebOutput キンゴWebアウトプット
type KingoWebOutput struct {
	Phase int                   `json:"phase"`
	Seats []*KingoWebOutputSeat `json:"seats"`
	// BankerSeat は親の席。
	BankerSeat int `json:"bankerSeat"`
	// RoundNumber と Rounds で残りが読める。
	RoundNumber int `json:"roundNumber"`
	Rounds      int `json:"rounds"`
	HumanSeat   int `json:"humanSeat"`
	// IsHumanBanker は人間が親か。**張りと配るのどちらを出すかがこれで決まる。**
	IsHumanBanker bool `json:"isHumanBanker"`
	IsHumanTurn   bool `json:"isHumanTurn"`
	// HandSize は 1 席に配る枚数 (3)。
	HandSize int `json:"handSize"`
	// PayoutArashi と PayoutPair は役ごとの配当倍率。**画面に書き写させない。**
	PayoutArashi   int  `json:"payoutArashi"`
	PayoutPair     int  `json:"payoutPair"`
	RemainingCards int  `json:"remainingCards"`
	WinnerSeat     int  `json:"winnerSeat"`
	GameEndFlag    bool `json:"gameEndFlag"`

	Config *KingoWebOutCfg `json:"config,omitempty"`
	WebOutputBase
}

// KingoWebController キンゴWebコントローラークラス
type KingoWebController = GameWebController[usecase.KingoInteractorIF, KingoWebInput, *KingoWebOutput]

// NewKingoWebController and NewKingoWebControllerWithProvider
// are the standard and provider-backed constructors.
var NewKingoWebController, NewKingoWebControllerWithProvider = webControllerPair[usecase.KingoInteractorIF, KingoWebInput, *KingoWebOutput](
	newKingoDefaultOutput, kingoDispatch,
)

func newKingoDefaultOutput(msg string) *KingoWebOutput {
	return &KingoWebOutput{
		Seats:         make([]*KingoWebOutputSeat, 0),
		HandSize:      domain.KingoHandSize,
		PayoutArashi:  domain.KingoPayout(domain.KingoRankArashi),
		PayoutPair:    domain.KingoPayout(domain.KingoRankPair),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func kingoDispatch(bc *baseController, w http.ResponseWriter, ci usecase.KingoInteractorIF, param KingoWebInput, newOut func(string) *KingoWebOutput) bool {
	switch param.Command {
	case "bet", "b":
		// **額は必須。** 0 を「省略」と同一視しないよう、未送信だけを弾く。
		if !requireParam(bc, w, newOut, param.Amount == nil, "param error: amount is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.Bet(*param.Amount))
	case "deal", "d":
		bc.writePresenterResponse(w, ci.Deal())
	case "next":
		bc.writePresenterResponse(w, ci.NextRound())
	case "hint":
		bc.writePresenterResponse(w, ci.Hint())
	default:
		return dispatchResetAndLog(param.Command, bc, w, ci.Reset, ci.ActionLog)
	}
	return true
}
