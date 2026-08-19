//go:build !js || !wasm || extra2

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BidEuchreWebInput ビッド・ユーカー Webインプット
type BidEuchreWebInput struct {
	BaseWebInput
	CardIndex *int                `json:"cardIndex,omitempty"`
	Value     *int                `json:"value,omitempty"`
	Trump     *int                `json:"trump,omitempty"`
	Config    *BidEuchreWebConfig `json:"config,omitempty"`
}

// BidEuchreWebConfig ビッド・ユーカー Web設定
type BidEuchreWebConfig struct {
	CpuDifficulty *int  `json:"cpuDifficulty,omitempty"`
	AllowNoTrump  *bool `json:"allowNoTrump,omitempty"`
}

// BidEuchreWebOutputBid ビッド・ユーカー Webアウトプット宣言
type BidEuchreWebOutputBid struct {
	Player int `json:"player"`
	// Value は宣言したトリック数 (0 ならパス)。
	Value int `json:"value"`
}

// BidEuchreWebOutputPlayer ビッド・ユーカー Webアウトプットプレイヤー
type BidEuchreWebOutputPlayer struct {
	ID      int  `json:"id"`
	IsHuman bool `json:"isHuman"`
	// Team は 0/2 が 0、1/3 が 1。
	Team      int `json:"team"`
	CardCount int `json:"cardCount"`
	// Cards は自分の手札のみ。**キティが無く、誰の手札も公開されない。**
	Cards         []*WebOutputCard `json:"cards"`
	TricksWon     int              `json:"tricksWon"`
	IsDealer      bool             `json:"isDealer"`
	IsDeclarer    bool             `json:"isDeclarer"`
	IsCurrentTurn bool             `json:"isCurrentTurn"`
}

// BidEuchreWebOutputResult ビッド・ユーカー Webアウトプット精算
type BidEuchreWebOutputResult struct {
	// Points は各チームの得点。**未達側は宣言額ぶん引かれる。**
	Points [domain.BidEuchreTeamCnt]int `json:"points"`
	// Tricks は各チームが取ったトリック数。
	Tricks [domain.BidEuchreTeamCnt]int `json:"tricks"`
	Made   bool                         `json:"made"`
	Bid    int                          `json:"bid"`
}

// BidEuchreWebOutput ビッド・ユーカー Webアウトプット
type BidEuchreWebOutput struct {
	Players []*BidEuchreWebOutputPlayer `json:"players"`
	Phase   int                         `json:"phase"`
	// HandNumber は何局目か。
	HandNumber       int                      `json:"handNumber"`
	CurrentPlayerIdx int                      `json:"currentPlayerIdx"`
	BidPlayerIdx     int                      `json:"bidPlayerIdx"`
	DealerIdx        int                      `json:"dealerIdx"`
	Bids             []*BidEuchreWebOutputBid `json:"bids"`
	HighBid          *BidEuchreWebOutputBid   `json:"highBid"`
	DeclarerIdx      int                      `json:"declarerIdx"`
	// Trump は宣言種別、TrumpSuit は切札スート (ノートランプなら 0)。
	Trump       int              `json:"trump"`
	TrumpSuit   int              `json:"trumpSuit"`
	TrumpChosen bool             `json:"trumpChosen"`
	Trick       []*WebOutputCard `json:"trick"`
	// ValidPlays は人間が出せる手札インデックス (追随が強制なため)。
	ValidPlays     []int `json:"validPlays"`
	TrickLeaderIdx int   `json:"trickLeaderIdx"`
	TrickNumber    int   `json:"trickNumber"`
	// TeamTricks / Scores は [team0, team1]。
	TeamTricks  [domain.BidEuchreTeamCnt]int `json:"teamTricks"`
	Scores      [domain.BidEuchreTeamCnt]int `json:"scores"`
	LastResult  *BidEuchreWebOutputResult    `json:"lastResult"`
	GameTarget  int                          `json:"gameTarget"`
	MinBid      int                          `json:"minBid"`
	MaxBid      int                          `json:"maxBid"`
	HandSize    int                          `json:"handSize"`
	GameEndFlag bool                         `json:"gameEndFlag"`
	WinnerTeam  int                          `json:"winnerTeam"`
	WebOutputBase
	Config BidEuchreWebOutputConfig `json:"config"`
}

// BidEuchreWebOutputConfig ビッド・ユーカー設定アウトプット
type BidEuchreWebOutputConfig struct {
	CpuDifficulty int  `json:"cpuDifficulty"`
	AllowNoTrump  bool `json:"allowNoTrump"`
}

// ToConfig builds a BidEuchreConfig from the nested web config, applying bounds checking.
func (c *BidEuchreWebConfig) ToConfig() domain.BidEuchreConfig {
	cfg := domain.DefaultBidEuchreConfig()
	cfg.CpuDifficulty = domain.BidEuchreCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty,
		int(domain.BidEuchreCpuDifficultyNormal), int(domain.BidEuchreCpuDifficultyNormal), int(cfg.CpuDifficulty)))
	if c.AllowNoTrump != nil {
		cfg.AllowNoTrump = *c.AllowNoTrump
	}
	return cfg
}

// ToConfig builds a BidEuchreConfig from the web input.
func (p BidEuchreWebInput) ToConfig() domain.BidEuchreConfig {
	return configOrDefault(p.Config, (*BidEuchreWebConfig).ToConfig, domain.DefaultBidEuchreConfig())
}

// BidEuchreWebController ビッド・ユーカー Webコントローラークラス
type BidEuchreWebController = GameWebController[usecase.BidEuchreInteractorIF, BidEuchreWebInput, *BidEuchreWebOutput]

// NewBidEuchreWebController and NewBidEuchreWebControllerWithProvider are
// the standard and provider-backed constructors for BidEuchreWebController.
var NewBidEuchreWebController, NewBidEuchreWebControllerWithProvider = webControllerPair[usecase.BidEuchreInteractorIF, BidEuchreWebInput, *BidEuchreWebOutput](
	newBidEuchreDefaultOutput, bidEuchreDispatch,
)

func newBidEuchreDefaultOutput(msg string) *BidEuchreWebOutput {
	return &BidEuchreWebOutput{
		Players:       make([]*BidEuchreWebOutputPlayer, 0),
		Bids:          make([]*BidEuchreWebOutputBid, 0),
		Trick:         make([]*WebOutputCard, 0),
		ValidPlays:    make([]int, 0),
		DeclarerIdx:   -1,
		WinnerTeam:    -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func bidEuchreDispatch(bc *baseController, w http.ResponseWriter, bi usecase.BidEuchreInteractorIF, param BidEuchreWebInput, newOut func(string) *BidEuchreWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, bi.ResetWithConfig(param.ToConfig()))
	case "b", "bid":
		if !requireParam(bc, w, newOut, param.Value == nil, "param error: value is required.") {
			return true
		}
		bc.writePresenterResponse(w, bi.Bid(*param.Value))
	case "ps", "pass":
		bc.writePresenterResponse(w, bi.PassBid())
	case "t", "trump":
		if !requireParam(bc, w, newOut, param.Trump == nil, "param error: trump is required.") {
			return true
		}
		bc.writePresenterResponse(w, bi.ChooseTrump(*param.Trump))
	case "p", "play":
		if !requireParam(bc, w, newOut, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, bi.PlayCard(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, bi.NextHand())
	default:
		return dispatchHintAndLog(param.Command, bc, w, bi.Hint, bi.ActionLog)
	}
	return true
}
