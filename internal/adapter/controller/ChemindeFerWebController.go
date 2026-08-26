//go:build !js || !wasm || extra4

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// ChemindeFerWebInput シュマン・ド・フェールWebインプット
type ChemindeFerWebInput struct {
	BaseWebInput
	// Stake は親が張るバンク額。
	Stake *int `json:"stake,omitempty"`
	// Amount は子が賭ける額。**0 は「降りる」という有効な値**なので、
	// 省略と区別するためにポインタで受ける。
	Amount *int `json:"amount,omitempty"`
	Rounds *int `json:"rounds,omitempty"`
	Chips  *int `json:"chips,omitempty"`
}

// ChemindeFerWebOutputPlayer はシュマン・ド・フェールWebアウトプットの席情報
type ChemindeFerWebOutputPlayer struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	IsHuman bool   `json:"isHuman"`
	Chips   int    `json:"chips"`
	Bet     int    `json:"bet"`
	// LastNet は直前の決済でこの席が増減させたチップ。ラウンド中は 0。
	//
	// **卓の結果 (banker/punter/tie) だけでは、自分の賭けが勝ったのか負けたのかが
	// 分からない。** チップの数字を前後で見比べさせないための値です。
	LastNet int `json:"lastNet"`
	// IsBanker は親かどうか。**席では決まらず、ラウンドごとに回る。**
	IsBanker bool `json:"isBanker"`
	// IsRepresentative は子側の代表 (最高額ベッター) かどうか。
	IsRepresentative bool `json:"isRepresentative"`
}

// ChemindeFerWebOutput シュマン・ド・フェールWebアウトプット
type ChemindeFerWebOutput struct {
	Players   []*ChemindeFerWebOutputPlayer `json:"players"`
	Phase     int                           `json:"phase"`
	BankerIdx int                           `json:"bankerIdx"`
	// BetTurn は次に賭ける子の席 (-1: 賭けは終わっている)。
	BetTurn        int `json:"betTurn"`
	Stake          int `json:"stake"`
	RemainingStake int `json:"remainingStake"`
	TotalBet       int `json:"totalBet"`
	// StakeMin / StakeMax は親がいま張れる額の範囲。
	StakeMin int `json:"stakeMin"`
	StakeMax int `json:"stakeMax"`
	// BetMin / BetMax は手番の子がいま賭けられる額の範囲。
	BetMin            int `json:"betMin"`
	BetMax            int `json:"betMax"`
	RepresentativeIdx int `json:"representativeIdx"`
	// PunterMayChoose は子側がいま自分で引き方を選べるか。**合計 5 のときだけ真。**
	PunterMayChoose bool                  `json:"punterMayChoose"`
	BankerHand      []*WebOutputCard      `json:"bankerHand"`
	PunterHand      []*WebOutputCard      `json:"punterHand"`
	BankerTotal     int                   `json:"bankerTotal"`
	PunterTotal     int                   `json:"punterTotal"`
	PunterDrew      bool                  `json:"punterDrew"`
	Result          int                   `json:"result"`
	RoundNumber     int                   `json:"roundNumber"`
	RemainingCards  int                   `json:"remainingCards"`
	IsHumanTurn     bool                  `json:"isHumanTurn"`
	GameEndFlag     bool                  `json:"gameEndFlag"`
	Config          *ChemindeFerWebOutCfg `json:"config,omitempty"`
	WebOutputBase
}

// ChemindeFerWebOutCfg はシュマン・ド・フェールの設定
type ChemindeFerWebOutCfg struct {
	Rounds       int `json:"rounds"`
	InitialChips int `json:"initialChips"`
}

// ChemindeFerWebController シュマン・ド・フェールWebコントローラークラス
type ChemindeFerWebController = GameWebController[usecase.ChemindeFerInteractorIF, ChemindeFerWebInput, *ChemindeFerWebOutput]

// NewChemindeFerWebController and NewChemindeFerWebControllerWithProvider are
// the standard and provider-backed constructors for ChemindeFerWebController.
var NewChemindeFerWebController, NewChemindeFerWebControllerWithProvider = webControllerPair[usecase.ChemindeFerInteractorIF, ChemindeFerWebInput, *ChemindeFerWebOutput](
	newChemindeFerDefaultOutput, chemindeFerDispatch,
)

func newChemindeFerDefaultOutput(msg string) *ChemindeFerWebOutput {
	return &ChemindeFerWebOutput{
		Players:           make([]*ChemindeFerWebOutputPlayer, 0),
		BankerHand:        make([]*WebOutputCard, 0),
		PunterHand:        make([]*WebOutputCard, 0),
		BetTurn:           -1,
		RepresentativeIdx: -1,
		WebOutputBase:     WebOutputBase{Message: msg},
	}
}

// chemindeFerHumanSeat は人間の席。**卓の 0 番が人間**という構成に合わせる。
const chemindeFerHumanSeat = 0

// chemindeFerDispatch はコマンドをインタラクタへ振り分ける。
//
// **引くか立つかは親と子で別コマンドにしてある。** コントローラはフェーズを知らない
// ので、1 つの "draw" をフェーズで振り分けることができない。ここで推測すると、
// 子の手番に親の手が通るような取り違えが静かに起きる。
func chemindeFerDispatch(bc *baseController, w http.ResponseWriter, ci usecase.ChemindeFerInteractorIF, param ChemindeFerWebInput, newOut func(string) *ChemindeFerWebOutput) bool {
	switch param.Command {
	case "s", "stake":
		if !requireParam(bc, w, newOut, param.Stake == nil, "param error: stake is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.SetStake(*param.Stake))
	case "b", "bet":
		// **0 は「降りる」という有効な賭け額。** 省略と区別が要る。
		if !requireParam(bc, w, newOut, param.Amount == nil, "param error: amount is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.PlaceBet(chemindeFerHumanSeat, *param.Amount))
	case "pd", "punterdraw":
		bc.writePresenterResponse(w, ci.PunterDraw())
	case "ps", "punterstand":
		bc.writePresenterResponse(w, ci.PunterStand())
	case "bd", "bankerdraw":
		bc.writePresenterResponse(w, ci.BankerDraw())
	case "bs", "bankerstand":
		bc.writePresenterResponse(w, ci.BankerStand())
	case "d", "draw":
		// 側を書かない経路。**振り分けるのはドメインのフェーズ**なので、
		// クライアントが側を推測することにはならない。CLI から使う。
		bc.writePresenterResponse(w, ci.DrawOrStand(true))
	case "st", "stand":
		bc.writePresenterResponse(w, ci.DrawOrStand(false))
	case "pb", "passbank":
		bc.writePresenterResponse(w, ci.PassBank())
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
