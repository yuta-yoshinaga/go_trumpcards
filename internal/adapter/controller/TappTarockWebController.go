//go:build !js || !wasm || extra

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// TappTarockWebConfig タップ・タロックの Web 設定。
type TappTarockWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	TargetDeals   *int `json:"targetDeals,omitempty"`
}

// ToConfig は TappTarockWebConfig を domain.TappTarockConfig に変換する (境界チェック付き)。
func (c *TappTarockWebConfig) ToConfig() domain.TappTarockConfig {
	cfg := domain.DefaultTappTarockConfig()
	cfg.CpuDifficulty = domain.TappTarockCpuDifficulty(webutil.BoundedIntPtr(
		c.CpuDifficulty,
		int(domain.TappTarockCpuDifficultyEasy), int(domain.TappTarockCpuDifficultyHard),
		int(cfg.CpuDifficulty)))
	cfg.TargetDeals = webutil.BoundedIntPtr(c.TargetDeals,
		domain.TappTarockMinDeals, domain.TappTarockMaxDeals, cfg.TargetDeals)
	return cfg
}

// TappTarockWebInput タップ・タロックの Web インプット。
type TappTarockWebInput struct {
	BaseWebInput
	// CardIndex プレイするカードのインデックス
	CardIndex *int `json:"cardIndex,omitempty"`
	// CardIndices 場札交換で伏せるカードのインデックス (6 枚)
	CardIndices []int `json:"cardIndices,omitempty"`
	// Bid 入札 ("dreier" / "solo")
	Bid *string `json:"bid,omitempty"`
	// Config ゲーム設定
	Config *TappTarockWebConfig `json:"config,omitempty"`
}

// ToConfig は Web インプットから domain.TappTarockConfig を構築する。
func (p TappTarockWebInput) ToConfig() domain.TappTarockConfig {
	return configOrDefault(p.Config, (*TappTarockWebConfig).ToConfig, domain.DefaultTappTarockConfig())
}

// TappTarockWebOutputPlayer タップ・タロックの Web アウトプットプレイヤー。
type TappTarockWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
	TrickCount int              `json:"trickCount"`
	CardPoints int              `json:"cardPoints"`
	Score      int              `json:"score"`
	IsDeclarer bool             `json:"isDeclarer"`
}

// TappTarockWebOutputBreakdown ディール精算の内訳。
type TappTarockWebOutputBreakdown struct {
	Contract   int    `json:"contract"`
	TeamPoints int    `json:"teamPoints"`
	Threshold  int    `json:"threshold"`
	Won        bool   `json:"won"`
	Solo       bool   `json:"solo"`
	Base       int    `json:"base"`
	Seats      []int  `json:"seats"`
	Loser      int    `json:"loser"`
	Name       string `json:"name"`
}

// TappTarockWebOutputHint ヒント出力。
type TappTarockWebOutputHint struct {
	Bid            *int   `json:"bid,omitempty"`
	CardIndex      *int   `json:"cardIndex,omitempty"`
	DiscardIndices []int  `json:"discardIndices"`
	Reason         string `json:"reason"`
}

// TappTarockWebOutputConfig 設定アウトプット。
type TappTarockWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	TargetDeals   int `json:"targetDeals"`
}

