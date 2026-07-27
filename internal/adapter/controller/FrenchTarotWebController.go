//go:build !js || !wasm || extra

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// FrenchTarotWebInput フレンチタロット (French Tarot) のWebインプット
type FrenchTarotWebInput struct {
	BaseWebInput
	// CardIndex プレイするカードのインデックス
	CardIndex *int `json:"cardIndex,omitempty"`
	// CardIndices エカルトで捨てるカードのインデックス (6 枚)
	CardIndices []int `json:"cardIndices,omitempty"`
	// Bid 入札 ("petite"/"garde"/"gardesans"/"gardecontre")
	Bid *string `json:"bid,omitempty"`
	// Config ゲーム設定
	Config *FrenchTarotWebConfig `json:"config,omitempty"`
}

// FrenchTarotWebConfig フレンチタロットのWeb設定
type FrenchTarotWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	TargetDeals   *int `json:"targetDeals,omitempty"`
}

// FrenchTarotWebOutputPlayer フレンチタロットのWebアウトプットプレイヤー
type FrenchTarotWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
	TrickCount int              `json:"trickCount"`
	CardPoints int              `json:"cardPoints"`
	Score      int              `json:"score"`
	IsDeclarer bool             `json:"isDeclarer"`
}

// FrenchTarotWebOutputHint ヒント出力
type FrenchTarotWebOutputHint struct {
	Bid         *int   `json:"bid,omitempty"`
	CardIndices []int  `json:"cardIndices"`
	Reason      string `json:"reason"`
}

// FrenchTarotWebOutput フレンチタロットのWebアウトプット
type FrenchTarotWebOutput struct {
	Players          []*FrenchTarotWebOutputPlayer    `json:"players"`
	Phase            int                              `json:"phase"`
	RoundNumber      int                              `json:"roundNumber"`
	TrickNumber      int                              `json:"trickNumber"`
	CurrentPlayerIdx int                              `json:"currentPlayerIdx"`
	LeadPlayerIdx    int                              `json:"leadPlayerIdx"`
	DealerIdx        int                              `json:"dealerIdx"`
	BidPlayerIdx     int                              `json:"bidPlayerIdx"`
	HighestBid       int                              `json:"highestBid"`
	HighestBidder    int                              `json:"highestBidder"`
	DeclarerIdx      int                              `json:"declarerIdx"`
	Contract         int                              `json:"contract"`
	ChienCount       int                              `json:"chienCount"`
	Chien            []*WebOutputCard                 `json:"chien"`
	ChienRevealed    bool                             `json:"chienRevealed"`
	StashOwner       int                              `json:"stashOwner"`
	CurrentTrick     []*WebOutputTrickCard            `json:"currentTrick"`
	PlayerScores     [domain.FrenchTarotPlayerCnt]int `json:"playerScores"`
	LastTrickWinner  int                              `json:"lastTrickWinner"`
	Outcome          int                              `json:"outcome"`
	Result           int                              `json:"result"`
	PlayableIndices  []int                            `json:"playableIndices"`
	GameEndFlag      bool                             `json:"gameEndFlag"`
	WinnerPlayer     int                              `json:"winnerPlayer"`
	IsHumanTurn      bool                             `json:"isHumanTurn"`
	IsHumanBidTurn   bool                             `json:"isHumanBidTurn"`
	IsHumanDiscard   bool                             `json:"isHumanDiscard"`
	Hint             *FrenchTarotWebOutputHint        `json:"hint,omitempty"`
	WebOutputBase
	Config FrenchTarotWebOutputConfig `json:"config"`
}

// FrenchTarotWebOutputConfig フレンチタロットの設定アウトプット
type FrenchTarotWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	TargetDeals   int `json:"targetDeals"`
}

// ToConfig builds a FrenchTarotConfig from the nested web config, applying bounds checking.
func (c *FrenchTarotWebConfig) ToConfig() domain.FrenchTarotConfig {
	cfg := domain.DefaultFrenchTarotConfig()
	cfg.CpuDifficulty = domain.FrenchTarotCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.FrenchTarotCpuDifficultyEasy), int(domain.FrenchTarotCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.TargetDeals, c.TargetDeals, 1, 1000)
	return cfg
}

// ToConfig builds a FrenchTarotConfig from the web input.
func (p FrenchTarotWebInput) ToConfig() domain.FrenchTarotConfig {
	return configOrDefault(p.Config, (*FrenchTarotWebConfig).ToConfig, domain.DefaultFrenchTarotConfig())
}

// frenchTarotParseBid 入札文字列を FrenchTarotBid に変換する (Pass=不正/パス)。
func frenchTarotParseBid(s string) domain.FrenchTarotBid {
	switch s {
	case "petite", "p":
		return domain.FrenchTarotBidPetite
	case "garde", "g":
		return domain.FrenchTarotBidGarde
	case "gardesans", "gs", "garde-sans":
		return domain.FrenchTarotBidGardeSans
	case "gardecontre", "gc", "garde-contre":
		return domain.FrenchTarotBidGardeContre
	default:
		return domain.FrenchTarotBidPass
	}
}

// FrenchTarotWebController フレンチタロットのWebコントローラークラス
type FrenchTarotWebController = GameWebController[usecase.FrenchTarotInteractorIF, FrenchTarotWebInput, *FrenchTarotWebOutput]

// NewFrenchTarotWebController and NewFrenchTarotWebControllerWithProvider are
// the standard and provider-backed constructors for FrenchTarotWebController.
var NewFrenchTarotWebController, NewFrenchTarotWebControllerWithProvider = webControllerPair[usecase.FrenchTarotInteractorIF, FrenchTarotWebInput, *FrenchTarotWebOutput](
	newFrenchTarotDefaultOutput, frenchTarotDispatch,
)

func newFrenchTarotDefaultOutput(msg string) *FrenchTarotWebOutput {
	return &FrenchTarotWebOutput{
		Players:         make([]*FrenchTarotWebOutputPlayer, 0),
		CurrentTrick:    make([]*WebOutputTrickCard, 0),
		Chien:           make([]*WebOutputCard, 0),
		PlayableIndices: make([]int, 0),
		DeclarerIdx:     -1,
		HighestBidder:   -1,
		LastTrickWinner: -1,
		WinnerPlayer:    -1,
		WebOutputBase:   WebOutputBase{Message: msg},
	}
}

func frenchTarotDispatch(bc *baseController, w http.ResponseWriter, di usecase.FrenchTarotInteractorIF, param FrenchTarotWebInput, newDefault func(string) *FrenchTarotWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, di.ResetWithConfig(param.ToConfig()))
	case "b", "bid":
		if !requireParam(bc, w, newDefault, param.Bid == nil, "param error: bid is required.") {
			return true
		}
		bc.writePresenterResponse(w, di.Bid(frenchTarotParseBid(*param.Bid)))
	case "pass":
		bc.writePresenterResponse(w, di.Pass())
	case "d", "discard":
		if !requireParam(bc, w, newDefault, param.CardIndices == nil, "param error: cardIndices is required.") {
			return true
		}
		bc.writePresenterResponse(w, di.Discard(param.CardIndices))
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
