//go:build !js || !wasm || extra

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// GleekWebInput グリーク (Gleek) のWebインプット
type GleekWebInput struct {
	BaseWebInput
	// CardIndex プレイするカードのインデックス
	CardIndex *int `json:"cardIndex,omitempty"`
	// Bid 競り額 (0=降りる)
	Bid *int `json:"bid,omitempty"`
	// DiscardIndices 捨て札フェーズで捨てる札のインデックス (ちょうど 7 枚)
	DiscardIndices []int `json:"discardIndices,omitempty"`
	// Config ゲーム設定
	Config *GleekWebConfig `json:"config,omitempty"`
}

// GleekWebConfig グリークのWeb設定
type GleekWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	TargetRounds  *int `json:"targetRounds,omitempty"`
}

// GleekWebOutputPlayer グリークのWebアウトプットプレイヤー
type GleekWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
	TrickCount int              `json:"trickCount"`
	Score      int              `json:"score"`
	// IsBuyer はこの席がストックを買ったか。
	IsBuyer bool `json:"isBuyer"`
	// Bid はこの席が置いた競り額、Passed は降りたか。
	Bid    int  `json:"bid"`
	Passed bool `json:"passed"`
	// TrickPoints はこのディールで取ったトリック点 (3 点 × トリック + 名札)。
	TrickPoints int `json:"trickPoints"`
	// Ruff はこの席のラフ (同一スートの最高合計) と、そのスート。
	Ruff     int `json:"ruff"`
	RuffSuit int `json:"ruffSuit"`
}

// GleekWebOutputMeld 申告されたグリーク / マーニヴァル 1 件。
type GleekWebOutputMeld struct {
	PlayerIdx int `json:"playerIdx"`
	Rank      int `json:"rank"`
	Count     int `json:"count"`
	Value     int `json:"value"`
}

// GleekWebOutput グリークのWebアウトプット
type GleekWebOutput struct {
	Players          []*GleekWebOutputPlayer `json:"players"`
	Phase            int                     `json:"phase"`
	RoundNumber      int                     `json:"roundNumber"`
	TrickNumber      int                     `json:"trickNumber"`
	CurrentPlayerIdx int                     `json:"currentPlayerIdx"`
	CurrentBidderIdx int                     `json:"currentBidderIdx"`
	LeadPlayerIdx    int                     `json:"leadPlayerIdx"`
	DealerIdx        int                     `json:"dealerIdx"`
	ElderIdx         int                     `json:"elderIdx"`
	BuyerIdx         int                     `json:"buyerIdx"`
	WinningBid       int                     `json:"winningBid"`
	HighestBid       int                     `json:"highestBid"`
	// NextBidAmount は次に置ける額 (0=これ以上競り上げられない)。**サーバが弾く
	// 選択肢を画面に出さないための値。**
	NextBidAmount int                        `json:"nextBidAmount"`
	TrumpSuit     int                        `json:"trumpSuit"`
	TurnUp        *WebOutputCard             `json:"turnUp,omitempty"`
	CurrentTrick  []*WebOutputTrickCard      `json:"currentTrick"`
	PlayerScores  [domain.GleekPlayerCnt]int `json:"playerScores"`
	// DiscardCount は捨て札フェーズで捨てる枚数。
	DiscardCount int `json:"discardCount"`
	// RuffWinnerIdx はラフを取った席 (-1=未確定)、Melds は申告されたメルド。
	RuffWinnerIdx int                   `json:"ruffWinnerIdx"`
	Melds         []*GleekWebOutputMeld `json:"melds"`
	// DealPoints はこのディールに実際にあった点、Par はその 1/3 (精算の基準)。
	DealPoints         int                `json:"dealPoints"`
	Par                int                `json:"par"`
	LastTrickWinner    int                `json:"lastTrickWinner"`
	Result             int                `json:"result"`
	PlayableIndices    []int              `json:"playableIndices"`
	GameEndFlag        bool               `json:"gameEndFlag"`
	WinnerPlayer       int                `json:"winnerPlayer"`
	IsHumanTurn        bool               `json:"isHumanTurn"`
	IsHumanBidTurn     bool               `json:"isHumanBidTurn"`
	IsHumanDiscardTurn bool               `json:"isHumanDiscardTurn"`
	Hint               *WebOutputCardHint `json:"hint,omitempty"`
	WebOutputBase
	Config GleekWebOutputConfig `json:"config"`
}

// GleekWebOutputConfig グリークの設定アウトプット
type GleekWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	TargetRounds  int `json:"targetRounds"`
}

// ToConfig builds a GleekConfig from the nested web config, applying bounds checking.
func (c *GleekWebConfig) ToConfig() domain.GleekConfig {
	cfg := domain.DefaultGleekConfig()
	cfg.CpuDifficulty = domain.GleekCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.GleekCpuDifficultyEasy), int(domain.GleekCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.TargetRounds, c.TargetRounds, 1, 1000)
	return cfg
}

// ToConfig builds a GleekConfig from the web input.
func (p GleekWebInput) ToConfig() domain.GleekConfig {
	return configOrDefault(p.Config, (*GleekWebConfig).ToConfig, domain.DefaultGleekConfig())
}

// GleekWebController グリークのWebコントローラークラス
type GleekWebController = GameWebController[usecase.GleekInteractorIF, GleekWebInput, *GleekWebOutput]

// NewGleekWebController and NewGleekWebControllerWithProvider are
// the standard and provider-backed constructors for GleekWebController.
var NewGleekWebController, NewGleekWebControllerWithProvider = webControllerPair[usecase.GleekInteractorIF, GleekWebInput, *GleekWebOutput](
	newGleekDefaultOutput, gleekDispatch,
)

func newGleekDefaultOutput(msg string) *GleekWebOutput {
	return &GleekWebOutput{
		Players:         make([]*GleekWebOutputPlayer, 0),
		CurrentTrick:    make([]*WebOutputTrickCard, 0),
		PlayableIndices: make([]int, 0),
		Melds:           make([]*GleekWebOutputMeld, 0),
		BuyerIdx:        -1,
		RuffWinnerIdx:   -1,
		TrumpSuit:       -1,
		LastTrickWinner: -1,
		WinnerPlayer:    -1,
		DiscardCount:    domain.GleekSwapSize,
		WebOutputBase:   WebOutputBase{Message: msg},
	}
}

func gleekDispatch(bc *baseController, w http.ResponseWriter, di usecase.GleekInteractorIF, param GleekWebInput, newDefault func(string) *GleekWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, di.ResetWithConfig(param.ToConfig()))
	case "b", "bid":
		if !requireParam(bc, w, newDefault, param.Bid == nil, "param error: bid is required.") {
			return true
		}
		bc.writePresenterResponse(w, di.Bid(*param.Bid))
	case "d", "discard":
		if !requireParam(bc, w, newDefault, param.DiscardIndices == nil, "param error: discardIndices is required.") {
			return true
		}
		bc.writePresenterResponse(w, di.Discard(param.DiscardIndices))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, di.Play(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, di.NextTrick())
	case "nr", "nextround":
		bc.writePresenterResponse(w, di.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, di.Hint, di.ActionLog)
	}
	return true
}
