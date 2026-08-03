//go:build !js || !wasm || extra2

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// SixBidSoloWebInput シックスビッド・ソロ Webインプット
type SixBidSoloWebInput struct {
	BaseWebInput
	CardIndex *int `json:"cardIndex,omitempty"`
	Bid       *int `json:"bid,omitempty"`
	Suit      *int `json:"suit,omitempty"`
	// CalledSuit / CalledValue はコール・ソロの指名札。
	CalledSuit  *int                 `json:"calledSuit,omitempty"`
	CalledValue *int                 `json:"calledValue,omitempty"`
	Config      *SixBidSoloWebConfig `json:"config,omitempty"`
}

// SixBidSoloWebConfig シックスビッド・ソロ Web設定
type SixBidSoloWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	TargetHands   *int `json:"targetHands,omitempty"`
}

// SixBidSoloWebOutputBid シックスビッド・ソロ Webアウトプット宣言
type SixBidSoloWebOutputBid struct {
	Player int `json:"player"`
	// Kind は 0=パス 1=ソロ 2=ハートソロ 3=ミゼール 4=ギャランティー 5=スプレッド 6=コール。
	Kind int `json:"kind"`
}

// SixBidSoloWebOutputPlayer シックスビッド・ソロ Webアウトプットプレイヤー
type SixBidSoloWebOutputPlayer struct {
	ID        int  `json:"id"`
	IsHuman   bool `json:"isHuman"`
	CardCount int  `json:"cardCount"`
	// Cards は自分の手札のみ。**スプレッド・ミゼールでは宣言者の手札も公開される。**
	Cards         []*WebOutputCard `json:"cards"`
	Points        int              `json:"points"`
	TricksWon     int              `json:"tricksWon"`
	Score         int              `json:"score"`
	IsDealer      bool             `json:"isDealer"`
	IsDeclarer    bool             `json:"isDeclarer"`
	IsCurrentTurn bool             `json:"isCurrentTurn"`
}

// SixBidSoloWebOutputResult シックスビッド・ソロ Webアウトプット精算
type SixBidSoloWebOutputResult struct {
	Kind     int `json:"kind"`
	Declarer int `json:"declarer"`
	// DeclarerPoints は宣言者のカード点 (ウィドウ込み)。
	DeclarerPoints int `json:"declarerPoints"`
	// WidowPoints はウィドウのカード点。**ミゼール系では 0。**
	WidowPoints int  `json:"widowPoints"`
	Target      int  `json:"target"`
	Made        bool `json:"made"`
	// Value は 1 人あたりの受け払い額。
	Value  int                             `json:"value"`
	Deltas [domain.SixBidSoloPlayerCnt]int `json:"deltas"`
}

// SixBidSoloWebOutput シックスビッド・ソロ Webアウトプット
type SixBidSoloWebOutput struct {
	Players []*SixBidSoloWebOutputPlayer `json:"players"`
	Phase   int                          `json:"phase"`
	// HandNumber は何局目か。
	HandNumber       int                       `json:"handNumber"`
	CurrentPlayerIdx int                       `json:"currentPlayerIdx"`
	BidPlayerIdx     int                       `json:"bidPlayerIdx"`
	DealerIdx        int                       `json:"dealerIdx"`
	Bids             []*SixBidSoloWebOutputBid `json:"bids"`
	HighBid          *SixBidSoloWebOutputBid   `json:"highBid"`
	DeclarerIdx      int                       `json:"declarerIdx"`
	TrumpSuit        int                       `json:"trumpSuit"`
	Declared         bool                      `json:"declared"`
	// CalledCard はコール・ソロで指名された札。
	CalledCard *WebOutputCard `json:"calledCard"`
	// SpreadOpen はスプレッド・ミゼールで宣言者の手札が公開済みか。
	SpreadOpen bool `json:"spreadOpen"`
	// Widow は精算後だけ中身が入る。**ウィドウは宣言者の得点に加算される。**
	Widow []*WebOutputCard `json:"widow"`
	// WidowSize は伏せ札の枚数 (公開前でも枚数だけは見せる)。
	WidowSize int              `json:"widowSize"`
	Trick     []*WebOutputCard `json:"trick"`
	// ValidPlays は人間が出せる手札インデックス (追随が強制なため)。
	ValidPlays     []int                      `json:"validPlays"`
	TrickLeaderIdx int                        `json:"trickLeaderIdx"`
	TrickNumber    int                        `json:"trickNumber"`
	LastResult     *SixBidSoloWebOutputResult `json:"lastResult"`
	// BidTargets は各ビッドの目標カード点 (0=パスは 0)。
	BidTargets [domain.SixBidSoloBidCount]int `json:"bidTargets"`
	// TotalPoints は場に出るカード点の総和 (120)。
	TotalPoints int `json:"totalPoints"`
	// BaseTarget は通常ビッドの基準点 (60)。**超えることが要る。**
	BaseTarget  int  `json:"baseTarget"`
	HandSize    int  `json:"handSize"`
	TargetHands int  `json:"targetHands"`
	GameEndFlag bool `json:"gameEndFlag"`
	WinnerIdx   int  `json:"winnerIdx"`
	WebOutputBase
	Config SixBidSoloWebOutputConfig `json:"config"`
}

