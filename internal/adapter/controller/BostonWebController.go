//go:build !js || !wasm || extra3

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BostonWebInput ボストン Webインプット
type BostonWebInput struct {
	BaseWebInput
	CardIndex *int             `json:"cardIndex,omitempty"`
	Level     *int             `json:"level,omitempty"`
	Suit      *int             `json:"suit,omitempty"`
	Partner   *int             `json:"partner,omitempty"`
	Config    *BostonWebConfig `json:"config,omitempty"`
}

// BostonWebConfig ボストン Web設定
type BostonWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	TargetHands   *int `json:"targetHands,omitempty"`
}

// BostonWebOutputBid ボストン Webアウトプット宣言
type BostonWebOutputBid struct {
	Player int `json:"player"`
	// Level は序列。0 がパス。
	Level int    `json:"level"`
	Name  string `json:"name"`
	Suit  int    `json:"suit"`
}

// BostonWebOutputBidOption は選べる宣言の 1 件。
//
// **序列表そのものを送る。**ミゼールがトリック宣言の間に挟まるので、フロントで
// 並べ直すと必ずずれる。
type BostonWebOutputBidOption struct {
	Level int    `json:"level"`
	Name  string `json:"name"`
	// Kind は 1=トリック数, 2=ミゼール, 3=ピッコリッシモ。
	Kind int `json:"kind"`
	// Tricks は目標トリック数 (ミゼールは 0、ピッコリッシモは 1)。
	Tricks int `json:"tricks"`
	// NeedsTrump は切札の指定が要るか。
	NeedsTrump bool `json:"needsTrump"`
	// Exposed は第1トリック後に手札を公開する宣言か。
	Exposed bool `json:"exposed"`
	// CanCallPartner はパートナーを指名できるか。
	CanCallPartner bool `json:"canCallPartner"`
	Payout         int  `json:"payout"`
}

// BostonWebOutputPlayer ボストン Webアウトプットプレイヤー
type BostonWebOutputPlayer struct {
	ID        int  `json:"id"`
	IsHuman   bool `json:"isHuman"`
	CardCount int  `json:"cardCount"`
	// Cards は自分の手札、公開宣言の落札者、および精算後のみ。
	Cards          []*WebOutputCard `json:"cards"`
	TricksWon      int              `json:"tricksWon"`
	Chips          int              `json:"chips"`
	IsDealer       bool             `json:"isDealer"`
	IsDeclarer     bool             `json:"isDeclarer"`
	IsPartner      bool             `json:"isPartner"`
	IsDeclarerSide bool             `json:"isDeclarerSide"`
	IsCurrentTurn  bool             `json:"isCurrentTurn"`
}

// BostonWebOutput ボストン Webアウトプット
type BostonWebOutput struct {
	Players []*BostonWebOutputPlayer `json:"players"`
	Phase   int                      `json:"phase"`
	// HandNumber は何局目か。
	HandNumber       int                   `json:"handNumber"`
	CurrentPlayerIdx int                   `json:"currentPlayerIdx"`
	BidPlayerIdx     int                   `json:"bidPlayerIdx"`
	DealerIdx        int                   `json:"dealerIdx"`
	Bids             []*BostonWebOutputBid `json:"bids"`
	HighBid          *BostonWebOutputBid   `json:"highBid"`
	// BidOptions は序列表。サーバーが並べて送る。
	BidOptions  []*BostonWebOutputBidOption `json:"bidOptions"`
	DeclarerIdx int                         `json:"declarerIdx"`
	PartnerIdx  int                         `json:"partnerIdx"`
	TrumpSuit   int                         `json:"trumpSuit"`
	Exposed     bool                        `json:"exposed"`
	Trick       []*WebOutputCard            `json:"trick"`
	// ValidPlays は人間が出せる手札インデックス (追随が強制なため)。
	ValidPlays     []int `json:"validPlays"`
	TrickLeaderIdx int   `json:"trickLeaderIdx"`
	TrickNumber    int   `json:"trickNumber"`
	DeclarerTricks int   `json:"declarerTricks"`
	BidMade        bool  `json:"bidMade"`
	HandSize       int   `json:"handSize"`
	TargetHands    int   `json:"targetHands"`
	GameEndFlag    bool  `json:"gameEndFlag"`
	WinnerIdx      int   `json:"winnerIdx"`
	WebOutputBase
	Config BostonWebOutputConfig `json:"config"`
}

// BostonWebOutputConfig ボストン設定アウトプット
type BostonWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	TargetHands   int `json:"targetHands"`
}

// ToConfig builds a BostonConfig from the nested web config, applying bounds checking.
func (c *BostonWebConfig) ToConfig() domain.BostonConfig {
	cfg := domain.DefaultBostonConfig()
	cfg.CpuDifficulty = domain.BostonCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty,
		int(domain.BostonCpuDifficultyNormal), int(domain.BostonCpuDifficultyNormal), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.TargetHands, c.TargetHands,
		domain.BostonTargetHandsMin, domain.BostonTargetHandsMax)
	return cfg
}

// ToConfig builds a BostonConfig from the web input.
func (p BostonWebInput) ToConfig() domain.BostonConfig {
	return configOrDefault(p.Config, (*BostonWebConfig).ToConfig, domain.DefaultBostonConfig())
}

// BostonWebController ボストン Webコントローラークラス
type BostonWebController = GameWebController[usecase.BostonInteractorIF, BostonWebInput, *BostonWebOutput]

// NewBostonWebController and NewBostonWebControllerWithProvider are
// the standard and provider-backed constructors for BostonWebController.
var NewBostonWebController, NewBostonWebControllerWithProvider = webControllerPair[usecase.BostonInteractorIF, BostonWebInput, *BostonWebOutput](
	newBostonDefaultOutput, bostonDispatch,
)

func newBostonDefaultOutput(msg string) *BostonWebOutput {
	return &BostonWebOutput{
		Players:       make([]*BostonWebOutputPlayer, 0),
		Bids:          make([]*BostonWebOutputBid, 0),
		BidOptions:    make([]*BostonWebOutputBidOption, 0),
		Trick:         make([]*WebOutputCard, 0),
		ValidPlays:    make([]int, 0),
		DeclarerIdx:   -1,
		PartnerIdx:    -1,
		WinnerIdx:     -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func bostonDispatch(bc *baseController, w http.ResponseWriter, bi usecase.BostonInteractorIF, param BostonWebInput, newOut func(string) *BostonWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, bi.ResetWithConfig(param.ToConfig()))
	case "b", "bid":
		if !requireParam(bc, w, newOut, param.Level == nil, "param error: level is required.") {
			return true
		}
		suit := 0
		if param.Suit != nil {
			suit = *param.Suit
		}
		bc.writePresenterResponse(w, bi.Bid(domain.BostonBidLevel(*param.Level), suit))
	case "ps", "pass":
		bc.writePresenterResponse(w, bi.PassBid())
	case "cp", "callpartner":
		if !requireParam(bc, w, newOut, param.Partner == nil, "param error: partner is required.") {
			return true
		}
		bc.writePresenterResponse(w, bi.CallPartner(*param.Partner))
	case "p", "play":
		if !requireParam(bc, w, newOut, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, bi.PlayCard(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, bi.NextHand())
	default:
		return dispatchLog(param.Command, bc, w, bi.ActionLog)
	}
	return true
}