// TappTarockWebOutput タップ・タロックの Web アウトプット。
type TappTarockWebOutput struct {
	Players            []*TappTarockWebOutputPlayer  `json:"players"`
	Phase              int                           `json:"phase"`
	RoundNumber        int                           `json:"roundNumber"`
	TotalRounds        int                           `json:"totalRounds"`
	TrickNumber        int                           `json:"trickNumber"`
	CurrentPlayerIdx   int                           `json:"currentPlayerIdx"`
	DealerIdx          int                           `json:"dealerIdx"`
	BidPlayerIdx       int                           `json:"bidPlayerIdx"`
	HighestBid         int                           `json:"highestBid"`
	DeclarerIdx        int                           `json:"declarerIdx"`
	Contract           int                           `json:"contract"`
	ContractName       string                        `json:"contractName"`
	TalonCount         int                           `json:"talonCount"`
	CurrentTrick       []*WebOutputTrickCard         `json:"currentTrick"`
	LastTrickWinner    int                           `json:"lastTrickWinner"`
	LastTrickCards     []*WebOutputCard              `json:"lastTrickCards"`
	Outcome            int                           `json:"outcome"`
	Breakdown          *TappTarockWebOutputBreakdown `json:"breakdown"`
	PlayableIndices    []int                         `json:"playableIndices"`
	DiscardableIndices []int                         `json:"discardableIndices"`
	GameEndFlag        bool                          `json:"gameEndFlag"`
	WinnerPlayer       int                           `json:"winnerPlayer"`
	IsHumanTurn        bool                          `json:"isHumanTurn"`
	Hint               *TappTarockWebOutputHint      `json:"hint,omitempty"`
	WebOutputBase
	Config TappTarockWebOutputConfig `json:"config"`
}

// tapptarockParseBid 入札文字列を TappTarockBid に変換する。
//
// **Trischaken は受け付けない。** 誰も落札しなかった結果としてしか成立しないので、
// 宣言できると「全員パス」との区別が付かなくなる。
func tapptarockParseBid(s string) domain.TappTarockBid {
	switch s {
	case "dreier", "d":
		return domain.TappTarockBidDreier
	case "solo", "s":
		return domain.TappTarockBidSolo
	default:
		return domain.TappTarockBidPass
	}
}

// TappTarockWebController タップ・タロックの Web コントローラー。
type TappTarockWebController = GameWebController[usecase.TappTarockInteractorIF, TappTarockWebInput, *TappTarockWebOutput]

// NewTappTarockWebController, NewTappTarockWebControllerWithProvider are the standard and
// provider-backed constructors for TappTarockWebController.
var NewTappTarockWebController, NewTappTarockWebControllerWithProvider = webControllerPair[usecase.TappTarockInteractorIF, TappTarockWebInput, *TappTarockWebOutput](
	newTappTarockDefaultOutput, tapptarockDispatch,
)

func newTappTarockDefaultOutput(msg string) *TappTarockWebOutput {
	return &TappTarockWebOutput{
		Players:            make([]*TappTarockWebOutputPlayer, 0),
		CurrentTrick:       make([]*WebOutputTrickCard, 0),
		LastTrickCards:     make([]*WebOutputCard, 0),
		PlayableIndices:    make([]int, 0),
		DiscardableIndices: make([]int, 0),
		DeclarerIdx:        -1,
		LastTrickWinner:    -1,
		WinnerPlayer:       -1,
		WebOutputBase:      WebOutputBase{Message: msg},
	}
}

func tapptarockDispatch(bc *baseController, w http.ResponseWriter, zi usecase.TappTarockInteractorIF, param TappTarockWebInput, newDefault func(string) *TappTarockWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, zi.ResetWithConfig(param.ToConfig()))
	case "b", "bid":
		if !requireParam(bc, w, newDefault, param.Bid == nil, "param error: bid is required.") {
			return true
		}
		// **知らない入札はここで名指しで断る。** ドメインまで運ぶと
		// 「bid 0 cannot be declared」という、送った文字列を含まない返事になる。
		bid := tapptarockParseBid(*param.Bid)
		if !requireParam(bc, w, newDefault, bid == domain.TappTarockBidPass,
			"param error: bid must be dreier or solo.") {
			return true
		}
		bc.writePresenterResponse(w, zi.Bid(bid))
	case "pass":
		bc.writePresenterResponse(w, zi.Pass())
	case "d", "discard":
		if !requireParam(bc, w, newDefault, param.CardIndices == nil, "param error: cardIndices is required.") {
			return true
		}
		bc.writePresenterResponse(w, zi.Discard(param.CardIndices))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, zi.Play(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, zi.NextTrick())
	case "nr", "nextround":
		bc.writePresenterResponse(w, zi.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, zi.Hint, zi.ActionLog)
	}
	return true
}
