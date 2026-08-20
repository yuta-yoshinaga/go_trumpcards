//go:build !js || !wasm || extra3

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// KaiserWebInput カイザー Webインプット
type KaiserWebInput struct {
	BaseWebInput
	CardIndex *int             `json:"cardIndex,omitempty"`
	Indices   []int            `json:"indices,omitempty"`
	Bid       *int             `json:"bid,omitempty"`
	Contract  *int             `json:"contract,omitempty"`
	Suit      *int             `json:"suit,omitempty"`
	Config    *KaiserWebConfig `json:"config,omitempty"`
}

// KaiserWebConfig カイザー Web設定
type KaiserWebConfig struct {
	CpuDifficulty *int  `json:"cpuDifficulty,omitempty"`
	AllowNoTrump  *bool `json:"allowNoTrump,omitempty"`
}

// KaiserWebOutputBid カイザー Webアウトプットビッド
type KaiserWebOutputBid struct {
	Player int `json:"player"`
	// Value は宣言した点数 (トリック数ではない)。パスなら 0。
	Value    int `json:"value"`
	Contract int `json:"contract"`
}

// KaiserWebOutputPlayer カイザー Webアウトプットプレイヤー
type KaiserWebOutputPlayer struct {
	ID      int  `json:"id"`
	IsHuman bool `json:"isHuman"`
	// Team は 0/2 が 0、1/3 が 1。
	Team      int `json:"team"`
	CardCount int `json:"cardCount"`
	// Cards は自分の手札のみ。相手は空で送る。
	Cards         []*WebOutputCard `json:"cards"`
	IsDealer      bool             `json:"isDealer"`
	IsDeclarer    bool             `json:"isDeclarer"`
	IsCurrentTurn bool             `json:"isCurrentTurn"`
}

// KaiserWebOutput カイザー Webアウトプット
type KaiserWebOutput struct {
	Players []*KaiserWebOutputPlayer `json:"players"`
	Phase   int                      `json:"phase"`
	// HandNumber は何局目か。
	HandNumber       int                   `json:"handNumber"`
	CurrentPlayerIdx int                   `json:"currentPlayerIdx"`
	BidPlayerIdx     int                   `json:"bidPlayerIdx"`
	DealerIdx        int                   `json:"dealerIdx"`
	Bids             []*KaiserWebOutputBid `json:"bids"`
	HighBid          *KaiserWebOutputBid   `json:"highBid"`
	DeclarerIdx      int                   `json:"declarerIdx"`
	TrumpSuit        int                   `json:"trumpSuit"`
	Contract         int                   `json:"contract"`
	KittySize        int                   `json:"kittySize"`
	Trick            []*WebOutputCard      `json:"trick"`
	TrickLeaderIdx   int                   `json:"trickLeaderIdx"`
	TrickNumber      int                   `json:"trickNumber"`
	// ValidPlays は人間が出せる手札インデックス (追随が強制なため)。
	ValidPlays []int `json:"validPlays"`
	// TeamHandPoints / TeamScores は [team0, team1]。
	TeamHandPoints [domain.KaiserTeamCnt]int `json:"teamHandPoints"`
	TeamScores     [domain.KaiserTeamCnt]int `json:"teamScores"`
	// HeartFiveBy / SpadeThreeBy は特殊札を取った席 (-1 なら未取得)。
	HeartFiveBy  int  `json:"heartFiveBy"`
	SpadeThreeBy int  `json:"spadeThreeBy"`
	BidMade      bool `json:"bidMade"`
	TargetScore  int  `json:"targetScore"`
	MinBid       int  `json:"minBid"`
	MaxBid       int  `json:"maxBid"`
	GameEndFlag  bool `json:"gameEndFlag"`
	WinnerTeam   int  `json:"winnerTeam"`
	WebOutputBase
	Config KaiserWebOutputConfig `json:"config"`
}

// KaiserWebOutputConfig カイザー設定アウトプット
type KaiserWebOutputConfig struct {
	CpuDifficulty int  `json:"cpuDifficulty"`
	AllowNoTrump  bool `json:"allowNoTrump"`
}

// ToConfig builds a KaiserConfig from the nested web config, applying bounds checking.
func (c *KaiserWebConfig) ToConfig() domain.KaiserConfig {
	cfg := domain.DefaultKaiserConfig()
	cfg.CpuDifficulty = domain.KaiserCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty,
		int(domain.KaiserCpuDifficultyNormal), int(domain.KaiserCpuDifficultyNormal), int(cfg.CpuDifficulty)))
	if c.AllowNoTrump != nil {
		cfg.AllowNoTrump = *c.AllowNoTrump
	}
	return cfg
}

// ToConfig builds a KaiserConfig from the web input.
func (p KaiserWebInput) ToConfig() domain.KaiserConfig {
	return configOrDefault(p.Config, (*KaiserWebConfig).ToConfig, domain.DefaultKaiserConfig())
}

// KaiserWebController カイザー Webコントローラークラス
type KaiserWebController = GameWebController[usecase.KaiserInteractorIF, KaiserWebInput, *KaiserWebOutput]

// NewKaiserWebController and NewKaiserWebControllerWithProvider are
// the standard and provider-backed constructors for KaiserWebController.
var NewKaiserWebController, NewKaiserWebControllerWithProvider = webControllerPair[usecase.KaiserInteractorIF, KaiserWebInput, *KaiserWebOutput](
	newKaiserDefaultOutput, kaiserDispatch,
)

func newKaiserDefaultOutput(msg string) *KaiserWebOutput {
	return &KaiserWebOutput{
		Players:       make([]*KaiserWebOutputPlayer, 0),
		Bids:          make([]*KaiserWebOutputBid, 0),
		Trick:         make([]*WebOutputCard, 0),
		ValidPlays:    make([]int, 0),
		DeclarerIdx:   -1,
		HeartFiveBy:   -1,
		SpadeThreeBy:  -1,
		WinnerTeam:    -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func kaiserDispatch(bc *baseController, w http.ResponseWriter, ki usecase.KaiserInteractorIF, param KaiserWebInput, newOut func(string) *KaiserWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, ki.ResetWithConfig(param.ToConfig()))
	case "b", "bid":
		if !requireParam(bc, w, newOut, param.Bid == nil, "param error: bid is required.") {
			return true
		}
		contract := 0
		if param.Contract != nil {
			contract = *param.Contract
		}
		bc.writePresenterResponse(w, ki.Bid(*param.Bid, domain.KaiserContract(contract)))
	case "ps", "pass":
		bc.writePresenterResponse(w, ki.PassBid())
	case "t", "trump":
		if !requireParam(bc, w, newOut, param.Suit == nil, "param error: suit is required.") {
			return true
		}
		bc.writePresenterResponse(w, ki.SetTrump(*param.Suit))
	case "d", "discard":
		if !requireParam(bc, w, newOut, param.Indices == nil, "param error: indices is required.") {
			return true
		}
		bc.writePresenterResponse(w, ki.Discard(param.Indices))
	case "p", "play":
		if !requireParam(bc, w, newOut, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, ki.PlayCard(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, ki.NextHand())
	default:
		return dispatchHintAndLog(param.Command, bc, w, ki.Hint, ki.ActionLog)
	}
	return true
}