// SixBidSoloWebOutputConfig シックスビッド・ソロ設定アウトプット
type SixBidSoloWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	TargetHands   int `json:"targetHands"`
}

// ToConfig builds a SixBidSoloConfig from the nested web config, applying bounds checking.
func (c *SixBidSoloWebConfig) ToConfig() domain.SixBidSoloConfig {
	cfg := domain.DefaultSixBidSoloConfig()
	cfg.CpuDifficulty = domain.SixBidSoloCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty,
		int(domain.SixBidSoloCpuDifficultyNormal), int(domain.SixBidSoloCpuDifficultyNormal), int(cfg.CpuDifficulty)))
	cfg.TargetHands = webutil.BoundedIntPtr(c.TargetHands,
		domain.SixBidSoloMinHands, domain.SixBidSoloMaxHands, cfg.TargetHands)
	return cfg
}

// ToConfig builds a SixBidSoloConfig from the web input.
func (p SixBidSoloWebInput) ToConfig() domain.SixBidSoloConfig {
	return configOrDefault(p.Config, (*SixBidSoloWebConfig).ToConfig, domain.DefaultSixBidSoloConfig())
}

// SixBidSoloWebController シックスビッド・ソロ Webコントローラークラス
type SixBidSoloWebController = GameWebController[usecase.SixBidSoloInteractorIF, SixBidSoloWebInput, *SixBidSoloWebOutput]

// NewSixBidSoloWebController and NewSixBidSoloWebControllerWithProvider are
// the standard and provider-backed constructors for SixBidSoloWebController.
var NewSixBidSoloWebController, NewSixBidSoloWebControllerWithProvider = webControllerPair[usecase.SixBidSoloInteractorIF, SixBidSoloWebInput, *SixBidSoloWebOutput](
	newSixBidSoloDefaultOutput, sixBidSoloDispatch,
)

func newSixBidSoloDefaultOutput(msg string) *SixBidSoloWebOutput {
	return &SixBidSoloWebOutput{
		Players:       make([]*SixBidSoloWebOutputPlayer, 0),
		Bids:          make([]*SixBidSoloWebOutputBid, 0),
		Widow:         make([]*WebOutputCard, 0),
		Trick:         make([]*WebOutputCard, 0),
		ValidPlays:    make([]int, 0),
		DeclarerIdx:   -1,
		WinnerIdx:     -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func sixBidSoloDispatch(bc *baseController, w http.ResponseWriter, si usecase.SixBidSoloInteractorIF, param SixBidSoloWebInput, newOut func(string) *SixBidSoloWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, si.ResetWithConfig(param.ToConfig()))
	case "b", "bid":
		if !requireParam(bc, w, newOut, param.Bid == nil, "param error: bid is required.") {
			return true
		}
		bc.writePresenterResponse(w, si.Bid(*param.Bid))
	case "ps", "pass":
		bc.writePresenterResponse(w, si.PassBid())
	case "d", "declare":
		if !requireParam(bc, w, newOut, param.Suit == nil, "param error: suit is required.") {
			return true
		}
		// **指名札はコール・ソロだけに要る。**片方だけ来たら不足として扱う。
		if !requireParam(bc, w, newOut, (param.CalledSuit == nil) != (param.CalledValue == nil),
			"param error: calledSuit and calledValue must be given together.") {
			return true
		}
		bc.writePresenterResponse(w, si.Declare(*param.Suit,
			webutil.BoundedIntPtr(param.CalledSuit, 0, domain.CardDesignDiamond, 0),
			webutil.BoundedIntPtr(param.CalledValue, 1, 13, 0)))
	case "p", "play":
		if !requireParam(bc, w, newOut, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, si.PlayCard(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, si.NextHand())
	default:
		return dispatchLog(param.Command, bc, w, si.ActionLog)
	}
	return true
}
