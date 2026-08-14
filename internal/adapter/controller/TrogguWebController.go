//go:build !js || !wasm || extra

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// TrogguWebConfig トロッグの Web 設定。
type TrogguWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	TargetDeals   *int `json:"targetDeals,omitempty"`
}

// ToConfig は TrogguWebConfig を domain.TrogguConfig に変換する (境界チェック付き)。
func (c *TrogguWebConfig) ToConfig() domain.TrogguConfig {
	cfg := domain.DefaultTrogguConfig()
	cfg.CpuDifficulty = domain.TrogguCpuDifficulty(webutil.BoundedIntPtr(
		c.CpuDifficulty,
		int(domain.TrogguCpuDifficultyEasy), int(domain.TrogguCpuDifficultyHard),
		int(cfg.CpuDifficulty)))
	cfg.TargetDeals = webutil.BoundedIntPtr(c.TargetDeals,
		domain.TrogguMinDeals, domain.TrogguMaxDeals, cfg.TargetDeals)
	return cfg
}

// TrogguWebInput トロッグの Web インプット。
type TrogguWebInput struct {
	BaseWebInput
	// CardIndex プレイするカードのインデックス
	CardIndex *int `json:"cardIndex,omitempty"`
	// Bid 入札 ("trois" / "solo" / "piccolo" / "misere")
	Bid *string `json:"bid,omitempty"`
	// Config ゲーム設定
	Config *TrogguWebConfig `json:"config,omitempty"`
}

// ToConfig は Web インプットから domain.TrogguConfig を構築する。
func (p TrogguWebInput) ToConfig() domain.TrogguConfig {
	return configOrDefault(p.Config, (*TrogguWebConfig).ToConfig, domain.DefaultTrogguConfig())
}

// TrogguWebOutputPlayer トロッグの Web アウトプットプレイヤー。
type TrogguWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
	TrickCount int              `json:"trickCount"`
	CardPoints int              `json:"cardPoints"`
	Score      int              `json:"score"`
	IsDeclarer bool             `json:"isDeclarer"`
}

// TrogguWebOutputBreakdown ディール精算の内訳。
type TrogguWebOutputBreakdown struct {
	Contract       int    `json:"contract"`
	ContractName   string `json:"contractName"`
	DeclarerPoints int    `json:"declarerPoints"`
	DeclarerTricks int    `json:"declarerTricks"`
	Target         int    `json:"target"`
	TargetIsTricks bool   `json:"targetIsTricks"`
	Won            bool   `json:"won"`
	Base           int    `json:"base"`
	Seats          []int  `json:"seats"`
}

// TrogguWebOutputHint ヒント出力。
type TrogguWebOutputHint struct {
	Bid       *int   `json:"bid,omitempty"`
	CardIndex *int   `json:"cardIndex,omitempty"`
	Reason    string `json:"reason"`
}

// TrogguWebOutputConfig 設定アウトプット。
type TrogguWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	TargetDeals   int `json:"targetDeals"`
}

// TrogguWebOutput トロッグの Web アウトプット。
type TrogguWebOutput struct {
	Players          []*TrogguWebOutputPlayer  `json:"players"`
	Phase            int                       `json:"phase"`
	RoundNumber      int                       `json:"roundNumber"`
	TotalRounds      int                       `json:"totalRounds"`
	TrickNumber      int                       `json:"trickNumber"`
	CurrentPlayerIdx int                       `json:"currentPlayerIdx"`
	DealerIdx        int                       `json:"dealerIdx"`
	BidPlayerIdx     int                       `json:"bidPlayerIdx"`
	HighestBid       int                       `json:"highestBid"`
	DeclarerIdx      int                       `json:"declarerIdx"`
	Contract         int                       `json:"contract"`
	ContractName     string                    `json:"contractName"`
	TalonCount       int                       `json:"talonCount"`
	CurrentTrick     []*WebOutputTrickCard     `json:"currentTrick"`
	LastTrickWinner  int                       `json:"lastTrickWinner"`
	LastTrickCards   []*WebOutputCard          `json:"lastTrickCards"`
	Outcome          int                       `json:"outcome"`
	Breakdown        *TrogguWebOutputBreakdown `json:"breakdown"`
	PlayableIndices  []int                     `json:"playableIndices"`
	GameEndFlag      bool                      `json:"gameEndFlag"`
	WinnerPlayer     int                       `json:"winnerPlayer"`
	IsHumanTurn      bool                      `json:"isHumanTurn"`
	Hint             *TrogguWebOutputHint      `json:"hint,omitempty"`
	WebOutputBase
	Config TrogguWebOutputConfig `json:"config"`
}

// trogguParseBid 入札文字列を TrogguBid に変換する (未知の値はパス)。
func trogguParseBid(s string) domain.TrogguBid {
	switch s {
	case "trois", "t":
		return domain.TrogguBidTrois
	case "solo", "s":
		return domain.TrogguBidSolo
	case "piccolo", "p":
		return domain.TrogguBidPiccolo
	case "misere", "m":
		return domain.TrogguBidMisere
	default:
		return domain.TrogguBidPass
	}
}

// TrogguWebController トロッグの Web コントローラー。
type TrogguWebController = GameWebController[usecase.TrogguInteractorIF, TrogguWebInput, *TrogguWebOutput]

// NewTrogguWebController, NewTrogguWebControllerWithProvider are the standard and
// provider-backed constructors for TrogguWebController.
var NewTrogguWebController, NewTrogguWebControllerWithProvider = webControllerPair[usecase.TrogguInteractorIF, TrogguWebInput, *TrogguWebOutput](
	newTrogguDefaultOutput, trogguDispatch,
)

func newTrogguDefaultOutput(msg string) *TrogguWebOutput {
	return &TrogguWebOutput{
		Players:         make([]*TrogguWebOutputPlayer, 0),
		CurrentTrick:    make([]*WebOutputTrickCard, 0),
		LastTrickCards:  make([]*WebOutputCard, 0),
		PlayableIndices: make([]int, 0),
		DeclarerIdx:     -1,
		LastTrickWinner: -1,
		WinnerPlayer:    -1,
		WebOutputBase:   WebOutputBase{Message: msg},
	}
}

func trogguDispatch(bc *baseController, w http.ResponseWriter, ti usecase.TrogguInteractorIF, param TrogguWebInput, newDefault func(string) *TrogguWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, ti.ResetWithConfig(param.ToConfig()))
	case "b", "bid":
		if !requireParam(bc, w, newDefault, param.Bid == nil, "param error: bid is required.") {
			return true
		}
		// **知らない入札はここで名指しで断る。** ドメインまで運ぶと、送った文字列を
		// 含まない汎用エラーになる。
		bid := trogguParseBid(*param.Bid)
		if !requireParam(bc, w, newDefault, bid == domain.TrogguBidPass,
			"param error: bid must be trois, solo, piccolo or misere.") {
			return true
		}
		bc.writePresenterResponse(w, ti.Bid(bid))
	case "pass":
		bc.writePresenterResponse(w, ti.Pass())
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, ti.Play(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, ti.NextTrick())
	case "nr", "nextround":
		bc.writePresenterResponse(w, ti.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, ti.Hint, ti.ActionLog)
	}
	return true
}
