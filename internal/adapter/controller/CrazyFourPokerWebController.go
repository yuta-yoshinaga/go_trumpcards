//go:build !js || !wasm || casino

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// CrazyFourPokerWebInput クレイジー 4 ポーカーWebインプット
type CrazyFourPokerWebInput struct {
	BaseWebInput
	// Ante はアンティ額。同額の Super Bonus が自動で付く。
	Ante *int `json:"ante,omitempty"`
	// QueensUp は任意のサイドベット額。**0 は「置かない」という有効な値**なので
	// 省略と区別するためにポインタで受ける。
	QueensUp *int `json:"queensUp,omitempty"`
	// Multiplier はプレイベットの倍率。
	Multiplier *int `json:"multiplier,omitempty"`
}

// CrazyFourPokerWebOutCfg はクレイジー 4 ポーカーの設定
type CrazyFourPokerWebOutCfg struct {
	InitialChips int `json:"initialChips"`
	DefaultAnte  int `json:"defaultAnte"`
}

// CrazyFourPokerWebOutput クレイジー 4 ポーカーWebアウトプット
type CrazyFourPokerWebOutput struct {
	Phase int `json:"phase"`
	// PlayerHand は配られた 5 枚。**勝負に使うのはこのうち 4 枚。**
	PlayerHand []*WebOutputCard `json:"playerHand"`
	// DealerHand はディーラーの 5 枚。決着前は空で返す。
	DealerHand []*WebOutputCard `json:"dealerHand"`
	PlayerBest []*WebOutputCard `json:"playerBest"`
	DealerBest []*WebOutputCard `json:"dealerBest"`
	// PlayerHandRank / DealerHandRank は 4 枚役のランク (1..8)。
	PlayerHandRank int `json:"playerHandRank"`
	DealerHandRank int `json:"dealerHandRank"`
	// HasAcesOrBetter はエースのペア以上か。**3 倍を出せる条件。**
	HasAcesOrBetter bool `json:"hasAcesOrBetter"`
	// MaxMultiplier はいま置ける上限倍率。ページはこれに従うこと。
	MaxMultiplier int `json:"maxMultiplier"`
	// PlayerQualifies はプレイヤーの手がキング以上か。ヒントが使う。
	PlayerQualifies bool `json:"playerQualifies"`
	DealerQualifies bool `json:"dealerQualifies"`
	AnteBet         int  `json:"anteBet"`
	SuperBet        int  `json:"superBet"`
	QueensUpBet     int  `json:"queensUpBet"`
	PlayBet         int  `json:"playBet"`
	PlayMultiplier  int  `json:"playMultiplier"`
	Result          int  `json:"result"`
	Payout          int  `json:"payout"`
	Chips           int  `json:"chips"`
	MinTotalWager   int  `json:"minTotalWager"`
	RoundNumber     int  `json:"roundNumber"`
	RemainingCards  int  `json:"remainingCards"`
	GameEndFlag     bool `json:"gameEndFlag"`

	// QueensUpPayouts は Queens Up サイドベットの配当表 (配当の高い順)。
	//
	// **賭ける前に見えなければ意味がない。** いくら置くかは、何が当たれば
	// 何倍かを知って決めるものです。
	QueensUpPayouts []*CrazyFourPokerPayoutRow `json:"queensUpPayouts"`

	Config *CrazyFourPokerWebOutCfg `json:"config,omitempty"`
	WebOutputBase
}

// CrazyFourPokerPayoutRow は配当表の 1 行。
type CrazyFourPokerPayoutRow struct {
	// Hand は 4 枚役のランク (domain.FourCardHand*)。
	Hand int `json:"hand"`
	// Name は役の表示名。
	Name string `json:"name"`
	// Multiplier は X:1 の X。
	Multiplier int `json:"multiplier"`
}

// CrazyFourPokerWebController クレイジー 4 ポーカーWebコントローラークラス
type CrazyFourPokerWebController = GameWebController[usecase.CrazyFourPokerInteractorIF, CrazyFourPokerWebInput, *CrazyFourPokerWebOutput]

// NewCrazyFourPokerWebController and NewCrazyFourPokerWebControllerWithProvider are
// the standard and provider-backed constructors for CrazyFourPokerWebController.
var NewCrazyFourPokerWebController, NewCrazyFourPokerWebControllerWithProvider = webControllerPair[usecase.CrazyFourPokerInteractorIF, CrazyFourPokerWebInput, *CrazyFourPokerWebOutput](
	newCrazyFourPokerDefaultOutput, crazyFourPokerDispatch,
)

func newCrazyFourPokerDefaultOutput(msg string) *CrazyFourPokerWebOutput {
	return &CrazyFourPokerWebOutput{
		PlayerHand:    make([]*WebOutputCard, 0),
		DealerHand:    make([]*WebOutputCard, 0),
		PlayerBest:    make([]*WebOutputCard, 0),
		DealerBest:    make([]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func crazyFourPokerDispatch(bc *baseController, w http.ResponseWriter, ci usecase.CrazyFourPokerInteractorIF, param CrazyFourPokerWebInput, newOut func(string) *CrazyFourPokerWebOutput) bool {
	switch param.Command {
	case "b", "bet":
		if !requireParam(bc, w, newOut, param.Ante == nil, "param error: ante is required.") {
			return true
		}
		// **Queens Up は省略できる。** 省略は 0 (置かない) と同じ扱いでよい。
		queensUp := 0
		if param.QueensUp != nil {
			queensUp = *param.QueensUp
		}
		bc.writePresenterResponse(w, ci.PlaceBet(*param.Ante, queensUp))
	case "p", "play":
		if !requireParam(bc, w, newOut, param.Multiplier == nil, "param error: multiplier is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.Play(*param.Multiplier))
	case "f", "fold":
		bc.writePresenterResponse(w, ci.Fold())
	case "next":
		bc.writePresenterResponse(w, ci.NextRound())
	case "h", "hint":
		bc.writePresenterResponse(w, ci.Hint())
	default:
		return dispatchResetAndLog(param.Command, bc, w, ci.Reset, ci.ActionLog)
	}
	return true
}
