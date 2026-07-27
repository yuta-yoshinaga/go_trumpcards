//go:build !js || !wasm || solo

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// CegoWebInput チェゴ (Cego) のWebインプット
type CegoWebInput struct {
	BaseWebInput
	// CardIndex プレイするカードのインデックス
	CardIndex *int `json:"cardIndex,omitempty"`
	// CardIndices Cego 交換で残すカードのインデックス (1 枚)
	CardIndices []int `json:"cardIndices,omitempty"`
	// Bid 入札 ("play")
	Bid *string `json:"bid,omitempty"`
	// Contract コントラクト ("cego" / "handspiel")
	Contract *string `json:"contract,omitempty"`
	// Config ゲーム設定
	Config *CegoWebConfig `json:"config,omitempty"`
}

// CegoWebConfig チェゴのWeb設定
type CegoWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	TargetDeals   *int `json:"targetDeals,omitempty"`
}

// CegoWebOutputPlayer チェゴのWebアウトプットプレイヤー。
type CegoWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
	TrickCount int              `json:"trickCount"`
	CardPoints int              `json:"cardPoints"`
	Score      int              `json:"score"`
	IsDeclarer bool             `json:"isDeclarer"`
}

// CegoWebOutputHint ヒント出力
type CegoWebOutputHint struct {
	Bid         *int   `json:"bid,omitempty"`
	Contract    *int   `json:"contract,omitempty"`
	CardIndices []int  `json:"cardIndices"`
	Reason      string `json:"reason"`
}

// CegoWebOutput チェゴのWebアウトプット。
// 場札 (Cego / blind) の中身は決して出力せず、BlindCount のみを公開する。
type CegoWebOutput struct {
	Players          []*CegoWebOutputPlayer    `json:"players"`
	Phase            int                       `json:"phase"`
	RoundNumber      int                       `json:"roundNumber"`
	TrickNumber      int                       `json:"trickNumber"`
	CurrentPlayerIdx int                       `json:"currentPlayerIdx"`
	LeadPlayerIdx    int                       `json:"leadPlayerIdx"`
	DealerIdx        int                       `json:"dealerIdx"`
	BidPlayerIdx     int                       `json:"bidPlayerIdx"`
	HighestBid       int                       `json:"highestBid"`
	HighestBidder    int                       `json:"highestBidder"`
	DeclarerIdx      int                       `json:"declarerIdx"`
	Contract         int                       `json:"contract"`
	ContractType     int                       `json:"contractType"`
	BlindCount       int                       `json:"blindCount"`
	Blind            []*WebOutputCard          `json:"blind"`
	StashOwner       int                       `json:"stashOwner"`
	CurrentTrick     []*WebOutputTrickCard     `json:"currentTrick"`
	PlayerScores     [domain.CegoPlayerCnt]int `json:"playerScores"`
	LastTrickWinner  int                       `json:"lastTrickWinner"`
	Outcome          int                       `json:"outcome"`
	Result           int                       `json:"result"`
	PlayableIndices  []int                     `json:"playableIndices"`
	GameEndFlag      bool                      `json:"gameEndFlag"`
	WinnerPlayer     int                       `json:"winnerPlayer"`
	IsHumanTurn      bool                      `json:"isHumanTurn"`
	IsHumanBidTurn   bool                      `json:"isHumanBidTurn"`
	IsHumanContract  bool                      `json:"isHumanContract"`
	IsHumanExchange  bool                      `json:"isHumanExchange"`
	Hint             *CegoWebOutputHint        `json:"hint,omitempty"`
	WebOutputBase
	Config CegoWebOutputConfig `json:"config"`
}

// CegoWebOutputConfig チェゴの設定アウトプット
type CegoWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	TargetDeals   int `json:"targetDeals"`
}

// ToConfig builds a CegoConfig from the nested web config, applying bounds checking.
func (c *CegoWebConfig) ToConfig() domain.CegoConfig {
	cfg := domain.DefaultCegoConfig()
	cfg.CpuDifficulty = domain.CegoCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.CegoCpuDifficultyEasy), int(domain.CegoCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.TargetDeals, c.TargetDeals, 1, 1000)
	return cfg
}

// ToConfig builds a CegoConfig from the web input.
func (p CegoWebInput) ToConfig() domain.CegoConfig {
	return configOrDefault(p.Config, (*CegoWebConfig).ToConfig, domain.DefaultCegoConfig())
}

// cegoParseBid 入札文字列を CegoBid に変換する (Pass=不正/パス)。
func cegoParseBid(s string) domain.CegoBid {
	switch s {
	case "play", "p":
		return domain.CegoBidPlay
	default:
		return domain.CegoBidPass
	}
}

// cegoParseContract コントラクト文字列を CegoContract に変換する (None=不正)。
func cegoParseContract(s string) domain.CegoContract {
	switch s {
	case "cego", "c":
		return domain.CegoContractCego
	case "handspiel", "solo", "h":
		return domain.CegoContractHandspiel
	default:
		return domain.CegoContractNone
	}
}

// CegoWebController チェゴのWebコントローラークラス
type CegoWebController = GameWebController[usecase.CegoInteractorIF, CegoWebInput, *CegoWebOutput]

// NewCegoWebController and NewCegoWebControllerWithProvider are the standard and
// provider-backed constructors for CegoWebController.
var NewCegoWebController, NewCegoWebControllerWithProvider = webControllerPair[usecase.CegoInteractorIF, CegoWebInput, *CegoWebOutput](
	newCegoDefaultOutput, cegoDispatch,
)

func newCegoDefaultOutput(msg string) *CegoWebOutput {
	return &CegoWebOutput{
		Players:         make([]*CegoWebOutputPlayer, 0),
		CurrentTrick:    make([]*WebOutputTrickCard, 0),
		Blind:           make([]*WebOutputCard, 0),
		PlayableIndices: make([]int, 0),
		DeclarerIdx:     -1,
		HighestBidder:   -1,
		LastTrickWinner: -1,
		WinnerPlayer:    -1,
		WebOutputBase:   WebOutputBase{Message: msg},
	}
}

func cegoDispatch(bc *baseController, w http.ResponseWriter, di usecase.CegoInteractorIF, param CegoWebInput, newDefault func(string) *CegoWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, di.ResetWithConfig(param.ToConfig()))
	case "b", "bid":
		if !requireParam(bc, w, newDefault, param.Bid == nil, "param error: bid is required.") {
			return true
		}
		bc.writePresenterResponse(w, di.Bid(cegoParseBid(*param.Bid)))
	case "pass":
		bc.writePresenterResponse(w, di.Pass())
	case "cego":
		bc.writePresenterResponse(w, di.ChooseContract(domain.CegoContractCego))
	case "handspiel":
		bc.writePresenterResponse(w, di.ChooseContract(domain.CegoContractHandspiel))
	case "ct", "contract":
		if !requireParam(bc, w, newDefault, param.Contract == nil, "param error: contract is required.") {
			return true
		}
		bc.writePresenterResponse(w, di.ChooseContract(cegoParseContract(*param.Contract)))
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
