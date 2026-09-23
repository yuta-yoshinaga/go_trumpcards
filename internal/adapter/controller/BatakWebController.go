package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BatakWebInput Batak Web インプット
type BatakWebInput struct {
	BaseWebInput
	Bid       *int            `json:"bid,omitempty"`
	CardIndex *int            `json:"cardIndex,omitempty"`
	Config    *BatakWebConfig `json:"config,omitempty"`
}

// BatakWebConfig Batak Web 設定
type BatakWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	MaxRounds     *int `json:"maxRounds,omitempty"`
}

// BatakWebOutputPlayer Batak Web アウトプットプレイヤー
//
// RoundScore / CumulativeScore は素の整数スコア。
type BatakWebOutputPlayer struct {
	ID              int              `json:"id"`
	IsHuman         bool             `json:"isHuman"`
	CardCount       int              `json:"cardCount"`
	Cards           []*WebOutputCard `json:"cards"`
	Bid             int              `json:"bid"`
	RoundScore      int              `json:"roundScore"`
	CumulativeScore int              `json:"cumulativeScore"`
	TrickCount      int              `json:"trickCount"`
}

// BatakWebOutputHint ヒント出力
type BatakWebOutputHint struct {
	CardIndex *int   `json:"cardIndex,omitempty"`
	Bid       *int   `json:"bid,omitempty"`
	Reason    string `json:"reason"`
}

// BatakWebOutput Batak Web アウトプット
type BatakWebOutput struct {
	Players          []*BatakWebOutputPlayer `json:"players"`
	Phase            int                     `json:"phase"`
	RoundNumber      int                     `json:"roundNumber"`
	TrickNumber      int                     `json:"trickNumber"`
	CurrentPlayerIdx int                     `json:"currentPlayerIdx"`
	BidPlayerIdx     int                     `json:"bidPlayerIdx"`
	DeclarerIdx      int                     `json:"declarerIdx"` // 親 (デクレアラー) のプレイヤーインデックス (-1 = 未確定)
	HighBid          int                     `json:"highBid"`     // 現在の最高ビッド (0 = 未宣言)
	MinLegalBid      int                     `json:"minLegalBid"` // 人間が宣言可能な最小ビッド (5〜13、13超時は0=パスのみ可能。人間のビッド手番でない間は0)
	CurrentTrick     []*WebOutputTrickCard   `json:"currentTrick"`
	SpadesBroken     bool                    `json:"spadesBroken"`
	GameEndFlag      bool                    `json:"gameEndFlag"`
	WinnerIdx        int                     `json:"winnerIdx"`
	LeadPlayerIdx    int                     `json:"leadPlayerIdx"`
	Hint             *BatakWebOutputHint     `json:"hint,omitempty"`
	ValidPlayIndices []int                   `json:"validPlayIndices"`
	WebOutputBase
	Config BatakWebOutputConfig `json:"config"`
}

// BatakWebOutputConfig Batak 設定アウトプット
type BatakWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	MaxRounds     int `json:"maxRounds"`
}

// ToConfig builds a BatakConfig from the nested web config, applying bounds checking.
func (c *BatakWebConfig) ToConfig() domain.BatakConfig {
	cfg := domain.DefaultBatakConfig()
	cfg.CpuDifficulty = domain.BatakCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.BatakCpuDifficultyEasy), int(domain.BatakCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.MaxRounds, c.MaxRounds, 1, 50)
	return cfg
}

// ToConfig builds a BatakConfig from the web input.
func (p BatakWebInput) ToConfig() domain.BatakConfig {
	return configOrDefault(p.Config, (*BatakWebConfig).ToConfig, domain.DefaultBatakConfig())
}

// BatakWebController Batak Web コントローラークラス
type BatakWebController = GameWebController[usecase.BatakInteractorIF, BatakWebInput, *BatakWebOutput]

// NewBatakWebController and NewBatakWebControllerWithProvider are
// the standard and provider-backed constructors for BatakWebController.
var NewBatakWebController, NewBatakWebControllerWithProvider = webControllerPair[usecase.BatakInteractorIF, BatakWebInput, *BatakWebOutput](
	newBatakDefaultOutput, batakDispatch,
)

func newBatakDefaultOutput(msg string) *BatakWebOutput {
	return &BatakWebOutput{
		Players:          make([]*BatakWebOutputPlayer, 0),
		CurrentTrick:     make([]*WebOutputTrickCard, 0),
		ValidPlayIndices: make([]int, 0),
		WinnerIdx:        -1,
		DeclarerIdx:      -1,
		HighBid:          0,
		MinLegalBid:      0,
		WebOutputBase:    WebOutputBase{Message: msg},
	}
}

func batakDispatch(bc *baseController, w http.ResponseWriter, ci usecase.BatakInteractorIF, param BatakWebInput, newDefault func(string) *BatakWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, ci.ResetWithConfig(param.ToConfig()))
	case "b", "bid":
		if !requireParam(bc, w, newDefault, param.Bid == nil, "param error: bid is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.Bid(*param.Bid))
	case "pass":
		bc.writePresenterResponse(w, ci.Bid(domain.BatakPassBid))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.Play(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, ci.NextTrick())
	case "nr", "nextround":
		bc.writePresenterResponse(w, ci.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, ci.Hint, ci.ActionLog)
	}
	return true
}
